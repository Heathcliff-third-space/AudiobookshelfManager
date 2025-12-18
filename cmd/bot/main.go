package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/Heathcliff-third-space/AudiobookshelfManager/internal/api"
	bot_pkg "github.com/Heathcliff-third-space/AudiobookshelfManager/internal/bot"
	"github.com/Heathcliff-third-space/AudiobookshelfManager/internal/config"
	"github.com/Heathcliff-third-space/AudiobookshelfManager/internal/models"
	"github.com/Heathcliff-third-space/AudiobookshelfManager/internal/services"
)

var allowedUserIDs map[int64]bool

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化允许的用户ID映射
	allowedUserIDs = make(map[int64]bool)
	for _, id := range cfg.AllowedUserIDs {
		allowedUserIDs[id] = true
	}
	log.Printf("允许访问的用户ID: %v", cfg.AllowedUserIDs)

	// 检查必要配置
	if cfg.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN 环境变量未设置")
	}

	// 初始化 Telegram Bot
	var telegramBot *tgbotapi.BotAPI
	var err error

	// 如果设置了代理，则通过代理连接 Telegram
	if cfg.ProxyAddress != "" {
		log.Printf("使用代理连接 Telegram: %s", cfg.ProxyAddress)
		proxyURL, err := url.Parse("http://" + cfg.ProxyAddress)
		if err != nil {
			log.Fatal("无效的代理地址:", err)
		}

		proxyClient := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
			Timeout: 30 * time.Second,
		}

		telegramBot, err = tgbotapi.NewBotAPIWithClient(cfg.TelegramBotToken, tgbotapi.APIEndpoint, proxyClient)
		if err != nil {
			log.Fatal("无法通过代理连接到 Telegram Bot API:", err)
		}
	} else {
		telegramBot, err = tgbotapi.NewBotAPI(cfg.TelegramBotToken)
		if err != nil {
			log.Fatal("无法连接到 Telegram Bot API:", err)
		}
	}

	if cfg.Debug {
		telegramBot.Debug = true
	}

	log.Printf("已授权账户 %s", telegramBot.Self.UserName)

	// 初始化 Audiobookshelf API 客户端 (不使用代理)
	audiobookshelfClient := api.NewClient(cfg)

	// 初始化服务器信息服务
	serverService := services.NewServerService(audiobookshelfClient)

	// 测试连接
	_, err = audiobookshelfClient.GetLibraries()
	if err != nil {
		log.Printf("警告：无法连接到 Audiobookshelf API: %v", err)
	} else {
		log.Println("成功连接到 Audiobookshelf API")
	}

	// 注册菜单命令
	err = bot_pkg.RegisterCommands(telegramBot)
	if err != nil {
		log.Printf("注册命令失败: %v", err)
	} else {
		log.Println("成功注册 Telegram 命令")
	}

	// 设置更新配置
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := telegramBot.GetUpdatesChan(u)

	// 处理中断信号以便优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 同时处理来自 Telegram 的更新和系统信号
	for {
		select {
		case update := <-updates:
			if update.Message != nil { // 如果我们收到一条消息
				if !isUserAllowed(update.Message.From.ID) {
					log.Printf("拒绝用户 %s (ID: %d) 的访问", update.Message.From.UserName, update.Message.From.ID)
					sendAccessDeniedMessage(telegramBot, update.Message.Chat.ID)
					continue
				}
				handleMessage(telegramBot, update.Message, serverService)
			} else if update.CallbackQuery != nil { // 如果我们收到一个回调查询（按钮点击）
				if !isUserAllowed(update.CallbackQuery.From.ID) {
					log.Printf("拒绝用户 %s (ID: %d) 的访问", update.CallbackQuery.From.UserName, update.CallbackQuery.From.ID)
					sendAccessDeniedMessage(telegramBot, update.CallbackQuery.Message.Chat.ID)
					// 响应回调查询，避免按钮loading状态持续太久
					callbackResp := tgbotapi.NewCallback(update.CallbackQuery.ID, "访问被拒绝")
					telegramBot.Send(callbackResp)
					continue
				}
				handleCallbackQuery(telegramBot, update.CallbackQuery, serverService)
			}

		case <-sigChan:
			log.Println("接收到中断信号，正在关闭...")
			return
		}
	}
}

// isUserAllowed 检查用户是否有权限使用机器人
func isUserAllowed(userID int64) bool {
	// 如果没有设置允许的用户ID，则允许所有用户访问（向后兼容）
	if len(allowedUserIDs) == 0 {
		return true
	}
	
	// 检查用户是否在允许列表中
	return allowedUserIDs[userID]
}

// sendAccessDeniedMessage 发送访问拒绝消息
func sendAccessDeniedMessage(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "🚫 抱歉，您没有权限使用此机器人。")
	bot.Send(msg)
}

// handleMessage 处理消息
func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message, serverService *services.ServerService) {
	log.Printf("[%s] %s", message.From.UserName, message.Text)

	// 只响应特定用户的私聊消息（可选安全措施）
	if message.Chat.Type != "private" {
		return
	}

	switch strings.ToLower(message.Text) {
	case "/start", "/help":
		sendMainMenu(bot, message.Chat.ID, 0)
	case "/serverinfo":
		sendServerInfo(bot, message.Chat.ID, 0, serverService)
	case "/users":
		sendUsersInfo(bot, message.Chat.ID, 0, serverService)
	case "/search":
		promptForSearchTerm(bot, message.Chat.ID, 0)
	case "/libraries":
		sendLibrariesList(bot, message.Chat.ID, 0, serverService)
	case "/mystats":
		sendMyStats(bot, message.Chat.ID, 0, serverService)
	default:
		// 检查是否是搜索查询
		log.Printf("检查是否是搜索查询: ReplyToMessage=%v, Text=%s", message.ReplyToMessage, message.Text)
		if message.ReplyToMessage != nil {
			log.Printf("ReplyToMessage Text: %s", message.ReplyToMessage.Text)
			if strings.Contains(message.ReplyToMessage.Text, "请输入您要搜索的图书名称") {
				log.Printf("识别为搜索请求: %s", message.Text)
				performBookSearch(bot, message.Chat.ID, message.Text, serverService)
				return
			}
		}

		// 检查是否是搜索关键词（不依赖ReplyToMessage）
		// 如果用户刚刚点击了搜索按钮，我们就认为下一条消息是搜索词
		performBookSearch(bot, message.Chat.ID, message.Text, serverService)
		return
	}
}

// handleCallbackQuery 处理回调查询（按钮点击）
func handleCallbackQuery(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery, serverService *services.ServerService) {
	// 响应回调查询，避免按钮loading状态持续太久
	callbackResp := tgbotapi.NewCallback(callback.ID, "")
	bot.Send(callbackResp)
	
	switch callback.Data {
	case "main_menu":
		editMainMenu(bot, callback.Message.Chat.ID, callback.Message.MessageID)
	case "system_info":
		// 显示加载状态
		edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, "📊 正在获取服务器信息，请稍候...")
		bot.Send(edit)
		// 执行实际操作
		editServerInfo(bot, callback.Message.Chat.ID, callback.Message.MessageID, serverService)
	case "search_books":
		promptForSearchTerm(bot, callback.Message.Chat.ID, callback.Message.MessageID)
	case "users_list":
		// 显示加载状态
		edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, "👥 正在获取用户信息，请稍候...")
		bot.Send(edit)
		// 执行实际操作
		sendUsersInfo(bot, callback.Message.Chat.ID, callback.Message.MessageID, serverService)
	case "my_stats":
		// 显示加载状态
		edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, "📈 正在获取个人统计信息，请稍候...")
		bot.Send(edit)
		// 执行实际操作
		sendMyStats(bot, callback.Message.Chat.ID, callback.Message.MessageID, serverService)
	case "libraries_list":
		// 显示加载状态
		edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, "📚 正在获取媒体库信息，请稍候...")
		bot.Send(edit)
		// 执行实际操作
		sendLibrariesList(bot, callback.Message.Chat.ID, callback.Message.MessageID, serverService)
	case "help":
		editHelpMessage(bot, callback.Message.Chat.ID, callback.Message.MessageID)
	}
}

// sendMainMenu 发送主菜单
func sendMainMenu(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	msg := tgbotapi.NewMessage(chatID, "🎧 *欢迎使用 Audiobookshelf 管理机器人*\n\n请选择您要执行的操作:")
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = bot_pkg.CreateMainMenu()
	bot.Send(msg)
}

// editMainMenu 编辑主菜单
func editMainMenu(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, "🎧 *欢迎使用 Audiobookshelf 管理机器人*\n\n请选择您要执行的操作:")
	edit.ParseMode = "Markdown"
	menu := bot_pkg.CreateMainMenu()
	edit.ReplyMarkup = &menu
	bot.Send(edit)
}

// sendServerInfo 发送服务器信息
func sendServerInfo(bot *tgbotapi.BotAPI, chatID int64, messageID int, serverService *services.ServerService) {
	info, err := serverService.GetFormattedServerInfo()
	if err != nil {
		if messageID > 0 {
			editMessage(bot, chatID, messageID, "❌ 获取服务器信息失败: "+err.Error())
		} else {
			sendMessage(bot, chatID, "❌ 获取服务器信息失败: "+err.Error())
		}
		return
	}

	if messageID > 0 {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, info)
		edit.ParseMode = "Markdown"
		menu := bot_pkg.CreateServerInfoMenu()
		edit.ReplyMarkup = &menu
		bot.Send(edit)
	} else {
		msg := tgbotapi.NewMessage(chatID, info)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = bot_pkg.CreateServerInfoMenu()
		bot.Send(msg)
	}
}

// editServerInfo 编辑服务器信息
func editServerInfo(bot *tgbotapi.BotAPI, chatID int64, messageID int, serverService *services.ServerService) {
	sendServerInfo(bot, chatID, messageID, serverService)
}

// sendLibrariesList 发送媒体库列表
func sendLibrariesList(bot *tgbotapi.BotAPI, chatID int64, messageID int, serverService *services.ServerService) {
	libraries, err := serverService.GetLibrariesWithStats()
	if err != nil {
		if messageID > 0 {
			editMessage(bot, chatID, messageID, "❌ 获取媒体库列表失败: "+err.Error())
		} else {
			sendMessage(bot, chatID, "❌ 获取媒体库列表失败: "+err.Error())
		}
		return
	}

	var text string
	if len(libraries) == 0 {
		text = "📭 没有找到媒体库"
	} else {
		text = "📚 *媒体库列表*:\n\n"
		for _, lib := range libraries {
			text += fmt.Sprintf("📖 %s\n", lib.Name)
		}
	}

	if messageID > 0 {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		edit.ParseMode = "Markdown"
		menu := bot_pkg.CreateLibrariesMenu()
		edit.ReplyMarkup = &menu
		bot.Send(edit)
	} else {
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = bot_pkg.CreateLibrariesMenu()
		bot.Send(msg)
	}
}

// editLibrariesList 编辑媒体库列表
func editLibrariesList(bot *tgbotapi.BotAPI, chatID int64, messageID int, serverService *services.ServerService) {
	sendLibrariesList(bot, chatID, messageID, serverService)
}

// promptForSearchTerm 提示用户输入搜索词
func promptForSearchTerm(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	// 如果已经有消息ID，则编辑现有消息
	if messageID > 0 {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, "🔍 请输入您要搜索的图书名称、作者或其他关键词：")
		menu := bot_pkg.CreateSearchMenu()
		edit.ReplyMarkup = &menu
		bot.Send(edit)
	} else {
		// 否则发送新消息
		msg := tgbotapi.NewMessage(chatID, "🔍 请输入您要搜索的图书名称、作者或其他关键词：")
		menu := bot_pkg.CreateSearchMenu()
		msg.ReplyMarkup = &menu
		bot.Send(msg)
	}
}

// performBookSearch 执行图书搜索
func performBookSearch(bot *tgbotapi.BotAPI, chatID int64, searchTerm string, serverService *services.ServerService) {
	// 添加调试日志
	log.Printf("执行图书搜索: %s", searchTerm)

	// 调用搜索服务时不指定特定的媒体库，让服务自行处理所有媒体库的搜索
	books, err := serverService.SearchBooks(searchTerm, "")
	if err != nil {
		log.Printf("搜索出错: %v", err)
		response := fmt.Sprintf("❌ 搜索出错: %v", err)
		msg := tgbotapi.NewMessage(chatID, response)
		msg.ReplyMarkup = bot_pkg.CreateMainMenu()
		bot.Send(msg)
		return
	}

	// 格式化搜索结果
	response := formatSearchResults(searchTerm, books, serverService)

	// 发送或编辑消息
	// 这里我们假设之前的提示消息是通过promptForSearchTerm函数发送的，
	// 并且我们可以通过某种方式获取到该消息的ID
	// 由于当前实现没有保存消息ID，我们需要重新设计
	msg := tgbotapi.NewMessage(chatID, response)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = bot_pkg.CreateMainMenu()
	bot.Send(msg)
}

// formatSearchResults 格式化搜索结果
func formatSearchResults(searchTerm string, books []models.Book, serverService *services.ServerService) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔎 搜索 \"%s\" 的结果:\n\n", searchTerm))

	if len(books) == 0 {
		sb.WriteString("未找到相关书籍。\n")
		return sb.String()
	}

	sb.WriteString("*📚 找到的书籍:*\n")
	for i, book := range books {
		if i >= 10 { // 限制显示前10个结果
			sb.WriteString(fmt.Sprintf("\n+ 还有 %d 本更多书籍...", len(books)-10))
			break
		}
		// 获取媒体库名称
		libraryName, err := serverService.GetLibraryName(book.LibraryID)
		if err != nil {
			libraryName = "未知媒体库"
		}
		var sizeUnit = services.FormatBytes(book.Size)
		// 格式化添加时间 - 正确处理毫秒级时间戳
		addedTime := time.Unix(book.AddedAt/1000, 0).Format("2006-01-02 15:04:05")
		// 添加书籍信息
		sb.WriteString(fmt.Sprintf("• **%s**\n  📁 媒体库: %s\n  💾 大小: %s\n  ⏳ 添加时间: %s\n\n",
			book.RelPath,
			libraryName,
			sizeUnit,
			addedTime))
	}

	return sb.String()
}

// sendUsersInfo 发送用户信息
func sendUsersInfo(bot *tgbotapi.BotAPI, chatID int64, messageID int, serverService *services.ServerService) {
	users, err := serverService.GetUsersWithProgress()
	if err != nil {
		if messageID > 0 {
			editMessage(bot, chatID, messageID, "❌ 获取用户信息失败: "+err.Error())
		} else {
			sendMessage(bot, chatID, "❌ 获取用户信息失败: "+err.Error())
		}
		return
	}

	var text string
	if len(users) == 0 {
		text = "📭 没有找到用户"
	} else {
		text = "*👥 用户信息:*\n\n"
		for _, user := range users {
			// 格式化创建时间
			createdAt := "未知"
			if user.CreatedAt > 0 {
				// createdAt 是毫秒时间戳
				createdTime := time.Unix(user.CreatedAt/1000, 0).Format("2006-01-02 15:04:05")
				createdAt = createdTime
			}

			// 格式化最后在线时间
			lastSeen := "从未登录"
			if user.LastSeen > 0 {
				// lastSeen 是毫秒时间戳
				lastSeenTime := time.Unix(user.LastSeen/1000, 0).Format("2006-01-02 15:04:05")
				lastSeen = lastSeenTime
			}

			// 计算播放进度数量
			progressCount := len(user.MediaProgress)

			activeStatus := "❌ 非活跃"
			if user.IsActive {
				activeStatus = "✅ 活跃"
			}

			userType := "👤 普通用户"
			// 根据Audiobookshelf的设计，root是管理员用户
			if user.ID == "root" {
				userType = "👑 管理员"
			}

			text += fmt.Sprintf("👤 *%s*\n", user.Username)
			text += fmt.Sprintf("   %s | %s\n", userType, activeStatus)
			text += fmt.Sprintf("   📅 创建于: %s\n", createdAt)
			text += fmt.Sprintf("   👀 最后在线: %s\n", lastSeen)
			text += fmt.Sprintf("   📊 播放进度: %d 个项目\n", progressCount)
			text += "\n"
		}
	}

	if messageID > 0 {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		edit.ParseMode = "Markdown"
		menu := bot_pkg.CreateUsersInfoMenu()
		edit.ReplyMarkup = &menu
		bot.Send(edit)
	} else {
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = bot_pkg.CreateUsersInfoMenu()
		bot.Send(msg)
	}
}

// sendMyStats 发送个人统计信息
func sendMyStats(bot *tgbotapi.BotAPI, chatID int64, messageID int, serverService *services.ServerService) {
	user, err := serverService.GetCurrentUserWithProgress()
	if err != nil {
		if messageID > 0 {
			editMessage(bot, chatID, messageID, "❌ 获取个人信息失败: "+err.Error())
		} else {
			sendMessage(bot, chatID, "❌ 获取个人信息失败: "+err.Error())
		}
		return
	}

	stats, err := serverService.GetListeningStats()
	if err != nil {
		if messageID > 0 {
			editMessage(bot, chatID, messageID, "❌ 获取收听统计失败: "+err.Error())
		} else {
			sendMessage(bot, chatID, "❌ 获取收听统计失败: "+err.Error())
		}
		return
	}

	// 格式化创建时间
	createdAt := "未知"
	if user.CreatedAt > 0 {
		// createdAt 是毫秒时间戳
		createdTime := time.Unix(user.CreatedAt/1000, 0).Format("2006-01-02 15:04:05")
		createdAt = createdTime
	}

	// 格式化最后在线时间
	lastSeen := "从未登录"
	if user.LastSeen > 0 {
		// lastSeen 是毫秒时间戳
		lastSeenTime := time.Unix(user.LastSeen/1000, 0).Format("2006-01-02 15:04:05")
		lastSeen = lastSeenTime
	}

	// 计算播放进度数量
	progressCount := len(user.MediaProgress)

	activeStatus := "❌ 非活跃"
	if user.IsActive {
		activeStatus = "✅ 活跃"
	}

	userType := "👤 普通用户"
	// 根据Audiobookshelf的设计，root是管理员用户
	if user.ID == "root" {
		userType = "👑 管理员"
	}

	// 获取收听统计信息 - 根据实际API文档
	var totalTimeListenStr = "0秒"
	if val, ok := stats["totalTime"]; ok {
		switch v := val.(type) {
		case float64:
			totalTimeListenStr = services.FormatDuration(time.Duration(v) * time.Second)
		case int64:
			totalTimeListenStr = services.FormatDuration(time.Duration(v) * time.Second)
		case int:
			totalTimeListenStr = services.FormatDuration(time.Duration(v) * time.Second)
		}
	}

	// items 是一个对象，而不是数组，包含以libraryItemId为键的对象
	// 获取最近播放的书籍
	recentlyPlayedText := ""
	if items, ok := stats["items"].(map[string]interface{}); ok && len(items) > 0 {
		recentlyPlayedText = "\n\n📚 *最近播放的书籍:*\n"

		// 创建一个切片来存储书籍并按时间排序
		type PlayedItem struct {
			ID            string
			TimeListening float64
			Title         string
			Author        string
		}

		var playedItems []PlayedItem

		// 遍历items对象
		for itemId, itemData := range items {
			if item, ok := itemData.(map[string]interface{}); ok {
				playedItem := PlayedItem{ID: itemId}

				// 获取收听时间
				if timeListening, ok := item["timeListening"].(float64); ok {
					playedItem.TimeListening = timeListening
				}

				// 获取媒体元数据
				if mediaMetadata, ok := item["mediaMetadata"].(map[string]interface{}); ok {
					if title, ok := mediaMetadata["title"].(string); ok {
						playedItem.Title = title
					} else {
						playedItem.Title = "未知书籍"
					}

					if author, ok := mediaMetadata["author"].(string); ok {
						playedItem.Author = author
					}
				}

				playedItems = append(playedItems, playedItem)
			}
		}

		// 按收听时间排序，收听时间最长的在前面
		sort.Slice(playedItems, func(i, j int) bool {
			return playedItems[i].TimeListening > playedItems[j].TimeListening
		})

		// 最多显示5本最近播放的书籍
		maxItems := 5
		if len(playedItems) < maxItems {
			maxItems = len(playedItems)
		}

		for i := 0; i < maxItems; i++ {
			item := playedItems[i]
			timeListeningStr := services.FormatDuration(time.Duration(item.TimeListening) * time.Second)

			// 如果有作者信息则显示
			if item.Author != "" {
				recentlyPlayedText += fmt.Sprintf("• %s\n  %s | 作者: %s\n", item.Title, timeListeningStr, item.Author)
			} else {
				recentlyPlayedText += fmt.Sprintf("• %s\n  %s\n", item.Title, timeListeningStr)
			}
		}
	}

	// 获取最近会话信息
	recentSessionsText := ""
	if recentSessions, ok := stats["recentSessions"].([]interface{}); ok && len(recentSessions) > 0 {
		recentSessionsText = "\n\n🕒 *最近会话:*\n"
		// 最多显示3个最近会话
		maxSessions := 3
		if len(recentSessions) < maxSessions {
			maxSessions = len(recentSessions)
		}

		for i := 0; i < maxSessions; i++ {
			if session, ok := recentSessions[i].(map[string]interface{}); ok {
				// 获取书籍标题
				bookTitle := "未知书籍"
				if mediaMetadata, ok := session["mediaMetadata"].(map[string]interface{}); ok {
					if title, ok := mediaMetadata["title"].(string); ok {
						bookTitle = title
					}
				}

				// 获取播放时间
				timeListeningStr := "0秒"
				if timeListening, ok := session["timeListening"].(float64); ok {
					timeListeningStr = services.FormatDuration(time.Duration(timeListening) * time.Second)
				}

				// 获取会话时间
				sessionTimeStr := ""
				if updatedAt, ok := session["updatedAt"].(float64); ok {
					sessionTime := time.Unix(int64(updatedAt/1000), 0).Format("01-02 15:04")
					sessionTimeStr = sessionTime
				}

				// 获取显示标题（可能是章节标题）
				displayTitle := ""
				if dTitle, ok := session["displayTitle"].(string); ok && dTitle != "" {
					displayTitle = fmt.Sprintf(" (%s)", dTitle)
				}

				recentSessionsText += fmt.Sprintf("• %s%s\n  %s | %s\n", bookTitle, displayTitle, timeListeningStr, sessionTimeStr)
			}
		}
	}

	text := fmt.Sprintf("*📈 我的统计信息:*\n\n")
	text += fmt.Sprintf("👤 *%s*\n", user.Username)
	text += fmt.Sprintf("   %s | %s\n", userType, activeStatus)
	text += fmt.Sprintf("   📅 创建于: %s\n", createdAt)
	text += fmt.Sprintf("   👀 最后在线: %s\n", lastSeen)
	text += fmt.Sprintf("   📊 播放进度: %d 个项目\n\n", progressCount)
	text += fmt.Sprintf("🎧 *收听统计:*\n")
	text += fmt.Sprintf("   ⏱ 总收听时间: %s\n", totalTimeListenStr)
	text += recentlyPlayedText
	text += recentSessionsText

	if messageID > 0 {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		edit.ParseMode = "Markdown"
		menu := bot_pkg.CreateMyStatsMenu()
		edit.ReplyMarkup = &menu
		bot.Send(edit)
	} else {
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = bot_pkg.CreateMyStatsMenu()
		bot.Send(msg)
	}
}

// editHelpMessage 编辑帮助信息
func editHelpMessage(bot *tgbotapi.BotAPI, chatID int64, messageID int) {
	helpText := `🎧 *Audiobookshelf 管理机器人帮助*

可用命令:
• /start - 显示主菜单
• /serverinfo - 获取服务器信息
• /users - 获取用户信息
• /libraries - 获取媒体库列表
• /search - 搜索图书
• /mystats - 获取个人统计信息
• /help - 显示此帮助信息

或者使用下方的菜单按钮进行操作。
`
	edit := tgbotapi.NewEditMessageText(chatID, messageID, helpText)
	edit.ParseMode = "Markdown"
	menu := bot_pkg.CreateMainMenu()
	edit.ReplyMarkup = &menu
	bot.Send(edit)
}

// sendMessage 发送简单文本消息
func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}

// editMessage 编辑简单文本消息
func editMessage(bot *tgbotapi.BotAPI, chatID int64, messageID int, text string) {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	bot.Send(edit)
}