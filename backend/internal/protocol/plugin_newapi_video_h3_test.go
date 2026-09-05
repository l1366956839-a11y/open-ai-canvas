package protocol

import (
	"archive/zip"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// readPluginManifest 从 .yingce-plugin（zip）包中读出 manifest.json。
func readPluginManifest(t *testing.T, pkg string) []byte {
	t.Helper()
	r, err := zip.OpenReader(pkg)
	if err != nil {
		t.Skipf("跳过：无法打开插件包 %s (%v)", pkg, err)
	}
	defer r.Close()
	for _, f := range r.File {
		if filepath.Base(f.Name) != "manifest.json" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("读取 %s 失败: %v", f.Name, err)
		}
		defer rc.Close()
		data, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("读取 %s 内容失败: %v", f.Name, err)
		}
		return data
	}
	t.Skipf("跳过：插件包中没有 manifest.json")
	return nil
}

// TestNewAPIVideoPluginHailuoH3RequestBody 验证 newapi-video-generations-v1 插件模板：
// H3 发整型 duration 且不带 seconds / generate_audio；非 H3 保持 seconds 字符串。
func TestNewAPIVideoPluginHailuoH3RequestBody(t *testing.T) {
	pkg := filepath.Join("..", "..", "..", "plugin-packages", "newapi-video-generations-v1.yingce-plugin")
	if _, err := os.Stat(pkg); err != nil {
		t.Skipf("跳过：未找到插件包 %s", pkg)
	}
	adapter, err := LoadManifest(readPluginManifest(t, pkg))
	if err != nil {
		t.Fatalf("LoadManifest 失败: %v", err)
	}
	t.Logf("provider = %s, execution = %q", adapter.Metadata().ID, adapter.Metadata().Execution)

	cases := []struct {
		name       string
		model      string
		resolution string
		duration   int
		wantDur    any    // 期望 duration 值（nil 表示不应出现）
		wantSec    bool   // 是否期望出现 seconds
		wantAudio  bool   // 是否期望出现 generate_audio
	}{
		{"H3-768p", "minimax_h3", "768p", 5, 5, false, false},
		{"H3-1080p-超长截断到8", "minimax_h3", "1080p", 12, 8, false, false},
		{"H3-时长不足抬到4", "MiniMax-H3", "768p", 2, 4, false, false},
		{"非H3-保持seconds", "sora-2", "768p", 6, nil, true, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec, err := adapter.BuildCreate(context.Background(), RequestContext{
				BaseURL: "https://new.xlcsh.top",
				Request: GenerationRequest{
					Model:         tc.model,
					Prompt:        "test",
					Duration:      tc.duration,
					AspectRatio:   "16:9",
					Resolution:    tc.resolution,
					GenerateAudio: true,
				},
			})
			if err != nil {
				t.Fatalf("BuildCreate 失败: %v", err)
			}
			raw, _ := json.Marshal(spec.Body)
			body, _ := spec.Body.(map[string]any)
			t.Logf("model=%s resolution=%s duration=%d -> %s", tc.model, tc.resolution, tc.duration, string(raw))

			if tc.wantSec {
				if _, ok := body["seconds"]; !ok {
					t.Errorf("非 H3 应发送 seconds，实际 %s", string(raw))
				}
			} else if _, ok := body["seconds"]; ok {
				t.Errorf("H3 不应发送 seconds，实际 %s", string(raw))
			}

			if tc.wantAudio {
				if _, ok := body["generate_audio"]; !ok {
					t.Errorf("非 H3 应发送 generate_audio，实际 %s", string(raw))
				}
			} else if _, ok := body["generate_audio"]; ok {
				t.Errorf("H3 不应发送 generate_audio，实际 %s", string(raw))
			}

			got, hasDur := body["duration"]
			if tc.wantDur == nil {
				if hasDur {
					t.Errorf("非 H3 不应发送 duration，实际 %s", string(raw))
				}
				return
			}
			if !hasDur {
				t.Fatalf("H3 应发送 duration，实际 %s", string(raw))
			}
			if got != tc.wantDur {
				t.Errorf("duration = %v (%T), want %v", got, got, tc.wantDur)
			}
			if _, ok := got.(float64); ok {
				t.Errorf("duration 必须是整型，实际是浮点 %v", got)
			}
		})
	}
}
