package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/gomail.v2"
)

var (
	rdb      *redis.Client
	mysqlDSN string
)

var updateMailIDScript = redis.NewScript(`
local mailid = redis.call('GET', KEYS[1])
if not mailid then
    mailid = 1
    redis.call('SET', KEYS[1], mailid)
else
    mailid = tonumber(mailid)
    if mailid == 5 then
        mailid = 1
    else
        mailid = mailid + 1
    end
    redis.call('SET', KEYS[1], mailid)
end
return mailid
`)

func main() {
	// 初始化环境变量
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

	// 初始化Redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       0,
	})

	// 启动HTTP服务器
	http.HandleFunc("/", handler)
	log.Println("Server starting on :22125...")
	log.Fatal(http.ListenAndServe(":22125", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	// 允许的方法检查
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := context.Background()
	var user, subject, body, altbody, tname string
	// 参数获取逻辑
	switch r.Method {
	case http.MethodGet:
		// GET 请求：仅从URL获取参数
		params := r.URL.Query()
		user = params.Get("user")
		subject = params.Get("subject")
		body = params.Get("body")
		altbody = params.Get("altbody")
		tname = params.Get("tname")
	case http.MethodPost:
		// POST 请求：根据Content-Type处理
		contentType := r.Header.Get("Content-Type")
		// 处理JSON请求
		if strings.HasPrefix(contentType, "application/json") {
			var data map[string]string
			if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
				http.Error(w, "Invalid JSON format", http.StatusBadRequest)
				return
			}
			user = data["user"]
			subject = data["subject"]
			body = data["body"]
			altbody = data["altbody"]
			tname = data["tname"]
		} else {
			// 处理表单数据（兼容所有类型）
			if strings.HasPrefix(contentType, "multipart/form-data") {
				// 解析multipart表单（支持文件上传）
				if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB内存，其余存临时文件
					http.Error(w, "Multipart parse error", http.StatusBadRequest)
					return
				}
			} else {
				// 解析普通表单
				if err := r.ParseForm(); err != nil {
					http.Error(w, "Form parse error", http.StatusBadRequest)
					return
				}
			}
			// 统一从POST BODY获取参数（不包含URL参数）
			postForm := r.PostForm
			user = postForm.Get("user")
			subject = postForm.Get("subject")
			body = postForm.Get("body")
			altbody = postForm.Get("altbody")
			tname = postForm.Get("tname")
		}
	}

	// 处理默认值
	if tname == "" {
		tname = "Libv 团队"
	}
	if altbody == "" {
		altbody = body
	}

	// 获取客户端IP
	ip := getClientIP(r)

	maxRetries := 3
	var success bool
	var resultInfo string

	for i := 0; i < maxRetries; i++ {
		mailID, err := updateMailID(ctx)
		if err != nil {
			resultInfo = fmt.Sprintf("Redis error: %v", err)
			// 新增Redis错误日志
			log.Printf("[发送失败] 来源IP:%s | 错误类型:Redis | 详情:%v", ip, err)
			continue
		}
		// 获取邮件配置
		host, port, sender, password, err := getMailConfig(mailID)
		if err != nil {
			resultInfo = fmt.Sprintf("MySQL error: %v", err)
			saveRecord(ctx, ip, user, subject, body, altbody, tname, mailID, resultInfo)
			// 新增MySQL错误日志
			log.Printf("[发送失败] 来源IP:%s | 邮件ID:%d | 错误类型:MySQL | 收件人:%s | 详情:%v",
				ip, mailID, user, err)
			continue
		}
		// 发送邮件
		sendResult := sendEmail(host, port, sender, password, user, subject, body, altbody, tname)
		if sendResult == "" {
			success = true
			resultInfo = "Message has been sent"
			// 新增成功日志
			log.Printf("[发送成功] 时间:%s | 来源IP:%s | 邮件ID:%d | 发件人:%s | 收件人:%s",
				time.Now().Format(time.RFC3339), ip, mailID, sender, user)
		} else {
			resultInfo = sendResult
			// 新增失败日志
			log.Printf("[发送失败] 时间:%s | 来源IP:%s | 邮件ID:%d | 发件人:%s | 收件人:%s | 错误:%s",
				time.Now().Format(time.RFC3339), ip, mailID, sender, user, sendResult)
		}
		saveRecord(ctx, ip, user, subject, body, altbody, tname, mailID, resultInfo)
		if success {
			break
		}
	}

	// 返回响应
	response := map[string]interface{}{
		"success": success,
		"info":    resultInfo,
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		return
	}
}

func updateMailID(ctx context.Context) (int, error) {
	return updateMailIDScript.Run(ctx, rdb, []string{"mailid"}).Int()
}

func getMailConfig(mailID int) (string, int, string, string, error) {
	db, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		return "", 0, "", "", fmt.Errorf("MySQL connection failed: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

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
	return host, port, sender, password, nil
}

func sendEmail(host string, port int, sender, password, to, subject, body, altbody, tname string) string {
	m := gomail.NewMessage()

	// 编码发件人名称
	encodedTname := "=?utf-8?B?" + base64.StdEncoding.EncodeToString([]byte(tname)) + "?="
	m.SetHeader("From", m.FormatAddress(sender, encodedTname))
	m.SetHeader("To", to)

	// 编码主题
	encodedSubject := "=?utf-8?B?" + base64.StdEncoding.EncodeToString([]byte(subject)) + "?="
	m.SetHeader("Subject", encodedSubject)

	m.SetBody("text/html", body)
	m.AddAlternative("text/plain", altbody)

	dialer := gomail.NewDialer(host, port, sender, password)
	dialer.SSL = true

	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Sprintf("Message could not be sent. Mailer Error: %v", err)
	}
	return ""
}

func saveRecord(ctx context.Context, ip, user, subject, body, altbody, tname string, mailID int, result string) {
	now := time.Now().Unix()
	key := strconv.FormatInt(now, 10)

	fields := map[string]interface{}{
		"ip":        ip,
		"user":      user,
		"subject":   subject,
		"body":      body,
		"altbody":   altbody,
		"team_name": tname,
		"mailid":    mailID,
		"result":    result,
	}

	if err := rdb.HMSet(ctx, key, fields).Err(); err != nil {
		log.Printf("Failed to save record: %v", err)
		return
	}

	if err := rdb.Expire(ctx, key, 7*24*time.Hour).Err(); err != nil {
		log.Printf("Failed to set expiration: %v", err)
	}
}

// 新增函数：获取客户端真实IP
func getClientIP(r *http.Request) string {
	// 1. 优先从X-Forwarded-For获取（可能有多个IP，取第一个有效IP）
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if parsedIP := net.ParseIP(ip); parsedIP != nil && !parsedIP.IsPrivate() {
				return ip
			}
		}
	}
	// 2. 检查X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if parsedIP := net.ParseIP(xri); parsedIP != nil && !parsedIP.IsPrivate() {
			return xri
		}
	}
	// 3. 最后从RemoteAddr获取
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // 当没有端口时直接返回
	}
	return ip
}
