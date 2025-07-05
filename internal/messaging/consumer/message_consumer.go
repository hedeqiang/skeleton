package consumer

import (
	"context"

	"github.com/hedeqiang/skeleton/internal/app"
	"github.com/hedeqiang/skeleton/internal/messaging"
	"github.com/hedeqiang/skeleton/internal/messaging/processors"

	"go.uber.org/zap"
)

// MessageConsumerService 消息消费服务
type MessageConsumerService struct {
	processorRegistry *messaging.ProcessorRegistry
	logger            *zap.Logger
	app               *app.App
}

// NewMessageConsumerService 创建消息消费服务
func NewMessageConsumerService(app *app.App) *MessageConsumerService {
	service := &MessageConsumerService{
		processorRegistry: messaging.NewProcessorRegistry(app.Logger()),
		logger:            app.Logger(),
		app:               app,
	}

	// 注册所有事件处理器
	service.registerEventProcessors()

	return service
}

// registerEventProcessors 注册事件处理器
func (s *MessageConsumerService) registerEventProcessors() {
	// 注册Hello处理器（示例）
	s.processorRegistry.RegisterProcessor(
		processors.NewHelloProcessor(s.logger),
	)

	// TODO: 在这里添加其他消息处理器
	// s.processorRegistry.RegisterProcessor(
	//     processors.NewUserEventProcessor(s.logger),
	// )
	// s.processorRegistry.RegisterProcessor(
	//     processors.NewOrderEventProcessor(s.logger),
	// )

	s.logger.Info("Event processors registered successfully")
}

// ConsumeMessage 消费消息的统一入口
func (s *MessageConsumerService) ConsumeMessage(ctx context.Context, messageBody []byte) error {
	s.logger.Info("Message consumer service received message",
		zap.Int("body_size", len(messageBody)),
	)

	// 委托给处理器注册表进行具体处理
	if err := s.processorRegistry.ProcessIncomingMessage(ctx, messageBody, s.app); err != nil {
		s.logger.Error("Failed to process incoming message", zap.Error(err))
		return err
	}

	s.logger.Info("Message processed successfully by consumer service")
	return nil
}

// GetRegisteredProcessorTypes 获取已注册的处理器类型（用于监控和调试）
func (s *MessageConsumerService) GetRegisteredProcessorTypes() []string {
	return s.processorRegistry.GetRegisteredTypes()
}

// Shutdown 优雅关闭消费服务
func (s *MessageConsumerService) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down message consumer service")
	// 这里可以添加清理逻辑，如等待正在处理的消息完成
	return nil
}
