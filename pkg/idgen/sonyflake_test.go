package idgen

import (
	"fmt"
	"testing"
	"time"
)

func TestSonyflakeGenerator(t *testing.T) {
	// 创建默认ID生成器
	generator, err := NewSonyflakeGenerator()
	if err != nil {
		t.Fatalf("Failed to create ID generator: %v", err)
	}

	// 测试生成ID
	ids := make(map[int64]bool)
	for i := 0; i < 1000; i++ {
		id, err := generator.NextID()
		if err != nil {
			t.Fatalf("Failed to generate ID: %v", err)
		}

		// 检查ID是否唯一
		if ids[id] {
			t.Fatalf("Duplicate ID generated: %d", id)
		}
		ids[id] = true

		// 检查ID是否为正数
		if id <= 0 {
			t.Fatalf("Invalid ID generated: %d", id)
		}
	}

	t.Logf("Successfully generated %d unique IDs", len(ids))
}

func TestSonyflakeGeneratorWithConfig(t *testing.T) {
	// 创建自定义配置
	config := Config{
		StartTime:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		MachineID:     1001, // 指定机器ID
		BitsSequence:  8,
		BitsMachineID: 16,
		TimeUnit:      10 * time.Millisecond,
	}

	generator, err := NewSonyflakeGeneratorWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create ID generator with config: %v", err)
	}

	// 测试生成ID
	id, err := generator.NextID()
	if err != nil {
		t.Fatalf("Failed to generate ID: %v", err)
	}

	if id <= 0 {
		t.Fatalf("Invalid ID generated: %d", id)
	}

	t.Logf("Generated ID with custom config: %d", id)
}

func TestNextIDString(t *testing.T) {
	generator, err := NewSonyflakeGenerator()
	if err != nil {
		t.Fatalf("Failed to create ID generator: %v", err)
	}

	idStr, err := generator.NextIDString()
	if err != nil {
		t.Fatalf("Failed to generate ID string: %v", err)
	}

	if idStr == "" {
		t.Fatalf("Empty ID string generated")
	}

	t.Logf("Generated ID string: %s", idStr)
}

func ExampleNewSonyflakeGenerator() {
	// 创建默认的Sonyflake ID生成器
	generator, err := NewSonyflakeGenerator()
	if err != nil {
		panic(err)
	}

	// 生成唯一ID
	id, err := generator.NextID()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Generated ID: %d\n", id)

	// 生成字符串形式的ID
	idStr, err := generator.NextIDString()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Generated ID string: %s\n", idStr)
}

func ExampleNewSonyflakeGeneratorWithConfig() {
	// 创建自定义配置
	config := Config{
		StartTime:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		MachineID:     1001,                  // 指定机器ID
		BitsSequence:  8,                     // 8位序列号
		BitsMachineID: 16,                    // 16位机器ID
		TimeUnit:      10 * time.Millisecond, // 10毫秒时间单位
	}

	// 使用自定义配置创建生成器
	generator, err := NewSonyflakeGeneratorWithConfig(config)
	if err != nil {
		panic(err)
	}

	// 生成ID
	id, err := generator.NextID()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Generated ID with custom config: %d\n", id)
}

func BenchmarkSonyflakeGenerator_NextID(b *testing.B) {
	generator, err := NewSonyflakeGenerator()
	if err != nil {
		b.Fatalf("Failed to create ID generator: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.NextID()
		if err != nil {
			b.Fatalf("Failed to generate ID: %v", err)
		}
	}
}
