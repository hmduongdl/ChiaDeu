package groups

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/hmduongdl/ChiaDeu/models"
)

var (
	ErrInvalidGroupName = errors.New("tên nhóm phải dài từ 1 đến 100 ký tự")
	ErrInvalidCurrency  = errors.New("loại tiền tệ phải là mã gồm 3 chữ cái ISO 4217")
	ErrInvalidShareCode = errors.New("mã chia sẻ không hợp lệ")
	ErrGroupArchived    = errors.New("nhóm đã bị lưu trữ")
)

// shareCodeAlphabet bỏ các ký tự dễ nhầm lẫn (0/O, 1/I) để mã dễ đọc khi chia sẻ.
const shareCodeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

const shareCodeLength = 8

// Service triển khai nghiệp vụ nhóm trên nền Store.
type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

// GroupDetail gồm thông tin nhóm và các thành viên đang hoạt động.
type GroupDetail struct {
	Group   models.Group
	Members []models.User
}

func (s *Service) CreateGroup(ctx context.Context, actorID, name, currency string) (models.Group, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 100 {
		return models.Group{}, ErrInvalidGroupName
	}
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		currency = "VND"
	}
	if len(currency) != 3 {
		return models.Group{}, ErrInvalidCurrency
	}

	exists, err := s.store.UserExists(ctx, actorID)
	if err != nil {
		return models.Group{}, err
	}
	if !exists {
		return models.Group{}, ErrUserNotFound
	}

	shareCode, err := s.generateUniqueShareCode(ctx)
	if err != nil {
		return models.Group{}, err
	}

	group := models.Group{
		Name:      name,
		CreatedBy: actorID,
		ShareCode: shareCode,
		Currency:  currency,
		Status:    models.GroupStatusActive,
	}
	admin := models.GroupMember{
		GroupID:  group.ID,
		UserID:   actorID,
		Role:     models.RoleAdmin,
		Status:   models.MemberStatusActive,
		JoinedAt: time.Now(),
	}
	created, err := s.store.CreateGroupWithAdmin(ctx, group, admin)
	if err != nil {
		return models.Group{}, err
	}
	return created, nil
}

func (s *Service) JoinGroup(ctx context.Context, actorID, shareCode string) (models.Group, error) {
	shareCode = models.NormalizeShareCode(shareCode)
	if shareCode == "" {
		return models.Group{}, ErrInvalidShareCode
	}

	group, err := s.store.GetGroupByShareCode(ctx, shareCode)
	if errors.Is(err, ErrGroupNotFound) {
		return models.Group{}, ErrInvalidShareCode
	}
	if err != nil {
		return models.Group{}, err
	}
	if group.Status != models.GroupStatusActive {
		return models.Group{}, ErrGroupArchived
	}

	exists, err := s.store.UserExists(ctx, actorID)
	if err != nil {
		return models.Group{}, err
	}
	if !exists {
		return models.Group{}, ErrUserNotFound
	}

	member := models.GroupMember{
		GroupID:  group.ID,
		UserID:   actorID,
		Role:     models.RoleMember,
		Status:   models.MemberStatusActive,
		JoinedAt: time.Now(),
	}
	if err := s.store.CreateMembership(ctx, member); err != nil {
		return models.Group{}, err
	}
	return group, nil
}

func (s *Service) GetGroup(ctx context.Context, actorID, groupID string) (GroupDetail, error) {
	group, err := s.store.GetGroup(ctx, groupID)
	if errors.Is(err, ErrGroupNotFound) {
		return GroupDetail{}, ErrGroupNotFound
	}
	if err != nil {
		return GroupDetail{}, err
	}

	member, err := s.requireActiveMember(ctx, groupID, actorID)
	if err != nil {
		return GroupDetail{}, err
	}
	_ = member

	members, err := s.store.ListActiveMembers(ctx, groupID)
	if err != nil {
		return GroupDetail{}, err
	}
	return GroupDetail{Group: group, Members: members}, nil
}

func (s *Service) ListGroups(ctx context.Context, actorID string) ([]models.Group, error) {
	return s.store.ListUserGroups(ctx, actorID)
}

// requireActiveMember trả về membership nếu người dùng đang hoạt động trong nhóm.
func (s *Service) requireActiveMember(ctx context.Context, groupID, userID string) (models.GroupMember, error) {
	member, err := s.store.GetMembership(ctx, groupID, userID)
	if errors.Is(err, ErrMembershipMissing) {
		return models.GroupMember{}, ErrNotMember
	}
	if err != nil {
		return models.GroupMember{}, err
	}
	if !member.IsActiveMember() {
		return models.GroupMember{}, ErrNotMember
	}
	return member, nil
}

// CanManageGroup cho phép các package khác kiểm tra quyền ADMIN của một user.
func (s *Service) CanManageGroup(ctx context.Context, groupID, userID string) (bool, error) {
	member, err := s.requireActiveMember(ctx, groupID, userID)
	if err != nil {
		return false, err
	}
	return member.IsAdmin(), nil
}

// generateUniqueShareCode tạo mã chia sẻ chưa bị dùng; thử lại tối đa 5 lần khi
// va chạm hiếm gặp trước khi báo lỗi.
func (s *Service) generateUniqueShareCode(ctx context.Context) (string, error) {
	for attempt := 0; attempt < 5; attempt++ {
		code, err := randomShareCode()
		if err != nil {
			return "", fmt.Errorf("generate share code: %w", err)
		}
		if _, err := s.store.GetGroupByShareCode(ctx, code); errors.Is(err, ErrGroupNotFound) {
			return code, nil
		} else if err != nil {
			return "", err
		}
	}
	return "", errors.New("không tạo được mã chia sẻ sau nhiều lần thử")
}

func randomShareCode() (string, error) {
	max := big.NewInt(int64(len(shareCodeAlphabet)))
	buffer := make([]byte, shareCodeLength)
	for index := range buffer {
		value, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		buffer[index] = shareCodeAlphabet[value.Int64()]
	}
	return string(buffer), nil
}
