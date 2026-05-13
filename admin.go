package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type EventHub struct {
	mu      sync.RWMutex
	clients map[chan string]struct{}
}

var eventHub = &EventHub{clients: make(map[chan string]struct{})}

func (h *EventHub) Subscribe() chan string {
	h.mu.Lock()
	defer h.mu.Unlock()
	ch := make(chan string, 10)
	h.clients[ch] = struct{}{}
	return ch
}

func (h *EventHub) Unsubscribe(ch chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, ch)
	close(ch)
}

func (h *EventHub) Broadcast(data string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- data:
		default:
		}
	}
}

var jwtSecret = []byte(func() string {
	v := os.Getenv("JWT_SECRET")
	if v == "" {
		panic("JWT_SECRET environment variable is required")
	}
	return v
}())

func getEnvDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

type loginClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func registerAdminRoutes(router *gin.Engine) {
	admin := router.Group("/admin")
	{
		admin.GET("", adminRootRedirect)
		admin.GET("/login", adminLoginPage)
		admin.POST("/login", adminLogin)
		admin.GET("/logout", adminLogout)

		authorized := admin.Group("/")
		authorized.Use(adminAuthMiddleware())
		{
			authorized.GET("/dashboard", adminDashboardPage)
			authorized.GET("/api/stats", adminStatsAPI)
			authorized.GET("/mails", mailPoolPage)
			authorized.GET("/api/mails", listMailAPI)
			authorized.POST("/api/mails", createMailAPI)
			authorized.PUT("/api/mails/:id", updateMailAPI)
			authorized.DELETE("/api/mails/:id", deleteMailAPI)
			authorized.GET("/api/logs/:id", getLogAPI)
			authorized.GET("/logs", logPage)
			authorized.GET("/api/logs", listLogAPI)
			authorized.GET("/api/logs-count", logsCountAPI)
authorized.GET("/api/export-logs", exportLogsAPI)
			authorized.GET("/api/mail-stats", mailStatsAPI)
			authorized.POST("/api/mails/:id/test", testMailAPI)
			authorized.GET("/templates", templatePage)
			authorized.GET("/api/templates", listTemplateAPI)
			authorized.GET("/api/templates/:id", getTemplateAPI)
			authorized.POST("/api/templates", createTemplateAPI)
			authorized.PUT("/api/templates/:id", updateTemplateAPI)
			authorized.DELETE("/api/templates/:id", deleteTemplateAPI)
			authorized.GET("/audit", auditPage)
			authorized.GET("/api/audit-logs", listAuditAPI)
			authorized.GET("/api/version", versionAPI)
		}
		admin.GET("/api/events", sseAuthMiddleware(), eventsHandler)
	}
}

func sseAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("admin_token")
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		token, err := jwt.ParseWithClaims(tokenString, &loginClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}

func eventsHandler(c *gin.Context) {
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	fmt.Fprint(c.Writer, ":ok\n\n")
	flusher.Flush()

	ch := eventHub.Subscribe()
	defer eventHub.Unsubscribe(ch)

	for {
		select {
		case data, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			flusher.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}

func adminRootRedirect(c *gin.Context) {
	c.Redirect(http.StatusFound, "/admin/login")
}

func adminLoginPage(c *gin.Context) {
	tokenString, err := c.Cookie("admin_token")
	if err == nil && tokenString != "" {
		token, err := jwt.ParseWithClaims(tokenString, &loginClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})
		if err == nil && token.Valid {
			c.Redirect(http.StatusFound, "/admin/dashboard")
			return
		}
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, loginHTML)
}

func adminLogin(c *gin.Context) {
	ip := getClientIP(c.Request)
	ctx := c.Request.Context()
	now := time.Now().Unix()

	if banned, remaining := isIPBanned(ctx, ip); banned {
		minutes := remaining / 60
		if minutes < 1 {
			minutes = 1
		}
		insertAuditLog(now, ip, "", "登录", "", "IP已被封禁", "封禁")
		c.JSON(http.StatusForbidden, gin.H{
			"error": fmt.Sprintf("登录尝试过多，IP 已被封禁，请 %d 分钟后再试", minutes),
		})
		return
	}

	username := c.PostForm("username")
	password := c.PostForm("password")

	expectedUser := os.Getenv("ADMIN_USER")
	expectedPass := os.Getenv("ADMIN_PASSWORD")

	if expectedUser == "" || expectedPass == "" {
		insertAuditLog(now, ip, username, "登录", "", "系统未配置管理员账号", "失败")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	if username != expectedUser || password != expectedPass {
		_, banSecs := recordLoginFail(ctx, ip)
		if banSecs > 0 {
			minutes := banSecs / 60
			if minutes < 1 {
				minutes = 1
			}
			insertAuditLog(now, ip, username, "登录", "", fmt.Sprintf("失败次数过多，封禁%d分钟", minutes), "封禁")
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("登录尝试过多，IP 已被封禁 %d 分钟", minutes),
			})
			return
		}
		insertAuditLog(now, ip, username, "登录", "", "密码错误", "失败")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	clearLoginFail(ctx, ip)
	insertAuditLog(now, ip, username, "登录", "", "", "成功")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, loginClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	c.SetCookie("admin_token", tokenString, 86400, "/admin", "", true, true)
	c.JSON(http.StatusOK, gin.H{"success": true, "redirect": "/admin/dashboard"})
}

func adminLogout(c *gin.Context) {
	user, _ := c.Get("admin_user")
	username, _ := user.(string)
	insertAuditLog(time.Now().Unix(), getClientIP(c.Request), username, "退出登录", "", "", "成功")
	c.SetCookie("admin_token", "", -1, "/admin", "", true, true)
	c.Redirect(http.StatusFound, "/admin/login")
}

func adminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("admin_token")
		if err != nil {
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &loginClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(*loginClaims); ok {
			c.Set("admin_user", claims.Username)
		}

		c.Next()
	}
}

func adminDashboardPage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, adminHTML)
}

func mailPoolPage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, adminHTML)
}

func logPage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, adminHTML)
}

type DailyStat struct {
	Date    string `json:"date"`
	Total   int    `json:"total"`
	Success int    `json:"success"`
	Failed  int    `json:"failed"`
}

type DomainStat struct {
	Domain     string  `json:"domain"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type StatsResponse struct {
	Total       int          `json:"total"`
	Success     int          `json:"success"`
	Failed      int          `json:"failed"`
	SuccessRate float64      `json:"successRate"`
	DailyStats  []DailyStat  `json:"dailyStats"`
	DomainStats []DomainStat `json:"domainStats"`
}

func adminStatsAPI(c *gin.Context) {
	days, _ := strconv.Atoi(c.Query("days"))
	if days < 1 || days > 30 {
		days = 7
	}
	stats, err := calculateStats(c.Request.Context(), days)
	if err != nil {
		log.Printf("Stats error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func listMailAPI(c *gin.Context) {
	mails, err := listMails()
	if err != nil {
		log.Printf("List mails error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}
	for i := range mails {
		if len(mails[i].Password) > 0 {
			mails[i].Password = "******"
		}
	}
	c.JSON(http.StatusOK, mails)
}

func auditUser(c *gin.Context) string {
	user, _ := c.Get("admin_user")
	username, _ := user.(string)
	return username
}

func createMailAPI(c *gin.Context) {
	var m Mail
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	if m.Domain == "" || m.Sender == "" || m.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain, sender, password 不能为空"})
		return
	}
	if m.Port == 0 {
		m.Port = 465
	}
	if err := createMail(&m); err != nil {
		log.Printf("Create mail error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}
	insertAuditLog(time.Now().Unix(), getClientIP(c.Request), auditUser(c), "新增邮箱", fmt.Sprintf("ID:%d", m.ID), m.Domain, "成功")
	c.JSON(http.StatusOK, m)
}

func updateMailAPI(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var m Mail
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	m.ID = id
	if m.Domain == "" || m.Sender == "" || m.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain, sender, password 不能为空"})
		return
	}
	if m.Port == 0 {
		m.Port = 465
	}
	if err := updateMail(&m); err != nil {
		log.Printf("Update mail error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}
	insertAuditLog(time.Now().Unix(), getClientIP(c.Request), auditUser(c), "修改邮箱", fmt.Sprintf("ID:%d", id), m.Domain, "成功")
	c.JSON(http.StatusOK, m)
}

func deleteMailAPI(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := deleteMail(id); err != nil {
		log.Printf("Delete mail error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}
	insertAuditLog(time.Now().Unix(), getClientIP(c.Request), auditUser(c), "删除邮箱", fmt.Sprintf("ID:%d", id), "", "成功")
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func listLogAPI(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	offset, _ := strconv.Atoi(c.Query("offset"))
	if offset < 0 {
		offset = 0
	}
	keyword := c.Query("keyword")
	status := c.Query("status")
	logs, err := listLogs(limit, offset, keyword, status)
	if err != nil {
		log.Printf("List logs error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func getLogAPI(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	logEntry, err := getLogByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "日志记录不存在"})
		return
	}
	c.JSON(http.StatusOK, logEntry)
}

func exportLogsAPI(c *gin.Context) {
	keyword := c.Query("keyword")
	status := c.Query("status")
	logs, err := listLogsAll(keyword, status)
	if err != nil {
		log.Printf("Export logs error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=mail_logs_%s.csv", time.Now().Format("20060102_150405")))

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"时间", "收件人", "主题", "邮箱ID", "来源IP", "耗时(ms)", "结果"})
	for _, l := range logs {
		w.Write([]string{
			l.CreatedAt,
			l.User,
			l.Subject,
			strconv.Itoa(l.MailID),
			l.Ip,
			strconv.FormatInt(l.DurationMs, 10),
			l.Result,
		})
	}
	w.Flush()
}

func logsCountAPI(c *gin.Context) {
	keyword := c.Query("keyword")
	status := c.Query("status")
	count, err := countLogs(keyword, status)
	if err != nil {
		log.Printf("Logs count error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": count})
}

func testMailAPI(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的邮箱ID"})
		return
	}

	host, port, sender, password, err := getMailConfig(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取邮箱配置失败"})
		return
	}

	sendResult := sendEmail(host, port, sender, password, sender, "SMTP 测试邮件", "<p>这是一封测试邮件，如果您收到说明 SMTP 配置正确。</p>", "这是一封测试邮件，如果您收到说明 SMTP 配置正确。", "邮件服务")
	result := "成功"
	if sendResult != "" {
		result = "失败"
	}
	insertAuditLog(time.Now().Unix(), getClientIP(c.Request), auditUser(c), "测试发送", fmt.Sprintf("ID:%d", id), sender, result)
	if sendResult == "" {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "测试邮件发送成功"})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": sendResult})
	}
}

func auditPage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, adminHTML)
}

func listAuditAPI(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	offset, _ := strconv.Atoi(c.Query("offset"))
	if offset < 0 {
		offset = 0
	}
	logs, err := listAuditLogs(limit, offset)
	if err != nil {
		log.Printf("List audit logs error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}
	c.JSON(http.StatusOK, logs)
}

func templatePage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, adminHTML)
}

func listTemplateAPI(c *gin.Context) {
	templates, err := listTemplates()
	if err != nil {
		log.Printf("List templates error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}
	c.JSON(http.StatusOK, templates)
}

func getTemplateAPI(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	t, err := getTemplate(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "模板不存在"})
		return
	}
	c.JSON(http.StatusOK, t)
}

func createTemplateAPI(c *gin.Context) {
	var t MailTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	if t.Name == "" || t.Subject == "" || t.Body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, subject, body 不能为空"})
		return
	}
	if err := createTemplate(&t); err != nil {
		log.Printf("Create template error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}
	insertAuditLog(time.Now().Unix(), getClientIP(c.Request), auditUser(c), "新增模板", fmt.Sprintf("ID:%d", t.ID), t.Name, "成功")
	c.JSON(http.StatusOK, t)
}

func updateTemplateAPI(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var t MailTemplate
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}
	t.ID = id
	if t.Name == "" || t.Subject == "" || t.Body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, subject, body 不能为空"})
		return
	}
	if err := updateTemplate(&t); err != nil {
		log.Printf("Update template error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}
	insertAuditLog(time.Now().Unix(), getClientIP(c.Request), auditUser(c), "修改模板", fmt.Sprintf("ID:%d", id), t.Name, "成功")
	c.JSON(http.StatusOK, t)
}

func deleteTemplateAPI(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := deleteTemplate(id); err != nil {
		log.Printf("Delete template error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}
	insertAuditLog(time.Now().Unix(), getClientIP(c.Request), auditUser(c), "删除模板", fmt.Sprintf("ID:%d", id), "", "成功")
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func versionAPI(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"version": version})
}

func mailStatsAPI(c *gin.Context) {
	days, _ := strconv.Atoi(c.Query("days"))
	if days < 1 || days > 30 {
		days = 7
	}
	since := time.Now().Add(-time.Duration(days) * 24 * time.Hour).Unix()

	rows, err := db.Query(`
		SELECT m.id, m.sender, m.domain,
		       COUNT(*) as total,
		       SUM(CASE WHEN ml.result LIKE '%Message has been sent%' THEN 1 ELSE 0 END) as success,
		       SUM(CASE WHEN ml.result NOT LIKE '%Message has been sent%' THEN 1 ELSE 0 END) as failed
		FROM mail m
		LEFT JOIN mail_log ml ON ml.mail_id = m.id AND ml.time_unix >= ?
		GROUP BY m.id, m.sender, m.domain
		ORDER BY m.id`, since)
	if err != nil {
		log.Printf("Mail stats error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}
	defer rows.Close()

	type MailStat struct {
		ID      int    `json:"id"`
		Sender  string `json:"sender"`
		Domain  string `json:"domain"`
		Total   int    `json:"total"`
		Success int    `json:"success"`
		Failed  int    `json:"failed"`
		Rate    string `json:"rate"`
	}

	stats := make([]MailStat, 0)
	for rows.Next() {
		var s MailStat
		if err := rows.Scan(&s.ID, &s.Sender, &s.Domain, &s.Total, &s.Success, &s.Failed); err != nil {
			log.Printf("Mail stats scan error: %v", err)
			continue
		}
		if s.Total > 0 {
			s.Rate = fmt.Sprintf("%.1f%%", float64(s.Success)/float64(s.Total)*100)
		} else {
			s.Rate = "-"
		}
		stats = append(stats, s)
	}
	c.JSON(http.StatusOK, stats)
}
