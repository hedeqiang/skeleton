package jobs

import (
	"time"

	"go.uber.org/zap"
)

// HelloJob Hello计划任务
type HelloJob struct {
	logger *zap.Logger
}

// NewHelloJob 创建Hello任务
func NewHelloJob(logger *zap.Logger) *HelloJob {
	return &HelloJob{
		logger: logger,
	}
}

// Execute 执行任务
func (j *HelloJob) Execute() {
	j.logger.Info("Hello scheduled job executed",
		zap.Time("executed_at", time.Now()),
		zap.String("job_type", "hello"),
	)
}

// Name 任务名称
func (j *HelloJob) Name() string {
	return "hello_job"
}

// Description 任务描述
func (j *HelloJob) Description() string {
	return "Hello world scheduled job for demonstration"
}
