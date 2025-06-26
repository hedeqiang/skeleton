package v1

import (
	"github.com/gin-gonic/gin"
	handlers "github.com/hedeqiang/skeleton/internal/handler/v1"
)

// RegisterSchedulerRoutes 注册计划任务相关路由
func RegisterSchedulerRoutes(group *gin.RouterGroup, schedulerHandler *handlers.SchedulerHandler) {
	scheduler := group.Group("/scheduler")
	{
		// 基础管理
		scheduler.GET("/jobs", schedulerHandler.GetJobs)          // 获取任务列表
		scheduler.POST("/start", schedulerHandler.StartScheduler) // 启动调度器
		scheduler.POST("/stop", schedulerHandler.StopScheduler)   // 停止调度器

		// 未来可以添加更多调度器功能
		// scheduler.POST("/jobs", schedulerHandler.CreateJob)        // 创建任务
		// scheduler.PUT("/jobs/:id", schedulerHandler.UpdateJob)     // 更新任务
		// scheduler.DELETE("/jobs/:id", schedulerHandler.DeleteJob)  // 删除任务
		// scheduler.POST("/jobs/:id/run", schedulerHandler.RunJob)   // 手动运行任务
		// scheduler.GET("/jobs/:id/logs", schedulerHandler.GetJobLogs) // 获取任务日志
	}
}
