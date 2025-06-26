package service

import (
	"github.com/hedeqiang/skeleton/internal/model"
	"github.com/hedeqiang/skeleton/pkg/mq"
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// HelloService Hello消息服务接口
type HelloService interface {
	PublishHelloMessage(ctx context.Context, req *model.PublishHelloRequest) (string, error)
}

// helloService Hello消息服务实现
type helloService struct {
	mqProducer *mq.Producer
}

// NewHelloService 创建Hello消息服务实例
func NewHelloService(mqProducer *mq.Producer) HelloService {
	return &helloService{
		mqProducer: mqProducer,
	}
}

// PublishHelloMessage 发布Hello消息到队列
func (s *helloService) PublishHelloMessage(ctx context.Context, req *model.PublishHelloRequest) (string, error) {
	messageID := fmt.Sprintf("msg-%d", time.Now().UnixNano())

	// 创建消息结构
	message := struct {
		MessageID   string `json:"message_id"`
		MessageType string `json:"message_type"`
		Payload     struct {
			Content   string `json:"content"`
			Sender    string `json:"sender"`
			Timestamp int64  `json:"timestamp"`
		} `json:"payload"`
		Timestamp int64 `json:"timestamp"`
	}{
		MessageID:   messageID,
		MessageType: "hello",
		Payload: struct {
			Content   string `json:"content"`
			Sender    string `json:"sender"`
			Timestamp int64  `json:"timestamp"`
		}{
			Content:   req.Content,
			Sender:    req.Sender,
			Timestamp: time.Now().Unix(),
		},
		Timestamp: time.Now().Unix(),
	}

	// 序列化消息
	body, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	// 创建AMQP消息
	amqpMsg := amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
		MessageId:    messageID,
		Timestamp:    time.Now(),
	}

	// 发布消息到队列
	if err := s.mqProducer.Publish(ctx, "hello.exchange", "hello", amqpMsg); err != nil {
		return "", fmt.Errorf("failed to publish message to queue: %w", err)
	}

	return messageID, nil
}
