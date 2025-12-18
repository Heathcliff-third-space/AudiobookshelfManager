package bot

import (
	"context"
	"os"
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TestTelegramBotConnection 测试 Telegram Bot 连接
func TestTelegramBotConnection(t *testing.T) {
	// 检查是否设置了跳过测试的环境变量
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("跳过集成测试，因为设置了 SKIP_INTEGRATION_TESTS=true")
	}

	// 设置测试超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 创建一个通道来接收测试结果
	done := make(chan bool, 1)
	var testResult struct {
		bot *tgbotapi.BotAPI
		err error
	}

	// 在 goroutine 中运行测试
	go func() {
		// 获取 Telegram Bot Token
		token := os.Getenv("TELEGRAM_BOT_TOKEN")
		if token == "" {
			t.Log("跳过测试: 未设置 TELEGRAM_BOT_TOKEN")
			done <- true
			return
		}

		// 创建 Telegram Bot 客户端
		t.Log("正在测试 Telegram Bot 连接...")
		testResult.bot, testResult.err = tgbotapi.NewBotAPI(token)
		done <- true
	}()

	// 等待测试完成或超时
	select {
	case <-done:
		// 测试完成
		if testResult.err != nil {
			t.Errorf("无法连接到 Telegram Bot API: %v", testResult.err)
		} else {
			me, err := testResult.bot.GetMe()
			if err != nil {
				t.Errorf("无法获取 Telegram Bot 信息: %v", err)
			} else {
				t.Logf("✅ 成功连接到 Telegram Bot: @%s", me.UserName)
			}
		}
	case <-ctx.Done():
		// 测试超时
		t.Fatal("测试超时")
	}
}