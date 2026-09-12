package kernel

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"
)

// NewID 生成唯一标识符。
func NewID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

// FirstNonEmpty 返回第一个非空字符串。
func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// UniqueNonEmpty 去重并过滤空字符串。
func UniqueNonEmpty(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

// DefaultString 返回非空值或回退值。
func DefaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

// Ptr 返回值的指针。
func Ptr[T any](value T) *T {
	return &value
}

// StringValue 从任意值提取字符串。
func StringValue(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprintf("%v", value)
}

// TruncateRunes 截断字符串到指定 rune 数。
func TruncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "..."
}

// LogJSONDecodeFailure 记录持久化 JSON 字段解析失败。
// 这些字段解析失败后调用方通常继续用零值往下走，没有日志就无法定位是哪条记录、
// 哪个字段的数据损坏，表现为「功能莫名不对」却查不到原因。
// 空的原始值代表字段尚未写入，是正常状态，不记录以免刷日志。
func LogJSONDecodeFailure(scope string, err error, raw string) {
	if strings.TrimSpace(raw) == "" {
		return
	}
	log.Printf("%s: JSON 解析失败，已按零值继续：err=%v, raw=%s", scope, err, TruncateRunes(raw, 200))
}

// Megabytes 转换 MB 到字节。
func Megabytes(value int64) int64 { return value << 20 }

// Gigabytes 转换 GB 到字节。
func Gigabytes(value int64) int64 { return value << 30 }
