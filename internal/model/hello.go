package model

// PublishHelloRequest 发布Hello消息到队列请求
type PublishHelloRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000" example:"Hello, World!"`
	Sender  string `json:"sender" validate:"required,min=1,max=100" example:"user123"`
}
