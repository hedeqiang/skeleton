package v1

import (
	"net/http"

	"github.com/hedeqiang/skeleton/internal/scheduler"
	"github.com/hedeqiang/skeleton/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SchedulerHandler 计划任务处理器
type SchedulerHandler struct {
	jobRegistry *scheduler.JobRegistry
	logger      *zap.Logger
}

// NewSchedulerHandler 创建计划任务处理器
func NewSchedulerHandler(jobRegistry *scheduler.JobRegistry, logger *zap.Logger) *SchedulerHandler {
	return &SchedulerHandler{
		jobRegistry: jobRegistry,
		logger:      logger,
	}
}

// GetJobs 获取任务列表
// @Summary 获取计划任务列表
// @Description 获取所有计划任务的状态信息
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]scheduler.JobInfo}
// @Router /api/v1/scheduler/jobs [get]
func (h *SchedulerHandler) GetJobs(c *gin.Context) {
	jobs := h.jobRegistry.GetJobsStatus()

	h.logger.Info("Jobs status retrieved",
		zap.Int("jobs_count", len(jobs)),
	)

	response.Success(c, gin.H{
		"jobs":       jobs,
		"jobs_count": len(jobs),
	})
}

// StartScheduler 启动调度器
// @Summary 启动计划任务调度器
// @Description 启动计划任务调度器服务
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=string}
// @Failure 500 {object} response.Response{data=string}
// @Router /api/v1/scheduler/start [post]
func (h *SchedulerHandler) StartScheduler(c *gin.Context) {
	if err := h.jobRegistry.Start(); err != nil {
		h.logger.Error("Failed to start scheduler", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "Failed to start scheduler")
		return
	}

	h.logger.Info("Scheduler started via API")
	response.Success(c, "Scheduler started successfully")
}

// StopScheduler 停止调度器
// @Summary 停止计划任务调度器
// @Description 停止计划任务调度器服务
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=string}
// @Failure 500 {object} response.Response{data=string}
// @Router /api/v1/scheduler/stop [post]
func (h *SchedulerHandler) StopScheduler(c *gin.Context) {
	if err := h.jobRegistry.Stop(); err != nil {
		h.logger.Error("Failed to stop scheduler", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "Failed to stop scheduler")
		return
	}

	h.logger.Info("Scheduler stopped via API")
	response.Success(c, "Scheduler stopped successfully")
}
