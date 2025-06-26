package logger

import (
	"github.com/hedeqiang/skeleton/internal/config"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New 根据提供的配置创建一个新的 zap Logger 实例
func New(cfg *config.Logger) (*zap.Logger, error) {
	// 设置日志级别
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	// 创建 zap core
	core := zapcore.NewCore(
		getEncoder(cfg.Encoding),
		getWriteSyncer(cfg.OutputPath),
		level,
	)

	// 创建 logger
	// zap.AddCaller() 会显示调用者信息
	// zap.AddCallerSkip(1) 可以跳过封装函数的调用栈，直接显示业务代码的位置
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return logger, nil
}

// getEncoder 根据配置返回不同的编码器
func getEncoder(encoding string) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	// 人性化时间格式
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 小写字母级别
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if encoding == "json" {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	// 默认为 console 格式
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// getWriteSyncer 根据配置返回不同的写入器，支持多重输出
func getWriteSyncer(outputPaths []string) zapcore.WriteSyncer {
	var writers []zapcore.WriteSyncer

	for _, path := range outputPaths {
		if path == "stdout" {
			writers = append(writers, zapcore.AddSync(os.Stdout))
		} else {
			// 如果是文件路径，可以添加对文件写入的支持
			// 为保持示例简洁，此处暂不实现文件日志轮转等复杂功能
			file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				writers = append(writers, zapcore.AddSync(file))
			}
		}
	}
	return zapcore.NewMultiWriteSyncer(writers...)
}
