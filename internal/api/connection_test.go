package api

import (
	"AudiobookshelfManager/internal/config"
	"context"
	"os"
	"testing"
	"time"
)

// TestAudiobookshelfConnection 测试 Audiobookshelf 服务器连接
func TestAudiobookshelfConnection(t *testing.T) {
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
		libraries []byte
		err       error
	}

	// 在 goroutine 中运行测试
	go func() {
		// 加载配置
		cfg := config.LoadConfig()

		// 检查是否配置了 Audiobookshelf Token
		if cfg.AudiobookshelfToken == "" {
			t.Log("跳过测试: 未设置 AUDIOBOOKSHELF_TOKEN")
			done <- true
			return
		}

		// 创建客户端
		client := NewClient(cfg)

		// 测试连接
		t.Log("正在测试 Audiobookshelf 连接...")
		testResult.libraries, testResult.err = client.GetLibraries()
		done <- true
	}()

	// 等待测试完成或超时
	select {
	case <-done:
		// 测试完成
		if testResult.err != nil {
			t.Logf("无法连接到 Audiobookshelf 服务器: %v", testResult.err)
			t.Log("这可能是由于网络问题、服务器不可达或认证失败导致的")
		} else {
			t.Log("✅ 成功连接到 Audiobookshelf 服务器")
			t.Logf("响应数据长度: %d 字节", len(testResult.libraries))
		}
	case <-ctx.Done():
		// 测试超时
		t.Fatal("测试超时")
	}
}