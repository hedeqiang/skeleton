package config

// Config 配置接口
type Config interface {
	GetApp() AppConfig
	GetLogger() LoggerConfig
	GetDatabase() DatabaseConfig
	GetRedis() RedisConfig
	GetRabbitMQ() RabbitMQConfig
	GetJWT() JWTConfig
	GetTrace() TraceConfig
	GetScheduler() SchedulerConfig
}

// AppConfig 应用配置
type AppConfig struct {
	Name string `yaml:"name" json:"name"`
	Env  string `yaml:"env" json:"env"`
	Host string `yaml:"host" json:"host"`
	Port int    `yaml:"port" json:"port"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string   `yaml:"level" json:"level"`
	Encoding   string   `yaml:"encoding" json:"encoding"`
	OutputPath []string `yaml:"output_path" json:"output_path"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Primary    Database `yaml:"primary" json:"primary"`
	Replica    Database `yaml:"replica" json:"replica"`
	Enabled    bool     `yaml:"enabled" json:"enabled"`
}

// Database 单个数据库配置
type Database struct {
	Type             string        `yaml:"type" json:"type"`
	DSN              string        `yaml:"dsn" json:"dsn"`
	MaxOpenConns     int           `yaml:"max_open_conns" json:"max_open_conns"`
	MaxIdleConns     int           `yaml:"max_idle_conns" json:"max_idle_conns"`
	ConnMaxLifetime  string        `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
	ConnMaxIdleTime  string        `yaml:"conn_max_idle_time" json:"conn_max_idle_time"`
	SlowThreshold    string        `yaml:"slow_threshold" json:"slow_threshold"`
	LoggerLevel      string        `yaml:"logger_level" json:"logger_level"`
	DisableColor     bool          `yaml:"disable_color" json:"disable_color"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string `yaml:"addr" json:"addr"`
	Password string `yaml:"password" json:"password"`
	DB       int    `yaml:"db" json:"db"`
	Enabled  bool   `yaml:"enabled" json:"enabled"`
}

// RabbitMQConfig RabbitMQ配置
type RabbitMQConfig struct {
	URL       string       `yaml:"url" json:"url"`
	Exchanges []Exchange   `yaml:"exchanges" json:"exchanges"`
	Queues    []Queue      `yaml:"queues" json:"queues"`
	Enabled   bool         `yaml:"enabled" json:"enabled"`
}

// Exchange 交换机配置
type Exchange struct {
	Name       string `yaml:"name" json:"name"`
	Type       string `yaml:"type" json:"type"`
	Durable    bool   `yaml:"durable" json:"durable"`
	AutoDelete bool   `yaml:"auto_delete" json:"auto_delete"`
}

// Queue 队列配置
type Queue struct {
	Name         string   `yaml:"name" json:"name"`
	Durable      bool     `yaml:"durable" json:"durable"`
	AutoDelete   bool     `yaml:"auto_delete" json:"auto_delete"`
	Exclusive    bool     `yaml:"exclusive" json:"exclusive"`
	Exchange     string   `yaml:"exchange" json:"exchange"`
	RoutingKeys  []string `yaml:"routing_keys" json:"routing_keys"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret           string `yaml:"secret" json:"secret"`
	ExpireDuration   string `yaml:"expire_duration" json:"expire_duration"`
	RefreshDuration  string `yaml:"refresh_duration" json:"refresh_duration"`
	Enabled          bool   `yaml:"enabled" json:"enabled"`
}

// TraceConfig 追踪配置
type TraceConfig struct {
	Enabled      bool   `yaml:"enabled" json:"enabled"`
	Endpoint     string `yaml:"endpoint" json:"endpoint"`
	SamplerType  string `yaml:"sampler_type" json:"sampler_type"`
	SamplerParam int    `yaml:"sampler_param" json:"sampler_param"`
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	Enabled bool           `yaml:"enabled" json:"enabled"`
	Jobs    []ScheduledJob `yaml:"jobs" json:"jobs"`
}

// ScheduledJob 计划任务
type ScheduledJob struct {
	Name        string `yaml:"name" json:"name"`
	Type        string `yaml:"type" json:"type"`
	Schedule    string `yaml:"schedule" json:"schedule"`
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	Description string `yaml:"description" json:"description"`
}