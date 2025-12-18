package api

import (
	"github.com/Heathcliff-third-space/AudiobookshelfManager/internal/config"
	"testing"
)

func TestNewClient(t *testing.T) {
	// 创建测试配置
	cfg := &config.Config{
		AudiobookshelfURL:   "http://localhost:13378",
		AudiobookshelfToken: "test_token",
		AudiobookshelfPort:  13378,
		ProxyAddress:        "", // 不使用代理
	}

	// 创建客户端
	client := NewClient(cfg)

	// 验证客户端是否正确创建
	if client == nil {
		t.Error("期望创建一个有效的客户端，但得到了 nil")
	}

	if client.baseURL != "http://localhost:13378" {
		t.Errorf("期望 baseURL 为 'http://localhost:13378'，实际得到 '%s'", client.baseURL)
	}

	if client.token != "test_token" {
		t.Errorf("期望 token 为 'test_token'，实际得到 '%s'", client.token)
	}

	// 测试默认 URL
	cfg.AudiobookshelfURL = ""
	client = NewClient(cfg)
	expectedURL := "http://localhost:13378"
	if client.baseURL != expectedURL {
		t.Errorf("期望 baseURL 为 '%s'，实际得到 '%s'", expectedURL, client.baseURL)
	}
}