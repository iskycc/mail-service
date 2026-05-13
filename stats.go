package main

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"
)

func calculateStats(ctx context.Context, days int) (*StatsResponse, error) {
	now := time.Now()
	startDate := now.Add(-time.Duration(days-1) * 24 * time.Hour).Truncate(24 * time.Hour)

	// 初始化每天统计
	dailyMap := make(map[string]*DailyStat)
	for i := 0; i < days; i++ {
		date := now.Add(-time.Duration(i) * 24 * time.Hour).Format("2006-01-02")
		dailyMap[date] = &DailyStat{Date: date}
	}

	domainMap := make(map[string]int)
	var total, success, failed int

	var cursor uint64
	for {
		keys, nextCursor, err := rdb.Scan(ctx, cursor, "mail_log:*", 100).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			// 读取必要字段
			fields, err := rdb.HMGet(ctx, key, "time_unix", "user", "result").Result()
			if err != nil || len(fields) < 3 || fields[0] == nil || fields[1] == nil || fields[2] == nil {
				continue
			}

			ts, _ := strconv.ParseInt(fields[0].(string), 10, 64)
			recordTime := time.Unix(ts, 0)
			if recordTime.Before(startDate) {
				continue
			}

			user := fields[1].(string)
			result := fields[2].(string)

			total++
			isSuccess := strings.Contains(result, "Message has been sent")
			if isSuccess {
				success++
			} else {
				failed++
			}

			// 按日期聚合
			dateStr := recordTime.Format("2006-01-02")
			if ds, ok := dailyMap[dateStr]; ok {
				ds.Total++
				if isSuccess {
					ds.Success++
				} else {
					ds.Failed++
				}
			}

			// 按邮箱后缀聚合
			if user != "" {
				parts := strings.Split(user, "@")
				if len(parts) == 2 {
					domain := parts[1]
					domainMap[domain]++
				}
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	// 组装 dailyStats（按日期倒序）
	dailyStats := make([]DailyStat, 0, days)
	for i := 0; i < days; i++ {
		date := now.Add(-time.Duration(i) * 24 * time.Hour).Format("2006-01-02")
		if ds, ok := dailyMap[date]; ok {
			dailyStats = append(dailyStats, *ds)
		}
	}

	// 组装 domainStats（按数量倒序）
	domainStats := make([]DomainStat, 0)
	for domain, count := range domainMap {
		pct := 0.0
		if total > 0 {
			pct = float64(count) / float64(total) * 100
		}
		domainStats = append(domainStats, DomainStat{
			Domain:     domain,
			Count:      count,
			Percentage: float64(int(pct*100)) / 100,
		})
	}

	// 简单排序：按数量降序
	for i := 0; i < len(domainStats); i++ {
		for j := i + 1; j < len(domainStats); j++ {
			if domainStats[j].Count > domainStats[i].Count {
				domainStats[i], domainStats[j] = domainStats[j], domainStats[i]
			}
		}
	}

	successRate := 0.0
	if total > 0 {
		successRate = float64(success) / float64(total) * 100
	}

	log.Printf("[Admin] 统计完成: 总发信=%d, 成功=%d, 失败=%d", total, success, failed)

	return &StatsResponse{
		Total:       total,
		Success:     success,
		Failed:      failed,
		SuccessRate: float64(int(successRate*100)) / 100,
		DailyStats:  dailyStats,
		DomainStats: domainStats,
	}, nil
}
