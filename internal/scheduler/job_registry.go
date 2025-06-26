package scheduler

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"

	"github.com/hedeqiang/skeleton/internal/config"
	"github.com/hedeqiang/skeleton/internal/scheduler/jobs"
)

// JobRegistry 任务注册器，负责任务的注册、初始化和生命周期管理
type JobRegistry struct {
	scheduler      *SchedulerService
	logger         *zap.Logger
	config         config.SchedulerConfig
	registeredJobs map[string]JobFactory
}

// JobFactory 任务工厂函数类型
type JobFactory func(*zap.Logger) Job

// Job 任务接口
type Job interface {
	Execute()
	Name() string
	Description() string
}

// NewJobRegistry 创建任务注册器
func NewJobRegistry(schedulerService *SchedulerService, logger *zap.Logger, config config.SchedulerConfig) *JobRegistry {
	registry := &JobRegistry{
		scheduler:      schedulerService,
		logger:         logger,
		config:         config,
		registeredJobs: make(map[string]JobFactory),
	}

	// 注册默认任务
	registry.registerDefaultJobs()

	return registry
}

// registerDefaultJobs 注册默认任务
func (r *JobRegistry) registerDefaultJobs() {
	r.registeredJobs["hello_job"] = func(logger *zap.Logger) Job {
		return jobs.NewHelloJob(logger)
	}
}

// RegisterJob 注册自定义任务
func (r *JobRegistry) RegisterJob(name string, factory JobFactory) {
	r.registeredJobs[name] = factory
	r.logger.Info("Custom job registered", zap.String("job_name", name))
}

// InitializeJobs 根据配置初始化任务
func (r *JobRegistry) InitializeJobs() error {
	if !r.config.Enabled {
		r.logger.Info("Scheduler is disabled")
		return nil
	}

	for _, jobConfig := range r.config.Jobs {
		if !jobConfig.Enabled {
			r.logger.Info("Job is disabled, skipping",
				zap.String("job_name", jobConfig.Name))
			continue
		}

		if err := r.addJob(jobConfig); err != nil {
			return fmt.Errorf("failed to add job %s: %w", jobConfig.Name, err)
		}
	}

	return nil
}

// addJob 根据配置添加单个任务
func (r *JobRegistry) addJob(jobConfig config.SchedulerJobConfig) error {
	factory, exists := r.registeredJobs[jobConfig.Name]
	if !exists {
		return fmt.Errorf("job factory not found for: %s", jobConfig.Name)
	}

	job := factory(r.logger)

	// 创建任务定义
	jobDefinition, err := r.createJobDefinition(jobConfig)
	if err != nil {
		return fmt.Errorf("failed to create job definition: %w", err)
	}

	// 创建任务
	task := gocron.NewTask(job.Execute)

	// 添加到调度器
	if err := r.scheduler.AddJob(jobDefinition, task,
		gocron.WithTags(jobConfig.Name, jobConfig.Type),
		gocron.WithName(jobConfig.Name),
	); err != nil {
		return fmt.Errorf("failed to add job to scheduler: %w", err)
	}

	r.logger.Info("Job initialized successfully",
		zap.String("job_name", jobConfig.Name),
		zap.String("job_type", jobConfig.Type),
		zap.String("schedule", jobConfig.Schedule),
		zap.String("description", jobConfig.Description),
	)

	return nil
}

// createJobDefinition 根据配置创建任务定义
func (r *JobRegistry) createJobDefinition(jobConfig config.SchedulerJobConfig) (gocron.JobDefinition, error) {
	switch jobConfig.Type {
	case "duration":
		duration, err := time.ParseDuration(jobConfig.Schedule)
		if err != nil {
			return nil, fmt.Errorf("invalid duration format: %w", err)
		}
		return gocron.DurationJob(duration), nil

	case "cron":
		return gocron.CronJob(jobConfig.Schedule, false), nil

	case "daily":
		// 解析时间格式，例如 "14:30" 表示每天14:30
		t, err := time.Parse("15:04", jobConfig.Schedule)
		if err != nil {
			return nil, fmt.Errorf("invalid daily time format (should be HH:MM): %w", err)
		}
		return gocron.DailyJob(1, gocron.NewAtTimes(
			gocron.NewAtTime(uint(t.Hour()), uint(t.Minute()), 0),
		)), nil

	default:
		return nil, fmt.Errorf("unsupported job type: %s", jobConfig.Type)
	}
}

// Start 启动任务注册器
func (r *JobRegistry) Start() error {
	if !r.config.Enabled {
		r.logger.Info("Scheduler is disabled, not starting")
		return nil
	}

	if err := r.InitializeJobs(); err != nil {
		return fmt.Errorf("failed to initialize jobs: %w", err)
	}

	r.scheduler.Start()
	return nil
}

// Stop 停止任务注册器
func (r *JobRegistry) Stop() error {
	return r.scheduler.Stop()
}

// GetJobsStatus 获取任务状态
func (r *JobRegistry) GetJobsStatus() []JobInfo {
	return r.scheduler.GetJobs()
}
