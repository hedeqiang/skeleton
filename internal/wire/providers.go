package wire

import (
	"errors"

	"github.com/hedeqiang/skeleton/internal/app"
	"github.com/hedeqiang/skeleton/internal/config"
	v1 "github.com/hedeqiang/skeleton/internal/handler/v1"
	"github.com/hedeqiang/skeleton/internal/repository"
	"github.com/hedeqiang/skeleton/internal/scheduler"
	"github.com/hedeqiang/skeleton/internal/service"
	"github.com/hedeqiang/skeleton/pkg/database"
	"github.com/hedeqiang/skeleton/pkg/i18n"
	"github.com/hedeqiang/skeleton/pkg/idgen"
	"github.com/hedeqiang/skeleton/pkg/logger"
	"github.com/hedeqiang/skeleton/pkg/mq"
	redispkg "github.com/hedeqiang/skeleton/pkg/redis"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

var (
	// ErrMainDatabaseNotFound 主数据库未找到错误
	ErrMainDatabaseNotFound = errors.New("main database connection not found")
)

// InfrastructureSet 基础设施层提供者集合
var InfrastructureSet = wire.NewSet(
	// 配置
	config.LoadConfig,
	ProvideLoggerConfig,
	ProvideDatabasesConfig,
	ProvideRedisConfig,
	ProvideRabbitMQConfig,
	ProvideI18nConfig,

	// 日志
	logger.New,

	// 国际化
	ProvideI18n,

	// 数据库
	database.NewDatabases,
	ProvideMainDatabase,

	// Redis
	redispkg.NewRedis,

	// RabbitMQ
	mq.NewRabbitMQ,
	ProvideProducer,

	// ID生成器
	ProvideIDGenerator,
)

// RepositorySet Repository 层提供者集合
var RepositorySet = wire.NewSet(
	repository.NewUserRepository,
)

// ServiceSet Service 层提供者集合
var ServiceSet = wire.NewSet(
	service.NewUserService,
	service.NewHelloService,
)

// HandlerSet Handler 层提供者集合
var HandlerSet = wire.NewSet(
	v1.NewUserHandler,
	v1.NewHelloHandler,
	v1.NewSchedulerHandler,
)

// SchedulerSet 调度器相关依赖
var SchedulerSet = wire.NewSet(
	ProvideSchedulerService,
	ProvideJobRegistry,
)

// AppSet App 层提供者集合
var AppSet = wire.NewSet(
	ProvideApp,
)

// AllSet 所有提供者的集合
var AllSet = wire.NewSet(
	InfrastructureSet,
	RepositorySet,
	ServiceSet,
	HandlerSet,
	SchedulerSet,
	AppSet,
)

// ProvideMainDatabase 提供主数据库连接
func ProvideMainDatabase(dataSources map[string]*gorm.DB) (*gorm.DB, error) {
	db, exists := dataSources["primary"]
	if !exists {
		return nil, ErrMainDatabaseNotFound
	}
	return db, nil
}

// ProvideLoggerConfig 提供日志配置
func ProvideLoggerConfig(cfg *config.Config) *config.Logger {
	return &cfg.Logger
}

// ProvideDatabasesConfig 提供数据库配置
func ProvideDatabasesConfig(cfg *config.Config) map[string]config.Database {
	return cfg.Databases
}

// ProvideRedisConfig 提供Redis配置
func ProvideRedisConfig(cfg *config.Config) *config.Redis {
	return &cfg.Redis
}

// ProvideRabbitMQConfig 提供RabbitMQ配置
func ProvideRabbitMQConfig(cfg *config.Config) *config.RabbitMQ {
	return &cfg.RabbitMQ
}

// ProvideI18nConfig 提供I18n配置
func ProvideI18nConfig(cfg *config.Config) *config.I18nConfig {
	return &cfg.I18n
}

// ProvideI18n 提供I18n实例
func ProvideI18n(cfg *config.I18nConfig, logger *zap.Logger) (*i18n.I18n, error) {
	i18nConfig := i18n.Config{
		DefaultLanguage: cfg.DefaultLanguage,
		SupportLangs:    cfg.SupportLanguages,
		MessagesPath:    cfg.MessagesPath,
	}

	i18n, err := i18n.New(i18nConfig)
	if err != nil {
		logger.Error("Failed to create i18n", zap.Error(err))
		return nil, err
	}

	logger.Info("I18n created successfully",
		zap.String("default_language", i18nConfig.DefaultLanguage),
		zap.Strings("support_languages", i18nConfig.SupportLangs))
	return i18n, nil
}

// ProvideProducer 提供 MQ Producer
func ProvideProducer(conn *amqp.Connection) *mq.Producer {
	return mq.NewProducer(conn)
}

// ProvideSchedulerService 提供调度器服务
func ProvideSchedulerService(logger *zap.Logger) (*scheduler.SchedulerService, error) {
	return scheduler.NewSchedulerService(logger)
}

// ProvideJobRegistry 提供任务注册器
func ProvideJobRegistry(schedulerService *scheduler.SchedulerService, logger *zap.Logger, cfg *config.Config) *scheduler.JobRegistry {
	return scheduler.NewJobRegistry(schedulerService, logger, cfg.Scheduler)
}

// ProvideApp 提供应用实例
func ProvideApp(
	logger *zap.Logger,
	config *config.Config,
	dataSources map[string]*gorm.DB,
	mainDB *gorm.DB,
	redisClient *redis.Client,
	rabbitMQ *amqp.Connection,
	idGenerator idgen.IDGenerator,
	i18n *i18n.I18n,
	userHandler *v1.UserHandler,
	helloHandler *v1.HelloHandler,
	schedulerHandler *v1.SchedulerHandler,
	jobRegistry *scheduler.JobRegistry,
) *app.App {
	return app.NewApp(
		logger,
		config,
		dataSources,
		mainDB,
		redisClient,
		rabbitMQ,
		idGenerator,
		i18n,
		userHandler,
		helloHandler,
		schedulerHandler,
		jobRegistry,
	)
}

// ProvideIDGenerator 提供ID生成器
func ProvideIDGenerator(cfg *config.Config, logger *zap.Logger) (idgen.IDGenerator, error) {
	// 如果配置中有ID生成器配置，使用自定义配置
	if cfg.IDGenerator != nil {
		config := idgen.Config{
			StartTime:     cfg.IDGenerator.StartTime,
			MachineID:     cfg.IDGenerator.MachineID,
			BitsSequence:  cfg.IDGenerator.BitsSequence,
			BitsMachineID: cfg.IDGenerator.BitsMachineID,
			TimeUnit:      cfg.IDGenerator.TimeUnit,
		}

		generator, err := idgen.NewSonyflakeGeneratorWithConfig(config)
		if err != nil {
			logger.Error("Failed to create ID generator with config", zap.Error(err))
			return nil, err
		}

		logger.Info("ID generator created with custom config")
		return generator, nil
	}

	// 使用默认配置
	generator, err := idgen.NewSonyflakeGenerator()
	if err != nil {
		logger.Error("Failed to create default ID generator", zap.Error(err))
		return nil, err
	}

	logger.Info("Default ID generator created")
	return generator, nil
}
