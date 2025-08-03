package database

import (
	"context"
	"fmt"
	"time"

	"github.com/hedeqiang/skeleton/internal/config"
	"github.com/hedeqiang/skeleton/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"log"
	"os"
)

// NewDatabases 初始化所有在配置中定义的数据源
func NewDatabases(dbConfigs map[string]config.Database) (map[string]*gorm.DB, error) {
	dataSources := make(map[string]*gorm.DB)

	for name, cfg := range dbConfigs {
		db, err := connect(&cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to data source [%s]: %w", name, err)
		}

		dataSources[name] = db
	}

	return dataSources, nil
}

func connect(cfg *config.Database) (*gorm.DB, error) {
	var dialector gorm.Dialector
	switch cfg.Type {
	case "mysql":
		dialector = mysql.Open(cfg.DSN)
	case "postgres":
		dialector = postgres.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	// 配置 GORM logger
	gormLog := gormlogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormlogger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLog,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}

// DBConfig 数据库配置
type DBConfig struct {
	Driver             string
	DSN                string
	MaxOpenConns       int
	MaxIdleConns       int
	ConnMaxLifetime    time.Duration
	ConnMaxIdleTime    time.Duration
	SlowThreshold      time.Duration
	LoggerLevel        gormlogger.LogLevel
	DisableColor       bool
	IgnoreRecordNotFound bool
}

// Database 数据库包装器
type Database struct {
	db     *gorm.DB
	config *DBConfig
}

// NewDatabase 创建数据库实例
func NewDatabase(config *DBConfig) (*Database, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	gormConfig := &gorm.Config{
		Logger: gormlogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormlogger.Config{
				SlowThreshold:             config.SlowThreshold,
				LogLevel:                  config.LoggerLevel,
				IgnoreRecordNotFoundError: config.IgnoreRecordNotFound,
				Colorful:                  !config.DisableColor,
			},
		),
	}

	var dialector gorm.Dialector
	switch config.Driver {
	case "mysql":
		dialector = mysql.Open(config.DSN)
	case "postgres":
		dialector = postgres.Open(config.DSN)
	default:
		return nil, errors.New(errors.ErrorTypeValidation, "unsupported database driver: "+config.Driver)
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "failed to connect to database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "failed to get underlying database")
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	if config.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	}

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeDatabase, "failed to ping database")
	}

	return &Database{
		db:     db,
		config: config,
	}, nil
}

// Validate 验证配置
func (c *DBConfig) Validate() error {
	if c.Driver == "" {
		return errors.New(errors.ErrorTypeValidation, "database driver is required")
	}
	if c.DSN == "" {
		return errors.New(errors.ErrorTypeValidation, "database DSN is required")
	}
	if c.MaxOpenConns <= 0 {
		return errors.New(errors.ErrorTypeValidation, "MaxOpenConns must be greater than 0")
	}
	if c.MaxIdleConns <= 0 {
		return errors.New(errors.ErrorTypeValidation, "MaxIdleConns must be greater than 0")
	}
	if c.MaxIdleConns > c.MaxOpenConns {
		return errors.New(errors.ErrorTypeValidation, "MaxIdleConns cannot be greater than MaxOpenConns")
	}
	return nil
}

// DB 获取GORM实例
func (d *Database) DB() *gorm.DB {
	return d.db
}

// Config 获取配置
func (d *Database) Config() *DBConfig {
	return d.config
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeDatabase, "failed to get underlying database")
	}
	return sqlDB.Close()
}

// Health 健康检查
func (d *Database) Health(ctx context.Context) error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeDatabase, "failed to get underlying database")
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return errors.Wrap(err, errors.ErrorTypeDatabase, "failed to ping database")
	}

	return nil
}

// Stats 获取连接池统计信息
func (d *Database) Stats() map[string]interface{} {
	sqlDB, err := d.db.DB()
	if err != nil {
		return map[string]interface{}{
			"error": "failed to get underlying database",
		}
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_conns":    stats.MaxOpenConnections,
		"open_conns":        stats.OpenConnections,
		"in_use":            stats.InUse,
		"idle":              stats.Idle,
		"wait_count":        stats.WaitCount,
		"wait_duration":     stats.WaitDuration.String(),
		"max_idle_closed":   stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}
}

// Transaction 执行事务
func (d *Database) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return d.db.WithContext(ctx).Transaction(fn)
}

// Begin 开始事务
func (d *Database) Begin(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx).Begin()
}

// WithContext 创建带上下文的数据库会话
func (d *Database) WithContext(ctx context.Context) *gorm.DB {
	return d.db.WithContext(ctx)
}

