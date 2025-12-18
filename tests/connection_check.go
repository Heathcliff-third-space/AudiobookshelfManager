package main

import (
	"AudiobookshelfManager/internal/api"
	"AudiobookshelfManager/internal/config"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	fmt.Println("开始连接测试...")

	// 加载配置
	cfg := config.LoadConfig()

	// 测试 Audiobookshelf 连接
	fmt.Println("\n1. 测试 Audiobookshelf 连接...")
	testAudiobookshelf(cfg)

	// 测试 Telegram 连接
	fmt.Println("\n2. 测试 Telegram 连接...")
	testTelegram(cfg)

	fmt.Println("\n连接测试完成!")
}

func testAudiobookshelf(cfg *config.Config) {
	if cfg.AudiobookshelfToken == "" {
		fmt.Println("  跳过测试: 未设置 AUDIOBOOKSHELF_TOKEN")
		return
	}

	client := api.NewClient(cfg)
	_, err := client.GetLibraries()
	if err != nil {
		fmt.Printf("  ❌ 无法连接到 Audiobookshelf: %v\n", err)
	} else {
		fmt.Println("  ✅ 成功连接到 Audiobookshelf")
	}
}

func testTelegram(cfg *config.Config) {
	if cfg.TelegramBotToken == "" {
		fmt.Println("  跳过测试: 未设置 TELEGRAM_BOT_TOKEN")
		return
	}

	var bot *tgbotapi.BotAPI
	var err error

	// 如果设置了代理，则通过代理连接 Telegram
	if cfg.ProxyAddress != "" {
		fmt.Printf("  使用代理连接 Telegram: %s\n", cfg.ProxyAddress)
		proxyURL, err := url.Parse("http://" + cfg.ProxyAddress)
		if err != nil {
			fmt.Printf("  ❌ 无效的代理地址: %v\n", err)
			return
		}

		proxyClient := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
			Timeout: 30 * time.Second,
		}

		bot, err = tgbotapi.NewBotAPIWithClient(cfg.TelegramBotToken, tgbotapi.APIEndpoint, proxyClient)
		if err != nil {
			fmt.Printf("  ❌ 无法通过代理连接到 Telegram Bot API: %v\n", err)
			return
		}
	} else {
		bot, err = tgbotapi.NewBotAPI(cfg.TelegramBotToken)
		if err != nil {
			fmt.Printf("  ❌ 无法连接到 Telegram Bot API: %v\n", err)
			return
		}
	}

	// 获取 Bot 信息
	botInfo, err := bot.GetMe()
	if err != nil {
		fmt.Printf("  ❌ 无法获取 Telegram Bot 信息: %v\n", err)
	} else {
		fmt.Printf("  ✅ 成功连接到 Telegram Bot API，用户名: @%s\n", botInfo.UserName)
	}
}