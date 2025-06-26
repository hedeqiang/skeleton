package main

import (
	"github.com/hedeqiang/skeleton/internal/app"
	"github.com/hedeqiang/skeleton/internal/messaging/consumer"
	"github.com/hedeqiang/skeleton/internal/wire"
	"github.com/hedeqiang/skeleton/pkg/mq"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	// 使用 Wire 初始化应用
	application, err := wire.InitializeApplication()
	if err != nil {
		fmt.Printf("Failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	application.Logger.Info("Starting message consumer service...")

	// 创建消息消费服务（自动注册所有事件处理器）
	messageConsumerService := consumer.NewMessageConsumerService(application)

	// 创建 RabbitMQ Consumer
	rabbitConsumer, err := mq.NewConsumer(application.RabbitMQ)
	if err != nil {
		application.Logger.Fatal("Failed to create RabbitMQ consumer", zap.Error(err))
	}
	defer rabbitConsumer.Close()

	// 使用配置化的方式设置 RabbitMQ 基础设施（避免重复定义）
	if err := rabbitConsumer.SetupInfrastructureFromConfig(&application.Config.RabbitMQ); err != nil {
		application.Logger.Fatal("Failed to setup RabbitMQ infrastructure from config", zap.Error(err))
	}

	// 启动消息消费
	if err := startMessageConsumption(application, messageConsumerService, rabbitConsumer); err != nil {
		application.Logger.Fatal("Failed to start message consumption", zap.Error(err))
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	application.Logger.Info("Message consumer service is running. Press Ctrl+C to exit.")
	<-quit

	application.Logger.Info("Received shutdown signal, stopping message consumer service...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := messageConsumerService.Shutdown(ctx); err != nil {
		application.Logger.Error("Error during message consumer service shutdown", zap.Error(err))
	}

	if err := application.Stop(); err != nil {
		application.Logger.Error("Error during application shutdown", zap.Error(err))
	}

	application.Logger.Info("Message consumer service stopped gracefully")
}

// startMessageConsumption 启动消息消费
func startMessageConsumption(app *app.App, messageConsumerService *consumer.MessageConsumerService, rabbitConsumer *mq.Consumer) error {
	app.Logger.Info("Starting message consumption...")

	// 从配置中获取队列名称
	if len(app.Config.RabbitMQ.Queues) == 0 {
		return fmt.Errorf("no queues configured")
	}

	// 为每个配置的队列启动消费者
	for _, queueConfig := range app.Config.RabbitMQ.Queues {
		if err := startQueueConsumer(app, messageConsumerService, rabbitConsumer, queueConfig.Name); err != nil {
			return fmt.Errorf("failed to start consumer for queue %s: %w", queueConfig.Name, err)
		}
	}

	app.Logger.Info("Message consumers started successfully",
		zap.Strings("supported_message_types", messageConsumerService.GetRegisteredProcessorTypes()),
	)

	return nil
}

// startQueueConsumer 启动单个队列的消费者
func startQueueConsumer(app *app.App, messageConsumerService *consumer.MessageConsumerService, rabbitConsumer *mq.Consumer, queueName string) error {
	app.Logger.Info("Starting consumer for queue", zap.String("queue", queueName))

	// 创建消息处理函数
	messageHandler := func(ctx context.Context, body []byte) error {
		app.Logger.Info("Processing message from queue",
			zap.String("queue", queueName),
			zap.Int("body_size", len(body)),
		)

		// 委托给消息消费服务处理
		if err := messageConsumerService.ConsumeMessage(ctx, body); err != nil {
			app.Logger.Error("Failed to consume message",
				zap.Error(err),
				zap.String("queue", queueName),
			)
			return err
		}

		app.Logger.Debug("Message processed successfully",
			zap.String("queue", queueName),
		)
		return nil
	}

	// 启动消费协程（mq.Consumer.Consume 会阻塞，所以放在 goroutine 中）
	go func() {
		app.Logger.Info("Started consuming messages from queue",
			zap.String("queue", queueName),
		)

		if err := rabbitConsumer.Consume(queueName, "", messageHandler); err != nil {
			app.Logger.Error("Consumer stopped with error",
				zap.Error(err),
				zap.String("queue", queueName),
			)
		}
	}()

	return nil
}
