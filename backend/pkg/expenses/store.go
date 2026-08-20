// Package expenses quản lý khoản chi và phần chia: tạo, sửa và đọc dữ liệu
// chi phí của một nhóm. Mọi thao tác ghi đều qua transaction.
package expenses

import (
	"context"
	"errors"

	"github.com/hmduongdl/ChiaDeu/pkg/audit"
	"github.com/hmduongdl/ChiaDeu/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrExpenseNotFound = errors.New("khoản chi không tồn tại trong nhóm này")
	ErrNotActiveMember = errors.New("người ứng hoặc người chia không phải thành viên hoạt động")
	ErrExpenseLocked   = errors.New("khoản chi đã thuộc kỳ quyết toán, không thể sửa")
	ErrExpenseVoided   = errors.New("khoản chi đã bị hủy")
	ErrNotOwner        = errors.New("chỉ người tạo khoản chi mới được sửa")
)

// Store giới hạn các thao tác khoản chi, giúp handler/service kiểm thử bằng fake.
type Store interface {
	CreateExpenseWithSplits(ctx context.Context, expense models.Expense, splits []models.ExpenseSplit) (models.Expense, error)
	UpdateExpenseWithSplits(ctx context.Context, expense models.Expense, splits []models.ExpenseSplit) (models.Expense, error)
	GetExpenseWithSplits(ctx context.Context, groupID, expenseID string) (models.Expense, []models.ExpenseSplit, error)
	ListUnsettledExpensesWithSplits(ctx context.Context, groupID string) ([]models.Expense, []models.ExpenseSplit, error)
}

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

// CreateExpenseWithSplits ghi khoản chi và các phần chia trong một transaction,
// sau khi xác nhận mọi người liên quan đều là thành viên hoạt động.
func (s *PostgresStore) CreateExpenseWithSplits(ctx context.Context, expense models.Expense, splits []models.ExpenseSplit) (models.Expense, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Expense{}, err
	}
	defer tx.Rollback(ctx)

	if err := s.requireActiveMembers(ctx, tx, expense.GroupID, involvedUsers(expense, splits)); err != nil {
		return models.Expense{}, err
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO expenses (group_id, created_by, paid_by, description, amount_minor, split_type, expense_date, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`,
		expense.GroupID, expense.CreatedBy, expense.PaidBy, expense.Description,
		expense.AmountMinor, expense.SplitType, expense.ExpenseDate, models.ExpenseStatusActive,
	).Scan(&expense.ID, &expense.CreatedAt)
	if err != nil {
		return models.Expense{}, err
	}

	for index := range splits {
		splits[index].ExpenseID = expense.ID
		if _, err := tx.Exec(ctx, `
			INSERT INTO expense_splits (expense_id, user_id, share_minor)
			VALUES ($1, $2, $3)`,
			expense.ID, splits[index].UserID, splits[index].ShareMinor,
		); err != nil {
			return models.Expense{}, err
		}
	}

	if err := audit.Append(ctx, tx, audit.Entry{
		GroupID:    expense.GroupID,
		ActorID:    expense.CreatedBy,
		Action:     "expense.created",
		EntityType: "expense",
		EntityID:   expense.ID,
		Metadata:   map[string]any{"amountMinor": expense.AmountMinor},
	}); err != nil {
		return models.Expense{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Expense{}, err
	}
	return expense, nil
}

// UpdateExpenseWithSplits thay thế thông tin và phần chia của một khoản chi chưa
// chốt, chỉ cho phép người tạo bản ghi.
func (s *PostgresStore) UpdateExpenseWithSplits(ctx context.Context, expense models.Expense, splits []models.ExpenseSplit) (models.Expense, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Expense{}, err
	}
	defer tx.Rollback(ctx)

	var batchID pgtype.UUID
	var currentStatus, createdBy string
	err = tx.QueryRow(ctx, `
		SELECT batch_id, status, created_by
		FROM expenses
		WHERE id = $1 AND group_id = $2
		FOR UPDATE`,
		expense.ID, expense.GroupID,
	).Scan(&batchID, &currentStatus, &createdBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Expense{}, ErrExpenseNotFound
	}
	if err != nil {
		return models.Expense{}, err
	}
	if currentStatus != models.ExpenseStatusActive {
		return models.Expense{}, ErrExpenseVoided
	}
	if batchID.Valid {
		return models.Expense{}, ErrExpenseLocked
	}
	if createdBy != expense.CreatedBy {
		return models.Expense{}, ErrNotOwner
	}

	if err := s.requireActiveMembers(ctx, tx, expense.GroupID, involvedUsers(expense, splits)); err != nil {
		return models.Expense{}, err
	}

	var updatedAt pgtype.Timestamptz
	err = tx.QueryRow(ctx, `
		UPDATE expenses
		SET paid_by = $1, description = $2, amount_minor = $3, split_type = $4, expense_date = $5, updated_at = now()
		WHERE id = $6 AND group_id = $7
		RETURNING created_at, updated_at`,
		expense.PaidBy, expense.Description, expense.AmountMinor, expense.SplitType,
		expense.ExpenseDate, expense.ID, expense.GroupID,
	).Scan(&expense.CreatedAt, &updatedAt)
	if err != nil {
		return models.Expense{}, err
	}
	if updatedAt.Valid {
		expense.UpdatedAt = &updatedAt.Time
	}

	if _, err := tx.Exec(ctx, `DELETE FROM expense_splits WHERE expense_id = $1`, expense.ID); err != nil {
		return models.Expense{}, err
	}
	for index := range splits {
		splits[index].ExpenseID = expense.ID
		if _, err := tx.Exec(ctx, `
			INSERT INTO expense_splits (expense_id, user_id, share_minor)
			VALUES ($1, $2, $3)`,
			expense.ID, splits[index].UserID, splits[index].ShareMinor,
		); err != nil {
			return models.Expense{}, err
		}
	}

	if err := audit.Append(ctx, tx, audit.Entry{
		GroupID:    expense.GroupID,
		ActorID:    expense.CreatedBy,
		Action:     "expense.updated",
		EntityType: "expense",
		EntityID:   expense.ID,
		Metadata:   map[string]any{"amountMinor": expense.AmountMinor},
	}); err != nil {
		return models.Expense{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.Expense{}, err
	}
	return expense, nil
}

// GetExpenseWithSplits trả về khoản chi cùng phần chia theo group và id.
func (s *PostgresStore) GetExpenseWithSplits(ctx context.Context, groupID, expenseID string) (models.Expense, []models.ExpenseSplit, error) {
	expense, err := scanSingleExpense(s.pool.QueryRow(ctx, `
		SELECT id, group_id, created_by, paid_by, description, amount_minor, split_type,
		       expense_date, batch_id, status, created_at, updated_at
		FROM expenses
		WHERE id = $1 AND group_id = $2`, expenseID, groupID))
	if err != nil {
		return models.Expense{}, nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, expense_id, user_id, share_minor, created_at
		FROM expense_splits
		WHERE expense_id = $1`, expenseID)
	if err != nil {
		return models.Expense{}, nil, err
	}
	splits, err := scanSplits(rows)
	if err != nil {
		return models.Expense{}, nil, err
	}
	return expense, splits, nil
}

// ListUnsettledExpensesWithSplits trả về các khoản chi ACTIVE chưa thuộc kỳ nào.
func (s *PostgresStore) ListUnsettledExpensesWithSplits(ctx context.Context, groupID string) ([]models.Expense, []models.ExpenseSplit, error) {
	expenseRows, err := s.pool.Query(ctx, `
		SELECT id, group_id, created_by, paid_by, description, amount_minor, split_type,
		       expense_date, batch_id, status, created_at, updated_at
		FROM expenses
		WHERE group_id = $1 AND status = $2 AND batch_id IS NULL
		ORDER BY expense_date DESC, created_at DESC`, groupID, models.ExpenseStatusActive)
	if err != nil {
		return nil, nil, err
	}
	expenses, err := scanExpenses(expenseRows)
	if err != nil {
		return nil, nil, err
	}

	splitRows, err := s.pool.Query(ctx, `
		SELECT es.id, es.expense_id, es.user_id, es.share_minor, es.created_at
		FROM expense_splits es
		JOIN expenses e ON e.id = es.expense_id
		WHERE e.group_id = $1 AND e.status = $2 AND e.batch_id IS NULL`, groupID, models.ExpenseStatusActive)
	if err != nil {
		return nil, nil, err
	}
	splits, err := scanSplits(splitRows)
	if err != nil {
		return nil, nil, err
	}
	return expenses, splits, nil
}

// requireActiveMembers xác nhận mọi user trong danh sách đều là thành viên ACTIVE
// của nhóm; list được khử trùng để đếm chính xác qua ANY.
func (s *PostgresStore) requireActiveMembers(ctx context.Context, tx pgx.Tx, groupID string, userIDs []string) error {
	unique := make([]string, 0, len(userIDs))
	seen := make(map[string]struct{})
	for _, id := range userIDs {
		if _, ok := seen[id]; ok || id == "" {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	if len(unique) == 0 {
		return ErrNotActiveMember
	}

	var matched int
	err := tx.QueryRow(ctx, `
		SELECT count(*)
		FROM group_members
		WHERE group_id = $1 AND user_id = ANY($2) AND status = $3`,
		groupID, unique, models.MemberStatusActive,
	).Scan(&matched)
	if err != nil {
		return err
	}
	if matched != len(unique) {
		return ErrNotActiveMember
	}
	return nil
}

func involvedUsers(expense models.Expense, splits []models.ExpenseSplit) []string {
	userIDs := []string{expense.CreatedBy, expense.PaidBy}
	for _, split := range splits {
		userIDs = append(userIDs, split.UserID)
	}
	return userIDs
}

func scanExpense(row pgx.Row) (models.Expense, error) {
	var expense models.Expense
	var batchID pgtype.UUID
	var updatedAt pgtype.Timestamptz
	err := row.Scan(&expense.ID, &expense.GroupID, &expense.CreatedBy, &expense.PaidBy, &expense.Description,
		&expense.AmountMinor, &expense.SplitType, &expense.ExpenseDate, &batchID, &expense.Status,
		&expense.CreatedAt, &updatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Expense{}, ErrExpenseNotFound
	}
	if err != nil {
		return models.Expense{}, err
	}
	if batchID.Valid {
		batchIDString := batchID.String()
		expense.BatchID = &batchIDString
	}
	if updatedAt.Valid {
		expense.UpdatedAt = &updatedAt.Time
	}
	return expense, nil
}

func scanSingleExpense(row pgx.Row) (models.Expense, error) {
	return scanExpense(row)
}

func scanExpenses(rows pgx.Rows) ([]models.Expense, error) {
	defer rows.Close()
	var expenses []models.Expense
	for rows.Next() {
		expense, err := scanExpense(rows)
		if err != nil {
			return nil, err
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
