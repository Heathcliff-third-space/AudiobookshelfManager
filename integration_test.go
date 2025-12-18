package main

import (
	"github.com/Heathcliff-third-space/AudiobookshelfManager/internal/api"
	"github.com/Heathcliff-third-space/AudiobookshelfManager/internal/config"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TestFullIntegration 测试整个系统的集成连接
func TestFullIntegration(t *testing.T) {
	// 检查是否设置了跳过集成测试的环境变量
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("跳过集成测试，因为设置了 SKIP_INTEGRATION_TESTS=true")
	}

	// 加载配置
	cfg := config.LoadConfig()

	t.Run("配置加载测试", func(t *testing.T) {
		if cfg == nil {
			t.Fatal("无法加载配置")
		}

		if cfg.AudiobookshelfPort <= 0 {
			t.Error("Audiobookshelf 端口未正确设置")
		}

		t.Logf("配置加载成功: %+v", cfg)
	})

	t.Run("Audiobookshelf 连接测试", func(t *testing.T) {
		if cfg.AudiobookshelfToken == "" {
			t.Skip("跳过测试: 未设置 AUDIOBOOKSHELF_TOKEN")
		}

		client := api.NewClient(cfg)
		_, err := client.GetLibraries()
		if err != nil {
			t.Errorf("无法连接到 Audiobookshelf 服务器: %v", err)
		} else {
			t.Log("✅ Audiobookshelf 连接成功")
		}
	})

	t.Run("Telegram Bot 连接测试", func(t *testing.T) {
		if cfg.TelegramBotToken == "" {
			t.Skip("跳过测试: 未设置 TELEGRAM_BOT_TOKEN")
		}

		var bot *tgbotapi.BotAPI
		var err error

		// 如果设置了代理，则通过代理连接 Telegram
		if cfg.ProxyAddress != "" {
			t.Logf("使用代理连接 Telegram: %s", cfg.ProxyAddress)
			proxyURL, pErr := url.Parse("http://" + cfg.ProxyAddress)
			if pErr != nil {
				t.Fatalf("无效的代理地址: %v", pErr)
			}

			proxyClient := &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				},
				Timeout: 30 * time.Second,
			}

			bot, err = tgbotapi.NewBotAPIWithClient(cfg.TelegramBotToken, tgbotapi.APIEndpoint, proxyClient)
		} else {
			bot, err = tgbotapi.NewBotAPI(cfg.TelegramBotToken)
		}

		if err != nil {
			t.Errorf("无法连接到 Telegram Bot API: %v", err)
		} else {
			_, err := bot.GetMe()
			if err != nil {
				t.Errorf("无法获取 Telegram Bot 信息: %v", err)
			} else {
				t.Log("✅ Telegram Bot 连接成功")
			}
		}
	})
}