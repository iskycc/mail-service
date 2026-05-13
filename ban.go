package main

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"
)

func getIPBanKey(ip string) string {
	return fmt.Sprintf("login_ban:%s", ip)
}

func getIPFailKey(ip string) string {
	return fmt.Sprintf("login_fail:%s", ip)
}

func isIPBanned(ctx context.Context, ip string) (bool, int64) {
	banKey := getIPBanKey(ip)
	val, err := rdb.Get(ctx, banKey).Result()
	if err != nil {
		return false, 0
	}
	banUntil, _ := strconv.ParseInt(val, 10, 64)
	now := time.Now().Unix()
	if now < banUntil {
		return true, banUntil - now
	}
	// 已过期，清理
	rdb.Del(ctx, banKey)
	rdb.Del(ctx, getIPFailKey(ip))
	return false, 0
}

func recordLoginFail(ctx context.Context, ip string) (int, int64) {
	failKey := getIPFailKey(ip)
	count, _ := rdb.Incr(ctx, failKey).Result()
	rdb.Expire(ctx, failKey, 24*time.Hour)

	if count >= 3 {
		exponent := float64(count - 3)
		banMinutes := 5 * math.Pow(2, exponent)
		if banMinutes > 1440 { // 上限 24 小时
			banMinutes = 1440
		}
		banSeconds := int64(banMinutes * 60)
		banUntil := time.Now().Unix() + banSeconds
		banKey := getIPBanKey(ip)
		rdb.Set(ctx, banKey, banUntil, time.Duration(banSeconds)*time.Second)
		return int(count), banSeconds
	}
	return int(count), 0
}

func clearLoginFail(ctx context.Context, ip string) {
	rdb.Del(ctx, getIPFailKey(ip))
	rdb.Del(ctx, getIPBanKey(ip))
}
