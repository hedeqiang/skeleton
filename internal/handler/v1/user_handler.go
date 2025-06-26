package v1

import (
	"github.com/hedeqiang/skeleton/internal/model"
	"github.com/hedeqiang/skeleton/internal/service"
	"github.com/hedeqiang/skeleton/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService service.UserService
	logger      *zap.Logger
	validator   *validator.Validate
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
		validator:   validator.New(),
	}
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body model.CreateUserRequest true "用户信息"
// @Success 201 {object} response.Response{data=model.UserResponse} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "请求参数格式错误")
		return
	}

	// 参数验证
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "请求参数验证失败: "+err.Error())
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessWithMsg(c, http.StatusCreated, "用户创建成功", user)
}

// GetUser 获取用户信息
// @Summary 获取用户信息
// @Description 根据用户ID获取用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=model.UserResponse} "获取成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "用户ID格式错误")
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		if err.Error() == "user not found" {
			response.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessWithMsg(c, http.StatusOK, "获取成功", user)
}

// UpdateUser 更新用户信息
// @Summary 更新用户信息
// @Description 根据用户ID更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param user body model.UpdateUserRequest true "更新信息"
// @Success 200 {object} response.Response{data=model.UserResponse} "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "用户ID格式错误")
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "请求参数格式错误")
		return
	}

	// 参数验证
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "请求参数验证失败: "+err.Error())
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), uint(id), &req)
	if err != nil {
		h.logger.Error("Failed to update user", zap.Error(err))
		if err.Error() == "user not found" {
			response.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessWithMsg(c, http.StatusOK, "更新成功", user)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 根据用户ID删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "用户ID格式错误")
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), uint(id))
	if err != nil {
		h.logger.Error("Failed to delete user", zap.Error(err))
		if err.Error() == "user not found" {
			response.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessWithMsg(c, http.StatusOK, "删除成功", nil)
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=response.PageResponse{list=[]model.UserResponse}} "获取成功"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	users, total, err := h.userService.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("Failed to list users", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	pageResp := response.PageResponse{
		List:     users,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	response.SuccessWithMsg(c, http.StatusOK, "获取成功", pageResp)
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录验证
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param login body LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=model.UserResponse} "登录成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "用户名或密码错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "请求参数格式错误")
		return
	}

	// 参数验证
	if err := h.validator.Struct(&req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "请求参数验证失败: "+err.Error())
		return
	}

	user, err := h.userService.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		h.logger.Error("Failed to login", zap.Error(err))
		if err.Error() == "invalid username or password" || err.Error() == "user account is disabled" {
			response.Error(c, http.StatusUnauthorized, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.SuccessWithMsg(c, http.StatusOK, "登录成功", user)
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
