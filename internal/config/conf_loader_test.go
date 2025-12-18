package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestConfDirectoryLoading 测试从 conf 目录加载配置
func TestConfDirectoryLoading(t *testing.T) {
	// 保存当前工作目录
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal("无法获取当前工作目录:", err)
	}

	// 创建临时目录结构用于测试
	tempDir := t.TempDir()
	confDir := filepath.Join(tempDir, "conf")
	
	// 创建 conf 目录
	err = os.Mkdir(confDir, 0755)
	if err != nil {
		t.Fatal("无法创建 conf 目录:", err)
	}

	// 创建测试 .env 文件
	envContent := `TEST_KEY=test_value_from_conf
ANOTHER_KEY=another_test_value`
	envFile := filepath.Join(confDir, ".env")
	err = os.WriteFile(envFile, []byte(envContent), 0644)
	if err != nil {
		t.Fatal("无法创建测试 .env 文件:", err)
	}

	// 切换到临时目录
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatal("无法切换到临时目录:", err)
	}

	// 确保在测试结束时恢复原始工作目录
	defer func() {
		os.Chdir(originalWd)
	}()

	// 测试加载配置
	LoadConfig()
	
	// 验证配置是否正确加载
	expectedValue := "test_value_from_conf"
	actualValue := getEnvWithDefault("TEST_KEY", "")
	if actualValue != expectedValue {
		t.Errorf("期望 TEST_KEY 值为 '%s'，实际得到 '%s'", expectedValue, actualValue)
	}

	expectedValue = "another_test_value"
	actualValue = getEnvWithDefault("ANOTHER_KEY", "")
	if actualValue != expectedValue {
		t.Errorf("期望 ANOTHER_KEY 值为 '%s'，实际得到 '%s'", expectedValue, actualValue)
	}
}