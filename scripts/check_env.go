package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("检查环境变量配置...")

	// 加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		fmt.Println("警告: 未找到 .env 文件，将使用系统环境变量")
	}

	// 检查关键环境变量
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	audiobookshelfToken := os.Getenv("AUDIOBOOKSHELF_TOKEN")

	fmt.Println("\n环境变量状态:")
	fmt.Printf("TELEGRAM_BOT_TOKEN: %s\n", formatStatus(telegramToken))
	fmt.Printf("AUDIOBOOKSHELF_TOKEN: %s\n", formatStatus(audiobookshelfToken))

	// 显示其他配置
	audiobookshelfURL := os.Getenv("AUDIOBOOKSHELF_URL")
	proxyAddress := os.Getenv("PROXY_ADDRESS")
	debug := os.Getenv("DEBUG")

	fmt.Printf("AUDIOBOOKSHELF_URL: %s\n", audiobookshelfURL)
	fmt.Printf("PROXY_ADDRESS: %s\n", proxyAddress)
	fmt.Printf("DEBUG: %s\n", debug)

	fmt.Println("\n检查完成!")
}

func formatStatus(value string) string {
	if value == "" {
		return "❌ 未设置"
	}
	return "✅ 已设置"
}