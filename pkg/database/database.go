package database

import (
	"github.com/hedeqiang/skeleton/internal/config"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
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
