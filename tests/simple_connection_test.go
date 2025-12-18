package tests

import (
	"AudiobookshelfManager/internal/api"
	"AudiobookshelfManager/internal/config"
	"os"
	"testing"
)

// TestEnvVariables 测试环境变量是否正确加载
func TestEnvVariables(t *testing.T) {
	// 加载配置
	cfg := config.LoadConfig()

	// 检查配置是否加载
	if cfg.TelegramBotToken == "" {
		t.Log("TELEGRAM_BOT_TOKEN 未设置")
	} else {
		t.Log("TELEGRAM_BOT_TOKEN 已设置")
	}

	if cfg.AudiobookshelfToken == "" {
		t.Log("AUDIOBOOKSHELF_TOKEN 未设置")
	} else {
		t.Log("AUDIOBOOKSHELF_TOKEN 已设置")
	}

	t.Logf("Audiobookshelf URL: %s", cfg.AudiobookshelfURL)
	t.Logf("Audiobookshelf Port: %d", cfg.AudiobookshelfPort)
	t.Logf("Proxy Address: %s", cfg.ProxyAddress)
}

// TestAudiobookshelfClientCreation 测试 Audiobookshelf 客户端创建
func TestAudiobookshelfClientCreation(t *testing.T) {
	cfg := config.LoadConfig()
	client := api.NewClient(cfg)
	
	if client == nil {
		t.Error("无法创建 Audiobookshelf 客户端")
	} else {
		t.Log("成功创建 Audiobookshelf 客户端")
	}
}

// TestConfigLoading 测试配置加载
func TestConfigLoading(t *testing.T) {
	cfg := config.LoadConfig()
	
	// 检查是否返回了配置对象
	if cfg == nil {
		t.Error("配置加载失败")
	} else {
		t.Log("配置加载成功")
	}
}