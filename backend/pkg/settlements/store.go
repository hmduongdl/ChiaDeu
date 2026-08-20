// Package settlements quản lý kỳ quyết toán và vòng đời giao dịch hoàn tiền.
// Chốt kỳ chạy trong một transaction: khóa khoản chi chưa chốt, tính số dư ròng,
// chạy resolver và tạo settlements — mọi người dùng chỉ có thể trả/nhận theo quan
// hệ from/to được sinh ra.
package settlements

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"github.com/hmduongdl/ChiaDeu/pkg/audit"
	"github.com/hmduongdl/ChiaDeu/models"
	"github.com/hmduongdl/ChiaDeu/services"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotGroupMember         = errors.New("bạn không phải là thành viên hoạt động của nhóm")
	ErrNotAdmin               = errors.New("chỉ quản trị viên mới được thực hiện thao tác này")
	ErrBatchAlreadyOpen       = errors.New("nhóm đã có một kỳ quyết toán đang mở")
	ErrNoOpenExpenses         = errors.New("không có khoản chi nào để chốt")
	ErrInvalidExpenseData     = errors.New("dữ liệu khoản chi không hợp lệ khi chốt kỳ")
	ErrBatchNotFound          = errors.New("kỳ quyết toán không tồn tại")
	ErrSettlementNotFound     = errors.New("giao dịch hoàn tiền không tồn tại")
	ErrBatchClosed            = errors.New("kỳ quyết toán không còn mở")
	ErrBatchHasPaidSettlement = errors.New("kỳ có giao dịch đã thanh toán, không thể hủy")
	ErrNotPayer               = errors.New("chỉ người phải trả mới được báo đã chuyển")
	ErrNotRecipient           = errors.New("chỉ người được nhận mới được xác nhận")
	ErrInvalidTransition      = errors.New("trạng thái giao dịch không cho phép thao tác này")
)

// BatchSnapshot là ảnh chụp một kỳ quyết toán cùng các settlement và khoản chi
// thuộc kỳ.
type BatchSnapshot struct {
	Batch       models.SettlementBatch
	Settlements []models.Settlement
	Expenses    []models.Expense
}

// SettlementContext nối settlement với nhóm và trạng thái kỳ chứa nó.
type SettlementContext struct {
	Settlement  models.Settlement
	GroupID     string
	BatchStatus string
}

// Store giới hạn các thao tác quyết toán, giúp handler kiểm thử bằng fake.
type Store interface {
	CloseBatch(ctx context.Context, groupID, actorID, idempotencyKey string) (BatchSnapshot, error)
	GetBatch(ctx context.Context, groupID, batchID, actorID string) (BatchSnapshot, error)
	GetSettlement(ctx context.Context, settlementID string) (SettlementContext, error)
	MarkSent(ctx context.Context, settlementID, actorID string) (models.Settlement, error)
	Confirm(ctx context.Context, settlementID, actorID string) (models.Settlement, error)
	Reject(ctx context.Context, settlementID, actorID string) (models.Settlement, error)
	CancelBatch(ctx context.Context, batchID, actorID string) (BatchSnapshot, error)
}

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

const (
	expenseColumns = `id, group_id, created_by, paid_by, description, amount_minor,
	split_type, expense_date, batch_id, status, created_at, updated_at`

	settlementColumns = `id, batch_id, from_user_id, to_user_id, amount_minor, payment_code,
	status, marked_sent_at, paid_at, created_at, updated_at`
)

// ----------------------------------------------------------------------------
// Chốt kỳ
// ----------------------------------------------------------------------------

// CloseBatch chốt toàn bộ khoản chi chưa chốt của nhóm thành một kỳ và tạo các
// settlement theo resolver của Chế độ chia đều linh hoạt. Chạy trong một
// transaction; idempotency_key giúp retry an toàn.
func (s *PostgresStore) CloseBatch(ctx context.Context, groupID, actorID, idempotencyKey string) (BatchSnapshot, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return BatchSnapshot{}, err
	}
	defer tx.Rollback(ctx)

	if ok, err := isAdminByQuery(ctx, tx, groupID, actorID); err != nil {
		return BatchSnapshot{}, err
	} else if !ok {
		return BatchSnapshot{}, ErrNotAdmin
	}

	// Idempotency: retry với cùng key trả về đúng kết quả đã chốt trước đó.
	if existing, found, err := s.batchByKey(ctx, tx, groupID, idempotencyKey); err != nil {
		return BatchSnapshot{}, err
	} else if found {
		return loadSnapshot(ctx, tx, groupID, existing.ID)
	}

	var openID string
	err = tx.QueryRow(ctx, `
		SELECT id FROM settlement_batches
		WHERE group_id = $1 AND status = $2`, groupID, models.BatchStatusOpen,
	).Scan(&openID)
	if err == nil {
		return BatchSnapshot{}, ErrBatchAlreadyOpen
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return BatchSnapshot{}, err
	}

	// Khóa các khoản chi đang treo để request đồng thời không chốt trùng.
	rows, err := tx.Query(ctx, `
		SELECT `+expenseColumns+`
		FROM expenses
		WHERE group_id = $1 AND status = $2 AND batch_id IS NULL
		ORDER BY expense_date, created_at
		FOR UPDATE`, groupID, models.ExpenseStatusActive)
	if err != nil {
		return BatchSnapshot{}, err
	}
	expenses, err := scanExpenses(rows)
	if err != nil {
		return BatchSnapshot{}, err
	}
	if len(expenses) == 0 {
		return BatchSnapshot{}, ErrNoOpenExpenses
	}

	expenseIDs := make([]string, len(expenses))
	for index := range expenses {
		expenseIDs[index] = expenses[index].ID
	}
	splitRows, err := tx.Query(ctx, `
		SELECT id, expense_id, user_id, share_minor, created_at
		FROM expense_splits
		WHERE expense_id = ANY($1)`, expenseIDs)
	if err != nil {
		return BatchSnapshot{}, err
	}
	splits, err := scanSplits(splitRows)
	if err != nil {
		return BatchSnapshot{}, err
	}

	balances, err := services.CalculateNetBalances(expenses, splits)
	if err != nil {
		return BatchSnapshot{}, fmt.Errorf("%w: %v", ErrInvalidExpenseData, err)
	}
	resolved, err := services.SimplifyDebts(balances)
	if err != nil {
		return BatchSnapshot{}, fmt.Errorf("%w: %v", ErrInvalidExpenseData, err)
	}

	var batch models.SettlementBatch
	err = tx.QueryRow(ctx, `
		INSERT INTO settlement_batches (group_id, created_by, idempotency_key)
		VALUES ($1, $2, $3)
		RETURNING id, status, created_at`,
		groupID, actorID, idempotencyKey,
	).Scan(&batch.ID, &batch.Status, &batch.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Đồng thời chốt: trả kết quả đã có theo key nếu request kia khớp.
			existing, found, existingErr := s.batchByKey(ctx, tx, groupID, idempotencyKey)
			if existingErr != nil {
				return BatchSnapshot{}, existingErr
			}
			if found {
				return loadSnapshot(ctx, tx, groupID, existing.ID)
			}
			return BatchSnapshot{}, ErrBatchAlreadyOpen
		}
		return BatchSnapshot{}, err
	}
	batch.GroupID = groupID

	saved := make([]models.Settlement, 0, len(resolved))
	for _, settlement := range resolved {
		identity, err := s.insertSettlement(ctx, tx, batch.ID, settlement)
		if err != nil {
			return BatchSnapshot{}, err
		}
		settlement.ID = identity.id
		settlement.PaymentCode = identity.paymentCode
		saved = append(saved, settlement)
	}

	if _, err := tx.Exec(ctx, `UPDATE expenses SET batch_id = $1 WHERE id = ANY($2)`, batch.ID, expenseIDs); err != nil {
		return BatchSnapshot{}, err
	}

	if err := audit.Append(ctx, tx, audit.Entry{
		GroupID:    groupID,
		ActorID:    actorID,
		Action:     "batch.created",
		EntityType: "settlement_batch",
		EntityID:   batch.ID,
		Metadata:   map[string]any{"expenseCount": len(expenses), "settlementCount": len(saved)},
	}); err != nil {
		return BatchSnapshot{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return BatchSnapshot{}, err
	}
	return BatchSnapshot{Batch: batch, Settlements: saved, Expenses: expenses}, nil
}

// ----------------------------------------------------------------------------
// Đọc ảnh chụp
// ----------------------------------------------------------------------------

// GetBatch trả về ảnh chụp một kỳ; chỉ thành viên hoạt động của nhóm mới xem được.
func (s *PostgresStore) GetBatch(ctx context.Context, groupID, batchID, actorID string) (BatchSnapshot, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return BatchSnapshot{}, err
	}
	defer tx.Rollback(ctx)

	if ok, err := isActiveMemberByQuery(ctx, tx, groupID, actorID); err != nil {
		return BatchSnapshot{}, err
	} else if !ok {
		return BatchSnapshot{}, ErrNotGroupMember
	}
	return loadSnapshot(ctx, tx, groupID, batchID)
}

// GetSettlement trả về settlement kèm ngữ cảnh nhóm và trạng thái kỳ.
func (s *PostgresStore) GetSettlement(ctx context.Context, settlementID string) (SettlementContext, error) {
	var context SettlementContext
	var settlement models.Settlement
	var markedSentAt, paidAt, createdAt, updatedAt pgtype.Timestamptz
	err := s.pool.QueryRow(ctx, `
		SELECT st.`+settlementColumns+`, b.group_id, b.status
		FROM settlements st
		JOIN settlement_batches b ON b.id = st.batch_id
		WHERE st.id = $1`, settlementID,
	).Scan(
		&settlement.ID, &settlement.BatchID, &settlement.FromUserID, &settlement.ToUserID,
		&settlement.AmountMinor, &settlement.PaymentCode, &settlement.Status,
		&markedSentAt, &paidAt, &createdAt, &updatedAt, &context.GroupID, &context.BatchStatus,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return SettlementContext{}, ErrSettlementNotFound
	}
	if err != nil {
		return SettlementContext{}, err
	}
	context.Settlement = applySettlementTimes(settlement, markedSentAt, paidAt, createdAt, updatedAt)
	return context, nil
}

// ----------------------------------------------------------------------------
// Vòng đời settlement
// ----------------------------------------------------------------------------

// MarkSent cho phép người trả báo đã chuyển tiền: PENDING → AWAITING_CONFIRMATION.
func (s *PostgresStore) MarkSent(ctx context.Context, settlementID, actorID string) (models.Settlement, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Settlement{}, err
	}
	defer tx.Rollback(ctx)

	existing, err := s.lockSettlement(ctx, tx, settlementID)
	if err != nil {
		return models.Settlement{}, err
	}
	groupID, batchStatus, err := s.batchContextByQuery(ctx, tx, existing.BatchID)
	if err != nil {
		return models.Settlement{}, err
	}
	if err := s.requireActiveMember(ctx, tx, groupID, actorID); err != nil {
		return models.Settlement{}, err
	}
	if existing.FromUserID != actorID {
		return models.Settlement{}, ErrNotPayer
	}
	if batchStatus != models.BatchStatusOpen {
		return models.Settlement{}, ErrBatchClosed
	}
	if existing.Status != models.SettlementStatusPending {
		return models.Settlement{}, ErrInvalidTransition
	}

	updated, err := s.updateSettlementStatus(ctx, tx, settlementID, models.SettlementStatusAwaitingConfirmation, "marked_sent_at")
	if err != nil {
		return models.Settlement{}, err
	}
	if err := s.appendEvent(ctx, tx, groupID, actorID, settlementID, "settlement.marked_sent", "MARKED_SENT"); err != nil {
		return models.Settlement{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return models.Settlement{}, err
	}
	return updated, nil
}

// Confirm cho phép người nhận xác nhận đã nhận; hoàn tất kỳ nếu mọi settlement
// trong kỳ đều đã thanh toán.
func (s *PostgresStore) Confirm(ctx context.Context, settlementID, actorID string) (models.Settlement, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Settlement{}, err
	}
	defer tx.Rollback(ctx)

	existing, err := s.lockSettlement(ctx, tx, settlementID)
	if err != nil {
		return models.Settlement{}, err
	}
	groupID, batchStatus, err := s.batchContextByQuery(ctx, tx, existing.BatchID)
	if err != nil {
		return models.Settlement{}, err
	}
	if err := s.requireActiveMember(ctx, tx, groupID, actorID); err != nil {
		return models.Settlement{}, err
	}
	if existing.ToUserID != actorID {
		return models.Settlement{}, ErrNotRecipient
	}
	if batchStatus != models.BatchStatusOpen {
		return models.Settlement{}, ErrBatchClosed
	}
	if existing.Status != models.SettlementStatusAwaitingConfirmation {
		return models.Settlement{}, ErrInvalidTransition
	}

	updated, err := s.updateSettlementStatus(ctx, tx, settlementID, models.SettlementStatusPaid, "paid_at")
	if err != nil {
		return models.Settlement{}, err
	}
	if err := s.appendEvent(ctx, tx, groupID, actorID, settlementID, "settlement.confirmed", "CONFIRMED"); err != nil {
		return models.Settlement{}, err
	}
	if err := s.completeBatchIfAllPaid(ctx, tx, existing.BatchID); err != nil {
		return models.Settlement{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return models.Settlement{}, err
	}
	return updated, nil
}

// Reject cho phép người nhận từ chối xác nhận, đưa giao dịch về PENDING.
func (s *PostgresStore) Reject(ctx context.Context, settlementID, actorID string) (models.Settlement, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Settlement{}, err
	}
	defer tx.Rollback(ctx)

	existing, err := s.lockSettlement(ctx, tx, settlementID)
	if err != nil {
		return models.Settlement{}, err
	}
	groupID, batchStatus, err := s.batchContextByQuery(ctx, tx, existing.BatchID)
	if err != nil {
		return models.Settlement{}, err
	}
	if err := s.requireActiveMember(ctx, tx, groupID, actorID); err != nil {
		return models.Settlement{}, err
	}
	if existing.ToUserID != actorID {
		return models.Settlement{}, ErrNotRecipient
	}
	if batchStatus != models.BatchStatusOpen {
		return models.Settlement{}, ErrBatchClosed
	}
	if existing.Status != models.SettlementStatusAwaitingConfirmation {
		return models.Settlement{}, ErrInvalidTransition
	}

	updated, err := s.updateSettlementStatus(ctx, tx, settlementID, models.SettlementStatusPending, "clear_marked_sent_at")
	if err != nil {
		return models.Settlement{}, err
	}
	if err := s.appendEvent(ctx, tx, groupID, actorID, settlementID, "settlement.rejected", "REJECTED"); err != nil {
		return models.Settlement{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return models.Settlement{}, err
	}
	return updated, nil
}

// CancelBatch hủy kỳ khi chưa có settlement PAID; trả các khoản chi về chưa chốt.
func (s *PostgresStore) CancelBatch(ctx context.Context, batchID, actorID string) (BatchSnapshot, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return BatchSnapshot{}, err
	}
	defer tx.Rollback(ctx)

	var groupID, status string
	err = tx.QueryRow(ctx, `SELECT group_id, status FROM settlement_batches WHERE id = $1 FOR UPDATE`, batchID).Scan(&groupID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return BatchSnapshot{}, ErrBatchNotFound
	}
	if err != nil {
		return BatchSnapshot{}, err
	}
	if status != models.BatchStatusOpen && status != models.BatchStatusCancelled {
		return BatchSnapshot{}, ErrBatchClosed
	}
	if ok, err := isAdminByQuery(ctx, tx, groupID, actorID); err != nil {
		return BatchSnapshot{}, err
	} else if !ok {
		return BatchSnapshot{}, ErrNotAdmin
	}

	var paidCount int
	if err := tx.QueryRow(ctx, `
		SELECT count(*) FROM settlements WHERE batch_id = $1 AND status = $2`,
		batchID, models.SettlementStatusPaid,
	).Scan(&paidCount); err != nil {
		return BatchSnapshot{}, err
	}
	if paidCount > 0 {
		return BatchSnapshot{}, ErrBatchHasPaidSettlement
	}

	if _, err := tx.Exec(ctx, `
		UPDATE settlements
		SET status = $2, updated_at = now()
		WHERE batch_id = $1 AND status IN ($3, $4)`,
		batchID, models.SettlementStatusCancelled,
		models.SettlementStatusPending, models.SettlementStatusAwaitingConfirmation,
	); err != nil {
		return BatchSnapshot{}, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE settlement_batches SET status = $2, cancelled_at = now()
		WHERE id = $1`, batchID, models.BatchStatusCancelled); err != nil {
		return BatchSnapshot{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE expenses SET batch_id = NULL WHERE batch_id = $1`, batchID); err != nil {
		return BatchSnapshot{}, err
	}
	if err := audit.Append(ctx, tx, audit.Entry{
		GroupID:    groupID,
		ActorID:    actorID,
		Action:     "batch.cancelled",
		EntityType: "settlement_batch",
		EntityID:   batchID,
	}); err != nil {
		return BatchSnapshot{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return BatchSnapshot{}, err
	}
	return s.GetBatch(ctx, groupID, batchID, actorID)
}

// ----------------------------------------------------------------------------
// Helpers nội bộ
// ----------------------------------------------------------------------------

func (s *PostgresStore) lockSettlement(ctx context.Context, tx pgx.Tx, settlementID string) (models.Settlement, error) {
	settlement, err := scanSettlement(tx.QueryRow(ctx, `
		SELECT `+settlementColumns+`
		FROM settlements
		WHERE id = $1
		FOR UPDATE`, settlementID))
	if errors.Is(err, ErrSettlementNotFound) {
		return models.Settlement{}, ErrSettlementNotFound
	}
	return settlement, err
}

func (s *PostgresStore) batchContextByQuery(ctx context.Context, tx pgx.Tx, batchID string) (groupID, status string, err error) {
	err = tx.QueryRow(ctx, `SELECT group_id, status FROM settlement_batches WHERE id = $1`, batchID).Scan(&groupID, &status)
	return groupID, status, err
}

func (s *PostgresStore) requireActiveMember(ctx context.Context, tx pgx.Tx, groupID, userID string) error {
	ok, err := isActiveMemberByQuery(ctx, tx, groupID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotGroupMember
	}
	return nil
}

func (s *PostgresStore) appendEvent(ctx context.Context, tx pgx.Tx, groupID, actorID, settlementID, action, eventType string) error {
	if err := audit.Append(ctx, tx, audit.Entry{
		GroupID:    groupID,
		ActorID:    actorID,
		Action:     action,
		EntityType: "settlement",
		EntityID:   settlementID,
	}); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO settlement_events (settlement_id, actor_id, event_type)
		VALUES ($1, $2, $3)`, settlementID, actorID, eventType)
	return err
}

// updateSettlementStatus cập nhật trạng thái settlement và trả dòng đầy đủ.
func (s *PostgresStore) updateSettlementStatus(ctx context.Context, tx pgx.Tx, settlementID, status, timeColumn string) (models.Settlement, error) {
	var setClause string
	switch timeColumn {
	case "paid_at":
		setClause = `status = $1, paid_at = now(), updated_at = now()`
	case "marked_sent_at":
		setClause = `status = $1, marked_sent_at = now(), updated_at = now()`
	case "clear_marked_sent_at":
		setClause = `status = $1, marked_sent_at = NULL, updated_at = now()`
	default:
		setClause = `status = $1, updated_at = now()`
	}
	query := fmt.Sprintf(`UPDATE settlements SET %s WHERE id = $2 RETURNING %s`, setClause, settlementColumns)
	return scanSettlement(tx.QueryRow(ctx, query, status, settlementID))
}

func (s *PostgresStore) completeBatchIfAllPaid(ctx context.Context, tx pgx.Tx, batchID string) error {
	var remaining int
	if err := tx.QueryRow(ctx, `
		SELECT count(*) FROM settlements WHERE batch_id = $1 AND status <> $2`,
		batchID, models.SettlementStatusPaid,
	).Scan(&remaining); err != nil {
		return err
	}
	if remaining == 0 {
		_, err := tx.Exec(ctx, `
			UPDATE settlement_batches SET status = $2, completed_at = now()
			WHERE id = $1`, batchID, models.BatchStatusCompleted)
		return err
	}
	return nil
}

func (s *PostgresStore) insertSettlement(ctx context.Context, tx pgx.Tx, batchID string, settlement models.Settlement) (settlementIdentity, error) {
	for attempt := 0; attempt < 5; attempt++ {
		paymentCode, err := randomPaymentCode()
		if err != nil {
			return settlementIdentity{}, err
		}
		var identity settlementIdentity
		err = tx.QueryRow(ctx, `
			INSERT INTO settlements (batch_id, from_user_id, to_user_id, amount_minor, payment_code, status)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, payment_code`,
			batchID, settlement.FromUserID, settlement.ToUserID, settlement.AmountMinor, paymentCode, models.SettlementStatusPending,
		).Scan(&identity.id, &identity.paymentCode)
		if err == nil {
			return identity, nil
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "uq_settlements_payment_code_lower" {
			continue // payment_code trùng; thử mã khác
		}
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return settlementIdentity{}, fmt.Errorf("%w: %s", ErrInvalidExpenseData, pgErr.ConstraintName)
		}
		return settlementIdentity{}, err
	}
	return settlementIdentity{}, errors.New("không tạo được mã payment code sau nhiều lần thử")
}

type settlementIdentity struct {
	id          string
	paymentCode string
}

func (s *PostgresStore) batchByKey(ctx context.Context, tx pgx.Tx, groupID, idempotencyKey string) (models.SettlementBatch, bool, error) {
	var batch models.SettlementBatch
	var completedAt, cancelledAt pgtype.Timestamptz
	err := tx.QueryRow(ctx, `
		SELECT id, group_id, created_by, idempotency_key, status, created_at, completed_at, cancelled_at
		FROM settlement_batches
		WHERE group_id = $1 AND idempotency_key = $2`, groupID, idempotencyKey,
	).Scan(&batch.ID, &batch.GroupID, &batch.CreatedBy, &batch.IdempotencyKey, &batch.Status, &batch.CreatedAt, &completedAt, &cancelledAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.SettlementBatch{}, false, nil
	}
	if err != nil {
		return models.SettlementBatch{}, false, err
	}
	if completedAt.Valid {
		batch.CompletedAt = &completedAt.Time
	}
	if cancelledAt.Valid {
		batch.CancelledAt = &cancelledAt.Time
	}
	return batch, true, nil
}

func isActiveMemberByQuery(ctx context.Context, tx pgx.Tx, groupID, userID string) (bool, error) {
	var status string
	err := tx.QueryRow(ctx, `
		SELECT status FROM group_members WHERE group_id = $1 AND user_id = $2`, groupID, userID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return status == models.MemberStatusActive, nil
}

func isAdminByQuery(ctx context.Context, tx pgx.Tx, groupID, userID string) (bool, error) {
	var role, status string
	err := tx.QueryRow(ctx, `
		SELECT role, status FROM group_members WHERE group_id = $1 AND user_id = $2`, groupID, userID).Scan(&role, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return status == models.MemberStatusActive && role == models.RoleAdmin, nil
}

// ----------------------------------------------------------------------------
// Scan helpers
// ----------------------------------------------------------------------------

// loadSnapshot đọc batch + settlements + expenses trong cùng querier (tx hoặc pool).
func loadSnapshot(ctx context.Context, querier snapshotQuerier, groupID, batchID string) (BatchSnapshot, error) {
	var batch models.SettlementBatch
	var completedAt, cancelledAt pgtype.Timestamptz
	err := querier.QueryRow(ctx, `
		SELECT id, group_id, created_by, idempotency_key, status, created_at, completed_at, cancelled_at
		FROM settlement_batches
		WHERE id = $1 AND group_id = $2`, batchID, groupID,
	).Scan(&batch.ID, &batch.GroupID, &batch.CreatedBy, &batch.IdempotencyKey, &batch.Status, &batch.CreatedAt, &completedAt, &cancelledAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return BatchSnapshot{}, ErrBatchNotFound
	}
	if err != nil {
		return BatchSnapshot{}, err
	}
	if completedAt.Valid {
		batch.CompletedAt = &completedAt.Time
	}
	if cancelledAt.Valid {
		batch.CancelledAt = &cancelledAt.Time
	}

	rows, err := querier.Query(ctx, `
		SELECT `+settlementColumns+`
		FROM settlements
		WHERE batch_id = $1
		ORDER BY created_at, id`, batchID)
	if err != nil {
		return BatchSnapshot{}, err
	}
	settlements, err := scanSettlements(rows)
	if err != nil {
		return BatchSnapshot{}, err
	}

	expenseRows, err := querier.Query(ctx, `
		SELECT `+expenseColumns+`
		FROM expenses
		WHERE batch_id = $1
		ORDER BY expense_date, created_at`, batchID)
	if err != nil {
		return BatchSnapshot{}, err
	}
	expenses, err := scanExpenses(expenseRows)
	if err != nil {
		return BatchSnapshot{}, err
	}
	return BatchSnapshot{Batch: batch, Settlements: settlements, Expenses: expenses}, nil
}

// snapshotQuerier trừu tượng hóa pgxpool.Pool và pgx.Tx cho loadSnapshot.
type snapshotQuerier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func scanExpenses(rows pgx.Rows) ([]models.Expense, error) {
	defer rows.Close()
	var expenses []models.Expense
	for rows.Next() {
		var expense models.Expense
		var batchID pgtype.UUID
		var updatedAt pgtype.Timestamptz
		if err := rows.Scan(&expense.ID, &expense.GroupID, &expense.CreatedBy, &expense.PaidBy, &expense.Description,
			&expense.AmountMinor, &expense.SplitType, &expense.ExpenseDate, &batchID, &expense.Status,
			&expense.CreatedAt, &updatedAt); err != nil {
			return nil, err
		}
		if batchID.Valid {
			value := batchID.String()
			expense.BatchID = &value
		}
		if updatedAt.Valid {
			expense.UpdatedAt = &updatedAt.Time
		}
		expenses = append(expenses, expense)
	}
	return expenses, rows.Err()
}

func scanSplits(rows pgx.Rows) ([]models.ExpenseSplit, error) {
	defer rows.Close()
	var splits []models.ExpenseSplit
	for rows.Next() {
		var split models.ExpenseSplit
		if err := rows.Scan(&split.ID, &split.ExpenseID, &split.UserID, &split.ShareMinor, &split.CreatedAt); err != nil {
			return nil, err
		}
		splits = append(splits, split)
	}
	return splits, rows.Err()
}

func scanSettlements(rows pgx.Rows) ([]models.Settlement, error) {
	defer rows.Close()
	var settlements []models.Settlement
	for rows.Next() {
		settlement, err := scanSettlement(rows)
		if err != nil {
			return nil, err
		}
		settlements = append(settlements, settlement)
	}
	return settlements, rows.Err()
}

func scanSettlement(row pgx.Row) (models.Settlement, error) {
	settlement := models.Settlement{}
	var markedSentAt, paidAt, createdAt, updatedAt pgtype.Timestamptz
	err := row.Scan(&settlement.ID, &settlement.BatchID, &settlement.FromUserID, &settlement.ToUserID,
		&settlement.AmountMinor, &settlement.PaymentCode, &settlement.Status,
		&markedSentAt, &paidAt, &createdAt, &updatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Settlement{}, ErrSettlementNotFound
	}
	if err != nil {
		return models.Settlement{}, err
	}
	return applySettlementTimes(settlement, markedSentAt, paidAt, createdAt, updatedAt), nil
}

func applySettlementTimes(settlement models.Settlement, markedSentAt, paidAt, createdAt, updatedAt pgtype.Timestamptz) models.Settlement {
	if markedSentAt.Valid {
		settlement.MarkedSentAt = &markedSentAt.Time
	}
	if paidAt.Valid {
		settlement.PaidAt = &paidAt.Time
	}
	settlement.CreatedAt = createdAt.Time
	if updatedAt.Valid {
		settlement.UpdatedAt = &updatedAt.Time
	}
	return settlement
}

const paymentCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

const paymentCodeLength = 6

func randomPaymentCode() (string, error) {
	max := big.NewInt(int64(len(paymentCodeAlphabet)))
	buffer := make([]byte, paymentCodeLength)
	for index := range buffer {
		value, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		buffer[index] = paymentCodeAlphabet[value.Int64()]
	}
	return string(buffer), nil
}
