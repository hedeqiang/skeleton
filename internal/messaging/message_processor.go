package messaging

import (
	"github.com/hedeqiang/skeleton/internal/app"
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

// BusinessMessage 业务消息接口
type BusinessMessage interface {
	GetMessageType() string
	GetMessageID() string
}

// MessageProcessor 消息处理器接口
type MessageProcessor interface {
	ProcessMessage(ctx context.Context, msg BusinessMessage, app *app.App) error
	GetSupportedMessageType() string
}

// ProcessorRegistry 消息处理器注册表
type ProcessorRegistry struct {
	processors map[string]MessageProcessor
	logger     *zap.Logger
}

// NewProcessorRegistry 创建新的处理器注册表
func NewProcessorRegistry(logger *zap.Logger) *ProcessorRegistry {
	return &ProcessorRegistry{
		processors: make(map[string]MessageProcessor),
		logger:     logger,
	}
}

// RegisterProcessor 注册消息处理器
func (r *ProcessorRegistry) RegisterProcessor(processor MessageProcessor) {
	messageType := processor.GetSupportedMessageType()
	r.processors[messageType] = processor
	r.logger.Info("Message processor registered",
		zap.String("message_type", messageType),
		zap.String("processor", fmt.Sprintf("%T", processor)),
	)
}

// ProcessIncomingMessage 处理接收到的消息
func (r *ProcessorRegistry) ProcessIncomingMessage(ctx context.Context, body []byte, app *app.App) error {
	// 先尝试解析基础消息结构
	var envelope MessageEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		r.logger.Error("Failed to unmarshal message envelope", zap.Error(err))
		return fmt.Errorf("failed to unmarshal message envelope: %w", err)
	}

	r.logger.Info("Received business message",
		zap.String("message_id", envelope.MessageID),
		zap.String("message_type", envelope.MessageType),
		zap.ByteString("payload", body),
	)

	// 查找对应的处理器
	processor, exists := r.processors[envelope.MessageType]
	if !exists {
		r.logger.Warn("No processor found for message type",
			zap.String("message_type", envelope.MessageType),
		)
		// 可以选择返回错误或者忽略
		return nil
	}

	// 让具体的处理器解析和处理消息
	return processor.ProcessMessage(ctx, &envelope, app)
}

// MessageEnvelope 消息信封结构
type MessageEnvelope struct {
	MessageID   string          `json:"message_id"`
	MessageType string          `json:"message_type"`
	Payload     json.RawMessage `json:"payload"` // 使用 RawMessage 延迟解析
	Timestamp   int64           `json:"timestamp"`
	Source      string          `json:"source,omitempty"`
	Version     string          `json:"version,omitempty"`
}

// GetMessageType 实现 BusinessMessage 接口
func (e *MessageEnvelope) GetMessageType() string {
	return e.MessageType
}

// GetMessageID 实现 BusinessMessage 接口
func (e *MessageEnvelope) GetMessageID() string {
	return e.MessageID
}

// UnmarshalPayload 解析消息载荷到具体结构
func (e *MessageEnvelope) UnmarshalPayload(v interface{}) error {
	return json.Unmarshal(e.Payload, v)
}

// GetRegisteredTypes 获取所有已注册的消息处理器类型
func (r *ProcessorRegistry) GetRegisteredTypes() []string {
	r.logger.Debug("Getting registered processor types")

	types := make([]string, 0, len(r.processors))
	for messageType := range r.processors {
		types = append(types, messageType)
	}

	return types
}
