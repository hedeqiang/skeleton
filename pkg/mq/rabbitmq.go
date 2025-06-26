package mq

import (
	"github.com/hedeqiang/skeleton/internal/config"
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// NewRabbitMQ 根据提供的配置初始化 RabbitMQ 连接
func NewRabbitMQ(cfg *config.RabbitMQ) (*amqp.Connection, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Producer 是一个 RabbitMQ 生产者
type Producer struct {
	conn *amqp.Connection
}

// NewProducer 创建一个新的生产者实例
func NewProducer(conn *amqp.Connection) *Producer {
	return &Producer{conn: conn}
}

// Publish 向指定的 exchange 发送一条消息
// exchange: 交换机名称
// routingKey: 路由键
// message: amqp.Publishing 结构，包含了消息体和各种属性
func (p *Producer) Publish(ctx context.Context, exchange, routingKey string, message amqp.Publishing) error {
	// 为保证线程安全，每次发布都创建一个新的 channel
	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	// 使用上下文进行发布
	return ch.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		message,
	)
}

// MessageHandler 定义消息处理函数接口
type MessageHandler func(ctx context.Context, body []byte) error

// Consumer 是一个 RabbitMQ 消费者
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// NewConsumer 创建一个新的消费者实例
func NewConsumer(conn *amqp.Connection) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// 设置 QoS，控制消费者预取消息数量
	err = ch.Qos(1, 0, false)
	if err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	return &Consumer{
		conn:    conn,
		channel: ch,
	}, nil
}

// DeclareExchange 声明交换机
func (c *Consumer) DeclareExchange(name, kind string, durable, autoDelete bool) error {
	return c.channel.ExchangeDeclare(
		name,       // name
		kind,       // type
		durable,    // durable
		autoDelete, // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
}

// DeclareQueue 声明队列
func (c *Consumer) DeclareQueue(name string, durable, autoDelete, exclusive bool) (amqp.Queue, error) {
	return c.channel.QueueDeclare(
		name,       // name
		durable,    // durable
		autoDelete, // delete when unused
		exclusive,  // exclusive
		false,      // no-wait
		nil,        // arguments
	)
}

// BindQueue 绑定队列到交换机
func (c *Consumer) BindQueue(queueName, routingKey, exchangeName string) error {
	return c.channel.QueueBind(
		queueName,    // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,
		nil,
	)
}

// Consume 开始消费消息
func (c *Consumer) Consume(queueName, consumerName string, handler MessageHandler) error {
	msgs, err := c.channel.Consume(
		queueName,    // queue
		consumerName, // consumer
		false,        // auto-ack (设置为false，手动确认)
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	// 创建一个 channel 来接收停止信号
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			ctx := context.Background()

			// 调用业务处理函数
			if err := handler(ctx, d.Body); err != nil {
				// 处理失败，拒绝消息并重新入队
				d.Nack(false, true)
			} else {
				// 处理成功，确认消息
				d.Ack(false)
			}
		}
	}()

	<-forever
	return nil
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	if c.channel != nil {
		return c.channel.Close()
	}
	return nil
}

// SetupInfrastructureFromConfig 根据配置设置RabbitMQ基础设施
func (c *Consumer) SetupInfrastructureFromConfig(cfg *config.RabbitMQ) error {
	// 设置交换机
	for _, exchangeCfg := range cfg.Exchanges {
		if err := c.DeclareExchange(exchangeCfg.Name, exchangeCfg.Type, exchangeCfg.Durable, exchangeCfg.AutoDelete); err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", exchangeCfg.Name, err)
		}
	}

	// 设置队列并绑定
	for _, queueCfg := range cfg.Queues {
		// 声明队列
		if _, err := c.DeclareQueue(queueCfg.Name, queueCfg.Durable, queueCfg.AutoDelete, queueCfg.Exclusive); err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queueCfg.Name, err)
		}

		// 绑定队列到交换机
		for _, routingKey := range queueCfg.RoutingKeys {
			if err := c.BindQueue(queueCfg.Name, routingKey, queueCfg.Exchange); err != nil {
				return fmt.Errorf("failed to bind queue %s to exchange %s with routing key %s: %w",
					queueCfg.Name, queueCfg.Exchange, routingKey, err)
			}
		}
	}

	return nil
}
