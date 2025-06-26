package processors

import (
	"github.com/hedeqiang/skeleton/internal/app"
	"github.com/hedeqiang/skeleton/internal/messaging"
	"context"
	"time"

	"go.uber.org/zap"
)

// HelloProcessor Hello World消息处理器
type HelloProcessor struct {
	logger *zap.Logger
}

// NewHelloProcessor 创建Hello处理器
func NewHelloProcessor(logger *zap.Logger) *HelloProcessor {
	return &HelloProcessor{
		logger: logger,
	}
}

// GetSupportedMessageType 返回支持的消息类型
func (p *HelloProcessor) GetSupportedMessageType() string {
	return "hello"
}

// ProcessMessage 处理Hello消息
func (p *HelloProcessor) ProcessMessage(ctx context.Context, msg messaging.BusinessMessage, app *app.App) error {
	p.logger.Info("Processing hello message", zap.String("message_id", msg.GetMessageID()))

	// 解析具体的消息数据
	var event HelloEvent
	if envelope, ok := msg.(*messaging.MessageEnvelope); ok {
		if err := envelope.UnmarshalPayload(&event); err != nil {
			p.logger.Error("Failed to unmarshal hello event", zap.Error(err))
			return err
		}
	}

	p.logger.Info("Hello event details",
		zap.String("content", event.Content),
		zap.String("sender", event.Sender),
		zap.Int64("timestamp", event.Timestamp),
	)

	// 简单的业务处理逻辑
	if err := p.handleHelloMessage(ctx, &event, app); err != nil {
		p.logger.Error("Failed to handle hello message", zap.Error(err))
		return err
	}

	p.logger.Info("Hello message processed successfully", zap.String("message_id", msg.GetMessageID()))
	return nil
}

// handleHelloMessage 处理Hello消息的业务逻辑
func (p *HelloProcessor) handleHelloMessage(ctx context.Context, event *HelloEvent, app *app.App) error {
	// 1. 记录到Redis (可选)
	if app.Redis != nil {
		key := "hello:messages:" + time.Now().Format("20060102")
		err := app.Redis.LPush(ctx, key, event.Content).Err()
		if err != nil {
			p.logger.Warn("Failed to save hello message to Redis", zap.Error(err))
		} else {
			// 设置过期时间为7天
			app.Redis.Expire(ctx, key, 7*24*time.Hour)
			p.logger.Info("Hello message saved to Redis", zap.String("key", key))
		}
	}

	// 2. 简单的响应逻辑
	p.logger.Info("Hello World response",
		zap.String("original_content", event.Content),
		zap.String("response", "Hello back from processor!"),
		zap.String("sender", event.Sender),
	)

	return nil
}

// HelloEvent Hello事件结构
type HelloEvent struct {
	Content   string `json:"content"`
	Sender    string `json:"sender"`
	Timestamp int64  `json:"timestamp"`
}
