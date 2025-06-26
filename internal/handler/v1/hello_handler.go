package v1

import (
	"net/http"

	"github.com/hedeqiang/skeleton/internal/model"
	"github.com/hedeqiang/skeleton/internal/service"
	"github.com/hedeqiang/skeleton/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// HelloHandler Hello消息处理器
type HelloHandler struct {
	helloService service.HelloService
	logger       *zap.Logger
	validator    *validator.Validate
}

// NewHelloHandler 创建Hello消息处理器实例
func NewHelloHandler(helloService service.HelloService, logger *zap.Logger) *HelloHandler {
	return &HelloHandler{
		helloService: helloService,
		logger:       logger,
		validator:    validator.New(),
	}
}

// PublishHelloMessage 发布Hello消息到队列
// @Summary 发布Hello消息到队列
// @Description 将Hello消息发布到RabbitMQ队列
// @Tags Hello消息管理
// @Accept json
// @Produce json
// @Param hello body model.PublishHelloRequest true "Hello消息信息"
// @Success 200 {object} response.Response{data=map[string]string} "发布成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/hello/publish [post]
func (h *HelloHandler) PublishHelloMessage(c *gin.Context) {
	var req model.PublishHelloRequest
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

	messageID, err := h.helloService.PublishHelloMessage(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to publish hello message", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Info("Hello message published successfully",
		zap.String("message_id", messageID),
		zap.String("content", req.Content),
		zap.String("sender", req.Sender),
	)

	response.SuccessWithMsg(c, http.StatusOK, "Hello消息发布成功", gin.H{
		"message_id": messageID,
	})
}
