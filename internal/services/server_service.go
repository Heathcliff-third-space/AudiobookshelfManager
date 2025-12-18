package services

import (
	"github.com/Heathcliff-third-space/AudiobookshelfManager/internal/api"
	"github.com/Heathcliff-third-space/AudiobookshelfManager/internal/models"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// ServerService 服务器信息服务
type ServerService struct {
	client *api.Client
	// 添加缓存相关字段
	librariesCache      []LibraryWithStats
	librariesCacheTime  time.Time
	librariesCacheMutex sync.RWMutex
	cacheExpiry         time.Duration
}

// NewServerService 创建服务器信息服务实例
func NewServerService(client *api.Client) *ServerService {
	return &ServerService{
		client:      client,
		cacheExpiry: 5 * time.Minute, // 默认5分钟缓存过期时间
	}
}

// GetFormattedServerInfo 获取格式化的服务器信息
func (s *ServerService) GetFormattedServerInfo() (string, error) {
	status, err := s.client.GetServerStatus()
	if err != nil {
		return "", fmt.Errorf("获取服务器状态失败: %w", err)
	}

	// 格式化服务器信息
	var sb strings.Builder

	sb.WriteString("📊 *Audiobookshelf 服务器信息*\n\n")

	// 注意：ServerStatus 模型中没有 App 字段，使用 ServerVersion 替代
	sb.WriteString(fmt.Sprintf("🖥 *版本*: `%s`\n", status.ServerVersion))
	sb.WriteString(fmt.Sprintf("🔤 *语言*: `%s`\n", status.Language))

	sb.WriteString("\n📚 *媒体库信息*\n")

	// 获取媒体库信息
	libraries, err := s.GetLibrariesWithStats()
	if err != nil {
		sb.WriteString("⚠️ 获取媒体库信息失败\n")
	} else {
		if len(libraries) == 0 {
			sb.WriteString("📭 暂无媒体库\n")
		} else {
			sb.WriteString(fmt.Sprintf("📁 媒体库总数: `%d`\n", len(libraries)))
			for _, lib := range libraries {
				sb.WriteString(fmt.Sprintf("📖 %s (📚 %d)\n", lib.Name, lib.ItemCount))
			}
		}
	}

	return sb.String(), nil
}

// FormatDuration 格式化持续时间
func FormatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%d天%d小时%d分钟%d秒", days, hours, minutes, seconds)
	}

	if hours > 0 {
		return fmt.Sprintf("%d小时%d分钟%d秒", hours, minutes, seconds)
	}

	if minutes > 0 {
		return fmt.Sprintf("%d分钟%d秒", minutes, seconds)
	}

	return fmt.Sprintf("%d秒", seconds)
}

// FormatBytes 格式化字节数
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// LibraryWithStats 带有统计信息的媒体库
type LibraryWithStats struct {
	models.LibraryInfo
	ItemCount int `json:"item_count"`
}

// GetLibrariesWithStats 获取带有统计信息的媒体库列表，带缓存功能
func (s *ServerService) GetLibrariesWithStats() ([]LibraryWithStats, error) {
	// 检查缓存
	s.librariesCacheMutex.RLock()
	if time.Since(s.librariesCacheTime) < s.cacheExpiry && s.librariesCache != nil {
		cached := s.librariesCache
		s.librariesCacheMutex.RUnlock()
		return cached, nil
	}
	s.librariesCacheMutex.RUnlock()

	// 缓存失效，获取新数据
	libraries, err := s.client.GetLibrariesInfo()
	if err != nil {
		return nil, err
	}

	// 获取每个库的详细统计信息，使用并行处理提高性能
	librariesWithStats := make([]LibraryWithStats, len(libraries))

	// 使用并行处理，最大并发数为4
	const maxConcurrency = 4
	semaphore := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	var mu sync.Mutex

	// 并行获取每个库的统计信息
	for i, library := range libraries {
		wg.Add(1)
		go func(index int, lib models.LibraryInfo) {
			defer wg.Done()

			// 控制并发数
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 获取库中媒体项的数量
			count, err := s.client.GetLibraryItemsCount(lib.ID)
			if err != nil {
				// 如果获取失败，设置为0
				mu.Lock()
				librariesWithStats[index].LibraryInfo = lib
				librariesWithStats[index].ItemCount = 0
				mu.Unlock()
			} else {
				mu.Lock()
				librariesWithStats[index].LibraryInfo = lib
				librariesWithStats[index].ItemCount = count
				mu.Unlock()
			}
		}(i, library)
	}

	// 等待所有goroutine完成
	wg.Wait()

	// 更新缓存
	s.librariesCacheMutex.Lock()
	s.librariesCache = librariesWithStats
	s.librariesCacheTime = time.Now()
	s.librariesCacheMutex.Unlock()

	return librariesWithStats, nil
}

// GetLibraryName 根据libraryId获取媒体库名称
func (s *ServerService) GetLibraryName(libraryId string) (string, error) {
	// 使用轻量级方法获取媒体库名称，避免获取统计信息
	libraries, err := s.getLibrariesBasicInfo()
	if err != nil {
		return "", err
	}

	// 查找指定ID的媒体库
	for _, lib := range libraries {
		if lib.ID == libraryId {
			return lib.Name, nil
		}
	}

	// 如果没有找到对应的媒体库，返回空字符串
	return "", fmt.Errorf("未找到ID为%s的媒体库", libraryId)
}

// getLibrariesBasicInfo 获取媒体库基本信息（ID和名称），不包含统计信息
func (s *ServerService) getLibrariesBasicInfo() ([]models.LibraryInfo, error) {
	// 直接调用API获取媒体库信息，不计算统计信息
	return s.client.GetLibrariesInfo()
}

// GetUsersWithProgress 获取用户列表及播放统计信息
func (s *ServerService) GetUsersWithProgress() ([]models.UserInfo, error) {
	// 获取用户列表
	users, err := s.client.GetUsers()
	if err != nil {
		return nil, fmt.Errorf("获取用户列表失败: %w", err)
	}

	// 获取每个用户的播放统计信息
	// 使用并行处理，最大并发数为4
	const maxConcurrency = 4
	semaphore := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := range users {
		wg.Add(1)
		go func(index int, user models.UserInfo) {
			defer wg.Done()

			// 控制并发数
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 获取用户的播放进度信息
			progress, err := s.client.GetUserMediaProgress(user.ID)
			if err != nil {
				// 如果获取失败，记录错误但不中断其他用户的信息获取
				log.Printf("获取用户 %s 的播放进度信息失败: %v", user.Username, err)
				return
			}

			// 更新用户信息
			mu.Lock()
			users[index].MediaProgress = progress
			mu.Unlock()
		}(i, users[i])
	}

	// 等待所有goroutine完成
	wg.Wait()

	return users, nil
}

// SearchBooks 搜索图书，使用并行处理提高性能
func (s *ServerService) SearchBooks(term string, libraryID string) ([]models.Book, error) {
	if term == "" {
		return nil, fmt.Errorf("搜索词不能为空")
	}

	books, err := s.client.SearchBooks(term, libraryID)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	return books, nil
}

// GetCurrentUserWithProgress 获取当前用户信息及播放统计
func (s *ServerService) GetCurrentUserWithProgress() (*models.UserInfo, error) {
	// 获取当前用户信息
	user, err := s.client.GetCurrentUser()
	if err != nil {
		return nil, fmt.Errorf("获取当前用户信息失败: %w", err)
	}

	// 获取用户的播放进度信息
	progress, err := s.client.GetUserMediaProgress(user.ID)
	if err != nil {
		log.Printf("获取用户 %s 的播放进度信息失败: %v", user.Username, err)
		// 即使获取播放进度失败，也返回用户基本信息
		return user, nil
	}

	user.MediaProgress = progress
	return user, nil
}

// GetListeningStats 获取当前用户的收听统计信息
func (s *ServerService) GetListeningStats() (map[string]interface{}, error) {
	stats, err := s.client.GetListeningStats()
	if err != nil {
		return nil, fmt.Errorf("获取收听统计信息失败: %w", err)
	}
	return stats, nil
}
