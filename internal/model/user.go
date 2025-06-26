package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null;size:50" validate:"required,min=3,max=50"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null;size:100" validate:"required,email"`
	Password  string         `json:"-" gorm:"not null;size:255" validate:"required,min=6"`
	Status    int            `json:"status" gorm:"default:1;comment:用户状态 1-正常 0-禁用"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username string `json:"username" validate:"omitempty,min=3,max=50"`
	Email    string `json:"email" validate:"omitempty,email"`
	Status   *int   `json:"status" validate:"omitempty,oneof=0 1"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
