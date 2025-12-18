package main

import (
	"github.com/Heathcliff-third-space/AudiobookshelfManager/internal/config"
	"testing"
)

// TestIntegration 集成测试
func TestIntegration(t *testing.T) {
	// 加载配置
	cfg := config.LoadConfig()

	// 检查是否加载了配置
	if cfg == nil {
		t.Error("无法加载配置")
	}

	// 验证关键配置项存在
	if cfg.AudiobookshelfPort == 0 {
		t.Error("AudiobookshelfPort 未正确设置")
	}

	t.Logf("配置加载成功: %+v", cfg)
}
