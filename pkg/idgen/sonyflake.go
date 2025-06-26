package idgen

import (
	"fmt"
	"net"
	"time"

	"github.com/sony/sonyflake/v2"
)

// IDGenerator ID生成器接口
type IDGenerator interface {
	NextID() (int64, error)
	NextIDString() (string, error)
}

// SonyflakeGenerator 基于Sonyflake的ID生成器
type SonyflakeGenerator struct {
	sf *sonyflake.Sonyflake
}

// NewSonyflakeGenerator 创建新的Sonyflake ID生成器
func NewSonyflakeGenerator() (*SonyflakeGenerator, error) {
	// 获取机器ID（使用本地IP地址的低16位）
	machineID, err := getMachineID()
	if err != nil {
		return nil, fmt.Errorf("failed to get machine ID: %w", err)
	}

	// 配置Sonyflake设置
	settings := sonyflake.Settings{
		StartTime:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), // 起始时间
		MachineID:      func() (int, error) { return machineID, nil },
		CheckMachineID: validateMachineID,
	}

	sf, err := sonyflake.New(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create sonyflake instance: %w", err)
	}

	return &SonyflakeGenerator{sf: sf}, nil
}

// NewSonyflakeGeneratorWithConfig 使用自定义配置创建Sonyflake ID生成器
func NewSonyflakeGeneratorWithConfig(config Config) (*SonyflakeGenerator, error) {
	var machineIDFunc func() (int, error)

	if config.MachineID != 0 {
		// 使用指定的机器ID
		machineIDFunc = func() (int, error) { return config.MachineID, nil }
	} else {
		// 自动获取机器ID
		machineID, err := getMachineID()
		if err != nil {
			return nil, fmt.Errorf("failed to get machine ID: %w", err)
		}
		machineIDFunc = func() (int, error) { return machineID, nil }
	}

	settings := sonyflake.Settings{
		StartTime:      config.StartTime,
		MachineID:      machineIDFunc,
		CheckMachineID: validateMachineID,
	}

	// 如果配置了自定义位数，则使用
	if config.BitsSequence > 0 {
		settings.BitsSequence = config.BitsSequence
	}
	if config.BitsMachineID > 0 {
		settings.BitsMachineID = config.BitsMachineID
	}
	if config.TimeUnit > 0 {
		settings.TimeUnit = config.TimeUnit
	}

	sf, err := sonyflake.New(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create sonyflake instance: %w", err)
	}

	return &SonyflakeGenerator{sf: sf}, nil
}

// NextID 生成下一个唯一ID
func (g *SonyflakeGenerator) NextID() (int64, error) {
	id, err := g.sf.NextID()
	if err != nil {
		return 0, fmt.Errorf("failed to generate ID: %w", err)
	}
	return int64(id), nil
}

// NextIDString 生成下一个唯一ID的字符串形式
func (g *SonyflakeGenerator) NextIDString() (string, error) {
	id, err := g.NextID()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", id), nil
}

// Config ID生成器配置
type Config struct {
	StartTime     time.Time     // 起始时间
	MachineID     int           // 机器ID（0表示自动获取）
	BitsSequence  int           // 序列号位数
	BitsMachineID int           // 机器ID位数
	TimeUnit      time.Duration // 时间单位
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		StartTime:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		MachineID:     0,                     // 自动获取
		BitsSequence:  8,                     // 8位序列号
		BitsMachineID: 16,                    // 16位机器ID
		TimeUnit:      10 * time.Millisecond, // 10毫秒时间单位
	}
}

// getMachineID 获取机器ID（基于本地IP地址）
func getMachineID() (int, error) {
	// 获取本地IP地址
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return 0, fmt.Errorf("failed to get interface addresses: %w", err)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				// 使用IPv4地址的低16位作为机器ID
				ip := ipnet.IP.To4()
				machineID := int(ip[2])<<8 + int(ip[3])
				return machineID, nil
			}
		}
	}

	return 0, fmt.Errorf("no valid IP address found")
}

// validateMachineID 验证机器ID的唯一性
func validateMachineID(machineID int) bool {
	// 简单验证：确保机器ID在有效范围内
	return machineID >= 0 && machineID < (1<<16)
}
