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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()

	// 解析表单
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// 获取参数
	user := r.PostForm.Get("user")
	subject := r.PostForm.Get("subject")
	body := r.PostForm.Get("body")
	altbody := r.PostForm.Get("altbody")
	tname := r.PostForm.Get("tname")

	// 处理默认值
	if tname == "" {
		tname = "喵滴团队"
	}
	if altbody == "" {
		altbody = body
	}

	// 获取客户端IP
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}

	maxRetries := 3
	var success bool
	var resultInfo string

	for i := 0; i < maxRetries; i++ {
		mailID, err := updateMailID(ctx)
		if err != nil {
			resultInfo = fmt.Sprintf("Redis error: %v", err)
			continue
		}

		// 获取邮件配置
		host, port, sender, password, err := getMailConfig(mailID)
		if err != nil {
			resultInfo = fmt.Sprintf("MySQL error: %v", err)
			saveRecord(ctx, ip, user, subject, body, altbody, tname, mailID, resultInfo)
			continue
		}

		// 发送邮件
		sendResult := sendEmail(host, port, sender, password, user, subject, body, altbody, tname)
		if sendResult == "" {
			success = true
			resultInfo = "Message has been sent"
		} else {
			resultInfo = sendResult
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
	json.NewEncoder(w).Encode(response)
}

func updateMailID(ctx context.Context) (int, error) {
	return updateMailIDScript.Run(ctx, rdb, []string{"mailid"}).Int()
}

func getMailConfig(mailID int) (string, int, string, string, error) {
	db, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		return "", 0, "", "", fmt.Errorf("MySQL connection failed: %v", err)
	}
	defer db.Close()

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
