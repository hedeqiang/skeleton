package service

import (
	"github.com/hedeqiang/skeleton/internal/model"
	"github.com/hedeqiang/skeleton/internal/repository"
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户服务接口
type UserService interface {
	CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.UserResponse, error)
	GetUser(ctx context.Context, id uint) (*model.UserResponse, error)
	UpdateUser(ctx context.Context, id uint, req *model.UpdateUserRequest) (*model.UserResponse, error)
	DeleteUser(ctx context.Context, id uint) error
	ListUsers(ctx context.Context, page, pageSize int) ([]*model.UserResponse, int64, error)
	Login(ctx context.Context, username, password string) (*model.UserResponse, error)
}

// userService 用户服务实现
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// CreateUser 创建用户
func (s *userService) CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.UserResponse, error) {
	// 检查用户名是否已存在
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	existingUser, err = s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Status:   1,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.toUserResponse(user), nil
}

// GetUser 获取用户
func (s *userService) GetUser(ctx context.Context, id uint) (*model.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return s.toUserResponse(user), nil
}

// UpdateUser 更新用户
func (s *userService) UpdateUser(ctx context.Context, id uint, req *model.UpdateUserRequest) (*model.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 更新字段
	if req.Username != "" {
		// 检查用户名是否已被其他用户使用
		existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check username: %w", err)
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, errors.New("username already exists")
		}
		user.Username = req.Username
	}

	if req.Email != "" {
		// 检查邮箱是否已被其他用户使用
		existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check email: %w", err)
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, errors.New("email already exists")
		}
		user.Email = req.Email
	}

	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.toUserResponse(user), nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers 获取用户列表
func (s *userService) ListUsers(ctx context.Context, page, pageSize int) ([]*model.UserResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	users, total, err := s.userRepo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	responses := make([]*model.UserResponse, len(users))
	for i, user := range users {
		responses[i] = s.toUserResponse(user)
	}

	return responses, total, nil
}

// Login 用户登录
func (s *userService) Login(ctx context.Context, username, password string) (*model.UserResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("user account is disabled")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	return s.toUserResponse(user), nil
}

// toUserResponse 转换为响应格式
func (s *userService) toUserResponse(user *model.User) *model.UserResponse {
	return &model.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
