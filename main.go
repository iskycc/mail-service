package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/gomail.v2"
)

const version = "1.10.13"

var (
	rdb            *redis.Client
	mysqlDSN       string
	db             *sql.DB
	trustedProxies []string
)

func initTrustedProxies() {
	proxies := os.Getenv("TRUSTED_PROXIES")
	if proxies != "" {
		for _, p := range strings.Split(proxies, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				trustedProxies = append(trustedProxies, p)
			}
		}
	}
}

func isTrustedProxy(host string) bool {
	for _, p := range trustedProxies {
		if p == host {
			return true
		}
	}
	return false
}

type RecordParams struct {
	TimeUnix   int64
	Ip         string
	User       string
	Subject    string
	Body       string
	Altbody    string
	Tname      string
	MailID     int
	Result     string
	DurationMs int64
}

var updateMailIDScript = redis.NewScript(`
local mailid = redis.call('GET', KEYS[1])
if not mailid then
    mailid = 1
    redis.call('SET', KEYS[1], mailid)
else
    mailid = tonumber(mailid)
    local maxId = tonumber(KEYS[2])
    if maxId < 1 then maxId = 5 end
    if mailid >= maxId then
        mailid = 1
    else
        mailid = mailid + 1
    end
    redis.call('SET', KEYS[1], mailid)
end
return mailid
`)

func getMailCount(ctx context.Context) int {
	val, err := rdb.Get(ctx, "mail_count").Int()
	if err == nil && val > 0 {
		return val
	}
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mail").Scan(&count)
	if err != nil || count < 1 {
		count = 5
	}
	rdb.Set(ctx, "mail_count", count, 1*time.Hour)
	return count
}

func main() {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PWD")

	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlPort := os.Getenv("MYSQL_PORT")
	if mysqlPort == "" {
		mysqlPort = "3306"
	}
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPassword := os.Getenv("MYSQL_PWD")
	mysqlDB := os.Getenv("MYSQL_DBNAME")

	mysqlDSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", mysqlUser, mysqlPassword, mysqlHost, mysqlPort, mysqlDB)

	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       0,
	})

	var err error
	db, err = sql.Open("mysql", mysqlDSN)
	if err != nil {
		log.Fatalf("MySQL connection failed: %v", err)
	}
	if err := initMailTable(); err != nil {
		log.Printf("Failed to init mail table: %v", err)
	}
	if err := initMailLogTable(); err != nil {
		log.Printf("Failed to init mail_log table: %v", err)
	}
	if err := initAuditLogTable(); err != nil {
		log.Printf("Failed to init audit_log table: %v", err)
	}
	if err := initMailTemplateTable(); err != nil {
		log.Printf("Failed to init mail_template table: %v", err)
	}

	initTrustedProxies()
	go syncRedisFromMySQL()

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/", handler)
	router.POST("/", handler)
	router.GET("/health", healthHandler)
	registerAdminRoutes(router)

	srv := &http.Server{Addr: ":22125", Handler: router}
	log.Println("Server starting on :22125...")

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}

func healthHandler(c *gin.Context) {
	clientIP := net.ParseIP(getClientIP(c.Request))
	if clientIP != nil && !clientIP.IsPrivate() && !clientIP.IsLoopback() {
		c.JSON(http.StatusForbidden, gin.H{"error": "health endpoint is internal only"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	mysqlOK := db.PingContext(ctx) == nil
	redisOK := rdb.Ping(ctx).Err() == nil

	status := http.StatusOK
	if !mysqlOK || !redisOK {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"mysql":  mysqlOK,
		"redis":  redisOK,
		"status": map[bool]string{true: "ok", false: "error"}[mysqlOK && redisOK],
	})
}

func renderTemplate(text string, vars map[string]string) string {
	for k, v := range vars {
		text = strings.ReplaceAll(text, "{{"+k+"}}", v)
	}
	return text
}

func handler(c *gin.Context) {
	var user, subject, body, altbody, tname, templateIDStr string
	var vars map[string]string

	requestID := fmt.Sprintf("%s-%04d", time.Now().Format("20060102150405"), rand.Intn(10000))

	if c.Request.Method == http.MethodGet {
		user = c.Query("user")
		subject = c.Query("subject")
		body = c.Query("body")
		altbody = c.Query("altbody")
		tname = c.Query("tname")
		if v := c.Query("vars"); v != "" {
			_ = json.Unmarshal([]byte(v), &vars)
		}
	} else {
		contentType := c.ContentType()
		if strings.Contains(contentType, "application/json") {
			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"info":    "Invalid JSON format",
				})
				return
			}
			user, _ = data["user"].(string)
			subject, _ = data["subject"].(string)
			body, _ = data["body"].(string)
			altbody, _ = data["altbody"].(string)
			tname, _ = data["tname"].(string)
			if v, ok := data["vars"].(map[string]interface{}); ok {
				vars = make(map[string]string)
				for k, val := range v {
					vars[k], _ = val.(string)
				}
			}
			if tid, ok := data["template_id"].(string); ok && tid != "" {
				templateIDStr = tid
			} else if tidFloat, ok := data["template_id"].(float64); ok {
				templateIDStr = strconv.Itoa(int(tidFloat))
			}
		} else {
			user = c.PostForm("user")
			subject = c.PostForm("subject")
			body = c.PostForm("body")
			altbody = c.PostForm("altbody")
			tname = c.PostForm("tname")
			if v := c.PostForm("vars"); v != "" {
				_ = json.Unmarshal([]byte(v), &vars)
			}
		}
	}

	// 如果传了 template_id，读取模板并替换变量
	if templateIDStr == "" {
		templateIDStr = c.Query("template_id")
	}
	if templateIDStr == "" {
		templateIDStr = c.PostForm("template_id")
	}
	if templateIDStr != "" {
		tid, _ := strconv.Atoi(templateIDStr)
		if tid > 0 {
			tpl, err := getTemplate(tid)
			if err == nil {
				if subject == "" {
					subject = tpl.Subject
				}
				if body == "" {
					body = tpl.Body
				}
			}
		}
	}

	if vars != nil {
		subject = renderTemplate(subject, vars)
		body = renderTemplate(body, vars)
	}

	if tname == "" {
		tname = "Libv团队"
	}
	if altbody == "" {
		altbody = body
	}

	ip := getClientIP(c.Request)
	log.Printf("[发送请求] request_id=%s 来源IP:%s 收件人:%s", requestID, ip, user)
	var records []RecordParams
	maxRetries := 3
	var success bool
	var resultInfo string

	ctx := context.Background()

	for i := 0; i < maxRetries; i++ {
		mailID, err := updateMailID(ctx)
		if err != nil {
			resultInfo = fmt.Sprintf("Redis error: %v", err)
			log.Printf("[发送失败] request_id=%s 来源IP:%s | 错误类型:Redis | 详情:%v", requestID, ip, err)
			now := time.Now().Unix()
			records = append(records, RecordParams{
				TimeUnix:   now,
				Ip:         ip,
				User:       user,
				Subject:    subject,
				Body:       body,
				Altbody:    altbody,
				Tname:      tname,
				MailID:     mailID,
				Result:     resultInfo,
				DurationMs: 0,
			})
			continue
		}

		if user == "" || subject == "" {
			success = false
			resultInfo = "Unknown Request, Please try again"
			log.Printf("[发送失败] request_id=%s 来源IP:%s | 错误:参数缺失 user=%q subject=%q", requestID, ip, user, subject)
			break
		}

		host, port, sender, password, err := getMailConfig(mailID)
		if err != nil {
			resultInfo = fmt.Sprintf("MySQL error: %v", err)
			now := time.Now().Unix()
			records = append(records, RecordParams{
				TimeUnix:   now,
				Ip:         ip,
				User:       user,
				Subject:    subject,
				Body:       body,
				Altbody:    altbody,
				Tname:      tname,
				MailID:     mailID,
				Result:     resultInfo,
				DurationMs: 0,
			})
			log.Printf("[发送失败] request_id=%s 来源IP:%s | 邮件ID:%d | 错误类型:MySQL | 收件人:%s | 详情:%v",
				requestID, ip, mailID, user, err)
			continue
		}

		sendStart := time.Now()
		sendResultCh := make(chan string, 1)
		go func() {
			sendResultCh <- sendEmail(host, port, sender, password, user, subject, body, altbody, tname)
		}()

		var sendResult string
		select {
		case sendResult = <-sendResultCh:
		case <-time.After(15 * time.Second):
			sendResult = "SMTP timeout: email sending exceeded 15 seconds"
		}
		durationMs := time.Since(sendStart).Milliseconds()

		if sendResult == "" {
			success = true
			resultInfo = "Message has been sent"
			log.Printf("[发送成功] request_id=%s 时间:%s | 来源IP:%s | 邮件ID:%d | 发件人:%s | 收件人:%s",
				requestID, time.Now().Format(time.RFC3339), ip, mailID, sender, user)
		} else {
			resultInfo = sendResult
			log.Printf("[发送失败] request_id=%s 时间:%s | 来源IP:%s | 邮件ID:%d | 发件人:%s | 收件人:%s | 错误:%s",
				requestID, time.Now().Format(time.RFC3339), ip, mailID, sender, user, sendResult)
		}

		now := time.Now().Unix()
		records = append(records, RecordParams{
			TimeUnix:   now,
			Ip:         ip,
			User:       user,
			Subject:    subject,
			Body:       body,
			Altbody:    altbody,
			Tname:      tname,
			MailID:     mailID,
			Result:     resultInfo,
			DurationMs: durationMs,
		})

		if success {
			break
		}
	}

	response := gin.H{
		"success":    success,
		"info":       resultInfo,
		"request_id": requestID,
	}
	c.JSON(http.StatusOK, response)

	go func(recs []RecordParams) {
		for _, rec := range recs {
			saveRecord(context.Background(), rec.TimeUnix, rec.Ip, rec.User, rec.Subject, rec.Body, rec.Altbody, rec.Tname, rec.MailID, rec.Result, rec.DurationMs)
		}
		if len(recs) > 0 {
			last := recs[len(recs)-1]
			event := map[string]interface{}{
				"time_unix":   last.TimeUnix,
				"ip":          last.Ip,
				"user":        last.User,
				"subject":     last.Subject,
				"body":        last.Body,
				"mail_id":     last.MailID,
				"result":      last.Result,
				"duration_ms": last.DurationMs,
				"success":     strings.Contains(last.Result, "Message has been sent"),
			}
			if data, err := json.Marshal(event); err == nil {
				eventHub.Broadcast(string(data))
			}
		}
	}(records)
}

func updateMailID(ctx context.Context) (int, error) {
	maxId := getMailCount(ctx)
	return updateMailIDScript.Run(ctx, rdb, []string{"mailid", strconv.Itoa(maxId)}).Int()
}

func getMailConfig(mailID int) (string, int, string, string, error) {
	// 首先尝试从Redis获取
	ctx := context.Background()
	cacheKey := fmt.Sprintf("mailconfig:%d", mailID)

	// 尝试获取缓存
	result, err := rdb.HMGet(ctx, cacheKey, "host", "port", "sender", "password").Result()
	if err == nil && result[0] != nil {
		// 缓存命中，解析数据
		host := result[0].(string)
		port, _ := strconv.Atoi(result[1].(string))
		sender := result[2].(string)
		password := result[3].(string)
		log.Printf("[MailConfig] Mail ID: %d cache hit (Redis)", mailID)
		return host, port, sender, password, nil
	}

	// 缓存未命中，从MySQL获取
	var host string
	var port int
	var sender string
	var password string
	err = db.QueryRow("SELECT domain, port, sender, password FROM mail WHERE id = ?", mailID).Scan(
		&host, &port, &sender, &password,
	)

	if err != nil {
		return "", 0, "", "", fmt.Errorf("query failed: %v", err)
	}

	rdb.HSet(ctx, cacheKey,
		"host", host,
		"port", strconv.Itoa(port),
		"sender", sender,
		"password", password,
	)

	log.Printf("[MailConfig] Mail ID: %d cache miss, loaded from MySQL", mailID)

	rdb.Expire(ctx, cacheKey, 7*24*time.Hour)

	return host, port, sender, password, nil
}

func sendEmail(host string, port int, sender, password, to, subject, body, altbody, tname string) string {
	m := gomail.NewMessage()
	encodedTname := "=?utf-8?B?" + base64.StdEncoding.EncodeToString([]byte(tname)) + "?="
	m.SetHeader("From", m.FormatAddress(sender, encodedTname))
	m.SetHeader("To", to)
	encodedSubject := "=?utf-8?B?" + base64.StdEncoding.EncodeToString([]byte(subject)) + "?="
	m.SetHeader("Subject", encodedSubject)
	m.SetBody("text/html", body)
	m.AddAlternative("text/plain", altbody)

	dialer := gomail.NewDialer(host, port, sender, password)
	dialer.SSL = true

	closer, err := dialer.Dial()
	if err != nil {
		return fmt.Sprintf("SMTP dial error: %v", err)
	}
	defer closer.Close()

	if err = gomail.Send(closer, m); err != nil {
		return fmt.Sprintf("Message could not be sent. Mailer Error: %v", err)
	}
	return ""
}

func saveRecord(ctx context.Context, timeUnix int64, ip, user, subject, body, altbody, tname string, mailID int, result string, durationMs int64) {
	key := fmt.Sprintf("mail_log:%d_%d", timeUnix, time.Now().UnixNano())
	fields := map[string]interface{}{
		"time_unix":   timeUnix,
		"ip":          ip,
		"user":        user,
		"subject":     subject,
		"body":        body,
		"altbody":     altbody,
		"team_name":   tname,
		"mailid":      mailID,
		"result":      result,
		"duration_ms": durationMs,
	}
	if err := rdb.HMSet(ctx, key, fields).Err(); err != nil {
		log.Printf("Failed to save record to redis: %v", err)
		return
	}
	if err := rdb.Expire(ctx, key, 7*24*time.Hour).Err(); err != nil {
		log.Printf("Failed to set expiration: %v", err)
	}

	// MySQL 持久化再开独立 goroutine，确保不影响 Redis 写入，也不阻塞主流程
	go func(timeUnix int64, ip, user, subject, body, altbody, tname string, mailID int, result string, durationMs int64) {
		rec := RecordParams{
			TimeUnix:   timeUnix,
			Ip:         ip,
			User:       user,
			Subject:    subject,
			Body:       body,
			Altbody:    altbody,
			Tname:      tname,
			MailID:     mailID,
			Result:     result,
			DurationMs: durationMs,
		}
		if err := insertMailLog(context.Background(), rec); err != nil {
			log.Printf("Failed to save record to mysql: %v", err)
		}
	}(timeUnix, ip, user, subject, body, altbody, tname, mailID, result, durationMs)
}

func syncRedisFromMySQL() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 不再删除现有 mail_log:* 键，仅从 MySQL 回填缺失记录。
	// 直接删除可能丢失 Redis 中比 MySQL 更新的未落盘数据。

	var totalRows int
	_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mail_log").Scan(&totalRows)
	log.Printf("[Sync] MySQL mail_log total rows: %d", totalRows)

	// 查询近 7 天的 MySQL 记录
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour).Unix()
	log.Printf("[Sync] Query condition: time_unix >= %d (%s)", sevenDaysAgo, time.Unix(sevenDaysAgo, 0).Format("2006-01-02 15:04:05"))

	rows, err := db.QueryContext(ctx, "SELECT time_unix, ip, user, subject, body, altbody, team_name, mail_id, result, duration_ms FROM mail_log WHERE time_unix >= ?", sevenDaysAgo)
	if err != nil {
		log.Printf("[Sync] Failed to query MySQL for sync: %v", err)
		return
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var rec RecordParams
		if err := rows.Scan(&rec.TimeUnix, &rec.Ip, &rec.User, &rec.Subject, &rec.Body, &rec.Altbody, &rec.Tname, &rec.MailID, &rec.Result, &rec.DurationMs); err != nil {
			log.Printf("[Sync] Failed to scan row for sync: %v", err)
			continue
		}
		key := fmt.Sprintf("mail_log:%d_%d", rec.TimeUnix, time.Now().UnixNano()+int64(count))
		fields := map[string]interface{}{
			"time_unix":   rec.TimeUnix,
			"ip":          rec.Ip,
			"user":        rec.User,
			"subject":     rec.Subject,
			"body":        rec.Body,
			"altbody":     rec.Altbody,
			"team_name":   rec.Tname,
			"mailid":      rec.MailID,
			"result":      rec.Result,
			"duration_ms": rec.DurationMs,
		}
		if err := rdb.HMSet(ctx, key, fields).Err(); err != nil {
			log.Printf("[Sync] Failed to sync record to redis: %v", err)
			continue
		}
		if err := rdb.Expire(ctx, key, 7*24*time.Hour).Err(); err != nil {
			log.Printf("[Sync] Failed to set expiration for synced record: %v", err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		log.Printf("[Sync] Rows iteration error: %v", err)
	}

	log.Printf("[Sync] Synced %d records from MySQL to Redis", count)
}

func getClientIP(r *http.Request) string {
	remoteHost := r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		remoteHost = host
	}
	remoteIP := net.ParseIP(remoteHost)

	// 信任条件：内网/本地回环，或在 TRUSTED_PROXIES 配置列表中
	trustProxy := remoteIP != nil && (remoteIP.IsPrivate() || remoteIP.IsLoopback() || isTrustedProxy(remoteHost))

	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ips := strings.Split(xff, ",")
			for _, ip := range ips {
				ip = strings.TrimSpace(ip)
				if parsedIP := net.ParseIP(ip); parsedIP != nil && !parsedIP.IsPrivate() {
					return ip
				}
			}
		}
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			if parsedIP := net.ParseIP(xri); parsedIP != nil && !parsedIP.IsPrivate() {
				return xri
			}
		}
	}

	return remoteHost
}
