package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestMain(m *testing.M) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       15, // 使用独立 DB 避免污染
	})
	code := m.Run()
	rdb.FlushDB(context.Background())
	rdb.Close()
	os.Exit(code)
}

func TestRecordAndClearLoginFail(t *testing.T) {
	ctx := context.Background()
	ip := "192.168.1.100"

	// 清理历史数据
	clearLoginFail(ctx, ip)

	// 第 1 次失败，不封禁
	count, banSecs := recordLoginFail(ctx, ip)
	if count != 1 {
		t.Errorf("expected count=1, got %d", count)
	}
	if banSecs != 0 {
		t.Errorf("expected banSecs=0, got %d", banSecs)
	}

	// 第 2 次失败，仍不封禁
	count, banSecs = recordLoginFail(ctx, ip)
	if count != 2 {
		t.Errorf("expected count=2, got %d", count)
	}
	if banSecs != 0 {
		t.Errorf("expected banSecs=0, got %d", banSecs)
	}

	// 第 3 次失败，触发封禁 5 分钟
	count, banSecs = recordLoginFail(ctx, ip)
	if count != 3 {
		t.Errorf("expected count=3, got %d", count)
	}
	if banSecs != 5*60 {
		t.Errorf("expected banSecs=300, got %d", banSecs)
	}

	// 验证被封禁
	banned, remaining := isIPBanned(ctx, ip)
	if !banned {
		t.Error("expected banned=true")
	}
	if remaining < 290 || remaining > 300 {
		t.Errorf("expected remaining ~300, got %d", remaining)
	}

	// 清除封禁
	clearLoginFail(ctx, ip)
	banned, _ = isIPBanned(ctx, ip)
	if banned {
		t.Error("expected banned=false after clear")
	}
}

func TestIsIPBanned_Expired(t *testing.T) {
	ctx := context.Background()
	ip := "192.168.1.101"

	clearLoginFail(ctx, ip)

	// 模拟已过期封禁：手动写入一个过去的时间戳
	banKey := getIPBanKey(ip)
	rdb.Set(ctx, banKey, time.Now().Unix()-1, time.Hour)

	banned, _ := isIPBanned(ctx, ip)
	if banned {
		t.Error("expected banned=false for expired ban")
	}
}
