package config

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// C 是一个全局变量，持有所有应用配置
var C *Config

// Config 是整个应用的配置结构体
type Config struct {
	App       App                 `mapstructure:"app"`
	Logger    Logger              `mapstructure:"logger"`
	Databases map[string]Database `mapstructure:"databases"`
	Redis     Redis               `mapstructure:"redis"`
	RabbitMQ  RabbitMQ            `mapstructure:"rabbitmq"`
	Scheduler SchedulerConfig     `mapstructure:"scheduler"`
	Trace     Trace               `mapstructure:"trace"`
	JWT       JWT                 `mapstructure:"jwt"`
}

// App 应用配置
type App struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// Logger 日志配置
type Logger struct {
	Level      string   `mapstructure:"level"`
	Encoding   string   `mapstructure:"encoding"`
	OutputPath []string `mapstructure:"output_path"`
}

// Database 单个数据源的配置
type Database struct {
	Type            string        `mapstructure:"type"`
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// Redis 配置
type Redis struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// RabbitMQ 配置
type RabbitMQ struct {
	URL       string           `mapstructure:"url"`
	Exchanges []ExchangeConfig `mapstructure:"exchanges"`
	Queues    []QueueConfig    `mapstructure:"queues"`
}

// ExchangeConfig 交换机配置
type ExchangeConfig struct {
	Name       string `mapstructure:"name"`
	Type       string `mapstructure:"type"`
	Durable    bool   `mapstructure:"durable"`
	AutoDelete bool   `mapstructure:"auto_delete"`
}

// QueueConfig 队列配置
type QueueConfig struct {
	Name        string   `mapstructure:"name"`
	Durable     bool     `mapstructure:"durable"`
	AutoDelete  bool     `mapstructure:"auto_delete"`
	Exclusive   bool     `mapstructure:"exclusive"`
	Exchange    string   `mapstructure:"exchange"`
	RoutingKeys []string `mapstructure:"routing_keys"`
}

// SchedulerConfig 计划任务配置
type SchedulerConfig struct {
	Enabled bool                 `mapstructure:"enabled"`
	Jobs    []SchedulerJobConfig `mapstructure:"jobs"`
}

// SchedulerJobConfig 计划任务配置
type SchedulerJobConfig struct {
	Name        string `mapstructure:"name"`
	Type        string `mapstructure:"type"`     // duration, cron, daily, weekly, monthly
	Schedule    string `mapstructure:"schedule"` // 调度表达式
	Enabled     bool   `mapstructure:"enabled"`
	Description string `mapstructure:"description"`
}

// Trace Tracing 配置
type Trace struct {
	Enabled      bool    `mapstructure:"enabled"`
	Endpoint     string  `mapstructure:"endpoint"`
	SamplerType  string  `mapstructure:"sampler_type"`
	SamplerParam float64 `mapstructure:"sampler_param"`
}

// JWT 认证配置
type JWT struct {
	Secret         string        `mapstructure:"secret"`
	ExpireDuration time.Duration `mapstructure:"expire_duration"`
}

// LoadConfig 加载配置并返回 Config 实例
func LoadConfig() (*Config, error) {
	v := viper.New()

	// 开启环境变量支持
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 从环境变量获取配置文件路径，默认使用 config.dev.yaml
	configFile := "configs/config.dev.yaml"
	if envConfigFile := v.GetString("CONFIG_FILE"); envConfigFile != "" {
		configFile = envConfigFile
	}

	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Load 从指定路径加载配置
// 优先级: 环境变量 > 配置文件
func Load(configPath string) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}

	// 开启环境变量支持
	// APP_PORT=8081 将会覆盖文件中的 app.port
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 将配置 unmarshal 到全局变量 C
	if err := v.Unmarshal(&C); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	log.Println("Configuration loaded successfully.")
}
