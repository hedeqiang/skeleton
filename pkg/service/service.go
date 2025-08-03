package service

import (
	"context"
	"fmt"

	"github.com/hedeqiang/skeleton/pkg/errors"
	"go.uber.org/zap"
)

// Service 服务接口
type Service interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Health(ctx context.Context) error
}

// BaseService 基础服务
type BaseService struct {
	name   string
	logger *zap.Logger
}

// NewBaseService 创建基础服务
func NewBaseService(name string, logger *zap.Logger) *BaseService {
	return &BaseService{
		name:   name,
		logger: logger,
	}
}

// Name 获取服务名称
func (s *BaseService) Name() string {
	return s.name
}

// Start 启动服务
func (s *BaseService) Start(ctx context.Context) error {
	s.logger.Info("Starting service", zap.String("service", s.name))
	return nil
}

// Stop 停止服务
func (s *BaseService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping service", zap.String("service", s.name))
	return nil
}

// Health 健康检查
func (s *BaseService) Health(ctx context.Context) error {
	s.logger.Debug("Health check", zap.String("service", s.name))
	return nil
}

// GetLogger 获取日志器
func (s *BaseService) GetLogger() *zap.Logger {
	return s.logger
}

// Services 服务集合
type Services struct {
	services []Service
	logger   *zap.Logger
}

// NewServices 创建服务集合
func NewServices(logger *zap.Logger) *Services {
	return &Services{
		services: make([]Service, 0),
		logger:   logger,
	}
}

// Add 添加服务
func (s *Services) Add(service Service) {
	s.services = append(s.services, service)
	s.logger.Info("Service added", zap.String("service", service.Name()))
}

// StartAll 启动所有服务
func (s *Services) StartAll(ctx context.Context) error {
	for _, service := range s.services {
		if err := service.Start(ctx); err != nil {
			return fmt.Errorf("failed to start service %s: %w", service.Name(), err)
		}
		s.logger.Info("Service started", zap.String("service", service.Name()))
	}
	return nil
}

// StopAll 停止所有服务
func (s *Services) StopAll(ctx context.Context) error {
	for i := len(s.services) - 1; i >= 0; i-- {
		service := s.services[i]
		if err := service.Stop(ctx); err != nil {
			s.logger.Error("Failed to stop service", 
				zap.String("service", service.Name()), 
				zap.Error(err))
		} else {
			s.logger.Info("Service stopped", zap.String("service", service.Name()))
		}
	}
	return nil
}

// HealthAll 检查所有服务健康状态
func (s *Services) HealthAll(ctx context.Context) error {
	for _, service := range s.services {
		if err := service.Health(ctx); err != nil {
			return errors.Wrap(err, errors.ErrorTypeInternal, 
				fmt.Sprintf("service %s health check failed", service.Name()))
		}
		s.logger.Debug("Service health check passed", zap.String("service", service.Name()))
	}
	return nil
}

// List 获取所有服务
func (s *Services) List() []Service {
	return s.services
}

// Get 根据名称获取服务
func (s *Services) Get(name string) Service {
	for _, service := range s.services {
		if service.Name() == name {
			return service
		}
	}
	return nil
}