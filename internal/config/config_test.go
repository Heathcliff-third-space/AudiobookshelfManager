package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// 保存原始环境变量
	oldToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	oldABToken := os.Getenv("AUDIOBOOKSHELF_TOKEN")
	oldABURL := os.Getenv("AUDIOBOOKSHELF_URL")
	oldProxy := os.Getenv("PROXY_ADDRESS")

	// 确保在测试结束后恢复环境变量
	defer func() {
		os.Setenv("TELEGRAM_BOT_TOKEN", oldToken)
		os.Setenv("AUDIOBOOKSHELF_TOKEN", oldABToken)
		os.Setenv("AUDIOBOOKSHELF_URL", oldABURL)
		os.Setenv("PROXY_ADDRESS", oldProxy)
	}()

	// 设置测试环境变量
	os.Setenv("TELEGRAM_BOT_TOKEN", "test_telegram_token")
	os.Setenv("AUDIOBOOKSHELF_TOKEN", "test_audiobookshelf_token")
	os.Setenv("AUDIOBOOKSHELF_URL", "http://localhost:13378")
	os.Setenv("PROXY_ADDRESS", "127.0.0.1:7890")

	// 加载配置
	cfg := LoadConfig()

	// 验证配置是否正确加载
	if cfg.TelegramBotToken != "test_telegram_token" {
		t.Errorf("期望 TelegramBotToken 为 'test_telegram_token'，实际得到 '%s'", cfg.TelegramBotToken)
	}

	if cfg.AudiobookshelfToken != "test_audiobookshelf_token" {
		t.Errorf("期望 AudiobookshelfToken 为 'test_audiobookshelf_token'，实际得到 '%s'", cfg.AudiobookshelfToken)
	}

	if cfg.AudiobookshelfURL != "http://localhost:13378" {
		t.Errorf("期望 AudiobookshelfURL 为 'http://localhost:13378'，实际得到 '%s'", cfg.AudiobookshelfURL)
	}

	if cfg.ProxyAddress != "127.0.0.1:7890" {
		t.Errorf("期望 ProxyAddress 为 '127.0.0.1:7890'，实际得到 '%s'", cfg.ProxyAddress)
	}

	// 测试默认端口
	os.Unsetenv("AUDIOBOOKSHELF_PORT")
	cfg = LoadConfig()
	if cfg.AudiobookshelfPort != 13378 {
		t.Errorf("期望 AudiobookshelfPort 为 13378，实际得到 %d", cfg.AudiobookshelfPort)
	}
}