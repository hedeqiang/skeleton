package scheduler

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
)

// SchedulerService 计划任务调度器服务
type SchedulerService struct {
	scheduler gocron.Scheduler
	logger    *zap.Logger
	jobs      []gocron.Job
}

// NewSchedulerService 创建调度器服务实例
func NewSchedulerService(logger *zap.Logger) (*SchedulerService, error) {
	scheduler, err := gocron.NewScheduler(
		gocron.WithLogger(NewCronLogger(logger)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	return &SchedulerService{
		scheduler: scheduler,
		logger:    logger,
		jobs:      make([]gocron.Job, 0),
	}, nil
}

// AddJob 添加任务
func (s *SchedulerService) AddJob(jobDefinition gocron.JobDefinition, task gocron.Task, options ...gocron.JobOption) error {
	job, err := s.scheduler.NewJob(jobDefinition, task, options...)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	s.jobs = append(s.jobs, job)

	// 获取下次运行时间
	nextRun, err := job.NextRun()
	nextRunStr := "unknown"
	if err == nil {
		nextRunStr = nextRun.Format(time.RFC3339)
	}

	s.logger.Info("Job added successfully",
		zap.String("job_id", job.ID().String()),
		zap.String("next_run", nextRunStr),
	)

	return nil
}

// Start 启动调度器
func (s *SchedulerService) Start() {
	s.scheduler.Start()
	s.logger.Info("Scheduler started", zap.Int("jobs_count", len(s.jobs)))
}

// Stop 停止调度器
func (s *SchedulerService) Stop() error {
	if err := s.scheduler.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown scheduler: %w", err)
	}

	s.logger.Info("Scheduler stopped")
	return nil
}

// GetJobs 获取所有任务信息
func (s *SchedulerService) GetJobs() []JobInfo {
	jobs := make([]JobInfo, len(s.jobs))
	for i, job := range s.jobs {
		nextRun, _ := job.NextRun()
		lastRun, lastRunErr := job.LastRun()

		jobInfo := JobInfo{
			ID:      job.ID().String(),
			NextRun: nextRun,
			Tags:    job.Tags(),
		}

		if lastRunErr == nil {
			jobInfo.LastRun = &lastRun
		}

		jobs[i] = jobInfo
	}
	return jobs
}

// JobInfo 任务信息
type JobInfo struct {
	ID      string     `json:"id"`
	NextRun time.Time  `json:"next_run"`
	LastRun *time.Time `json:"last_run,omitempty"`
	Tags    []string   `json:"tags,omitempty"`
}

// CronLogger gocron日志适配器
type CronLogger struct {
	logger *zap.Logger
}

// NewCronLogger 创建cron日志适配器
func NewCronLogger(logger *zap.Logger) *CronLogger {
	return &CronLogger{logger: logger}
}

// Debug 调试日志
func (l *CronLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Debug(msg, zapFields(keysAndValues)...)
}

// Info 信息日志
func (l *CronLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, zapFields(keysAndValues)...)
}

// Warn 警告日志
func (l *CronLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Warn(msg, zapFields(keysAndValues)...)
}

// Error 错误日志
func (l *CronLogger) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Error(msg, zapFields(keysAndValues)...)
}

// zapFields 转换为zap字段
func zapFields(keysAndValues []interface{}) []zap.Field {
	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := fmt.Sprintf("%v", keysAndValues[i])
			value := keysAndValues[i+1]
			fields = append(fields, zap.Any(key, value))
		}
	}
	return fields
}
