package config

import (
	"os"
	"testing"
)

// TestEnvFile 测试环境变量是否正确加载
func TestEnvFile(t *testing.T) {
	// 保存原始环境变量
	oldCwd, _ := os.Getwd()
	
	// 确保环境变量没有被之前的测试影响
	os.Clearenv()

	// 加载配置（会自动加载.env文件）
	cfg := LoadConfig()

	// 检查配置是否加载成功
	if cfg == nil {
		t.Fatal("无法加载配置")
	}

	// 检查必需的环境变量是否存在
	if cfg.TelegramBotToken == "" {
		t.Error("TELEGRAM_BOT_TOKEN 未设置或为空")
	} else {
		t.Log("TELEGRAM_BOT_TOKEN 已正确设置")
	}

	if cfg.AudiobookshelfToken == "" {
		t.Error("AUDIOBOOKSHELF_TOKEN 未设置或为空")
	} else {
		t.Log("AUDIOBOOKSHELF_TOKEN 已正确设置")
	}

	// 检查可选的环境变量
	if cfg.AudiobookshelfURL != "" {
		t.Logf("AUDIOBOOKSHELF_URL 设置为: %s", cfg.AudiobookshelfURL)
	} else {
		t.Log("AUDIOBOOKSHELF_URL 使用默认值")
	}

	if cfg.ProxyAddress != "" {
		t.Logf("PROXY_ADDRESS 设置为: %s", cfg.ProxyAddress)
	}

	t.Logf("Audiobookshelf 端口: %d", cfg.AudiobookshelfPort)
	t.Logf("Debug 模式: %t", cfg.Debug)
	
	// 恢复原始工作目录
	os.Chdir(oldCwd)
}