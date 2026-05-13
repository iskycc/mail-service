package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
)

// 脱敏 4-6 位连续数字 / 纯大写字母 / 纯小写字母（验证码常见格式）
// 跳过 HTML 标签内的内容（<...> 之间），避免破坏 HTML 结构
var sanitizeRegex = regexp.MustCompile(`\d{4,6}|[A-Z]{4,6}|[a-z]{4,6}`)

func sanitizeSensitive(text string) string {
	// 检测是否包含 HTML 标签
	if !strings.Contains(text, "<") {
		return sanitizeRegex.ReplaceAllString(text, "****")
	}
	// 分段处理：HTML 标签原样保留，标签外内容做脱敏
	var result strings.Builder
	inTag := false
	buf := strings.Builder{}
	for _, r := range text {
		if r == '<' {
			if buf.Len() > 0 {
				result.WriteString(sanitizeRegex.ReplaceAllString(buf.String(), "****"))
				buf.Reset()
			}
			inTag = true
			buf.WriteRune(r)
		} else if r == '>' {
			buf.WriteRune(r)
			result.WriteString(buf.String())
			buf.Reset()
			inTag = false
		} else {
			buf.WriteRune(r)
		}
	}
	if buf.Len() > 0 {
		if inTag {
			result.WriteString(buf.String())
		} else {
			result.WriteString(sanitizeRegex.ReplaceAllString(buf.String(), "****"))
		}
	}
	return result.String()
}

type Mail struct {
	ID       int    `json:"id"`
	Domain   string `json:"domain"`
	Port     int    `json:"port"`
	Sender   string `json:"sender"`
	Password string `json:"password"`
}

type MailLog struct {
	ID         int64  `json:"id"`
	TimeUnix   int64  `json:"time_unix"`
	Ip         string `json:"ip"`
	User       string `json:"user"`
	Subject    string `json:"subject"`
	Body       string `json:"body"`
	Altbody    string `json:"altbody"`
	TeamName   string `json:"team_name"`
	MailID     int    `json:"mail_id"`
	Result     string `json:"result"`
	DurationMs int64  `json:"duration_ms"`
	CreatedAt  string `json:"created_at"`
}

func tableExists(name string) (bool, error) {
	var dummy string
	query := fmt.Sprintf("SHOW TABLES LIKE '%s'", name)
	err := db.QueryRow(query).Scan(&dummy)
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return false, err
}

func ensureTable(name, createSQL string) error {
	exists, err := tableExists(name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = db.Exec(createSQL)
	return err
}

func initMailTable() error {
	query := `CREATE TABLE mail (
		id INT AUTO_INCREMENT PRIMARY KEY,
		domain VARCHAR(255) NOT NULL,
		port INT NOT NULL DEFAULT 465,
		sender VARCHAR(255) NOT NULL,
		password VARCHAR(255) NOT NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
	return ensureTable("mail", query)
}

func initMailLogTable() error {
	query := `CREATE TABLE mail_log (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		time_unix BIGINT NOT NULL,
		ip VARCHAR(64),
		user VARCHAR(255),
		subject VARCHAR(500),
		body TEXT,
		altbody TEXT,
		team_name VARCHAR(100),
		mail_id INT,
		result VARCHAR(500),
		duration_ms BIGINT DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_time_unix (time_unix),
		INDEX idx_created_at (created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
	if err := ensureTable("mail_log", query); err != nil {
		return err
	}
	// 兼容已有表：尝试添加 duration_ms 列（已存在则忽略错误）
	_, _ = db.Exec("ALTER TABLE mail_log ADD COLUMN duration_ms BIGINT DEFAULT 0")
	return nil
}

type AuditLog struct {
	ID        int64  `json:"id"`
	TimeUnix  int64  `json:"time_unix"`
	Ip        string `json:"ip"`
	Username  string `json:"username"`
	Action    string `json:"action"`
	Target    string `json:"target"`
	Detail    string `json:"detail"`
	Result    string `json:"result"`
	CreatedAt string `json:"created_at"`
}

func initAuditLogTable() error {
	query := `CREATE TABLE audit_log (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		time_unix BIGINT NOT NULL,
		ip VARCHAR(64),
		username VARCHAR(100),
		action VARCHAR(50),
		target VARCHAR(255),
		detail TEXT,
		result VARCHAR(50),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_time_unix (time_unix),
		INDEX idx_username (username),
		INDEX idx_action (action)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
	return ensureTable("audit_log", query)
}

type MailTemplate struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

func initMailTemplateTable() error {
	query := `CREATE TABLE mail_template (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		subject VARCHAR(500) NOT NULL,
		body TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
	return ensureTable("mail_template", query)
}

func listTemplates() ([]MailTemplate, error) {
	rows, err := db.Query("SELECT id, name, subject, body, created_at FROM mail_template ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]MailTemplate, 0)
	for rows.Next() {
		var t MailTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.Subject, &t.Body, &t.CreatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

func getTemplate(id int) (*MailTemplate, error) {
	var t MailTemplate
	err := db.QueryRow("SELECT id, name, subject, body, created_at FROM mail_template WHERE id = ?", id).Scan(
		&t.ID, &t.Name, &t.Subject, &t.Body, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func createTemplate(t *MailTemplate) error {
	result, err := db.Exec("INSERT INTO mail_template (name, subject, body) VALUES (?, ?, ?)",
		t.Name, t.Subject, t.Body)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	t.ID = int(id)
	return nil
}

func updateTemplate(t *MailTemplate) error {
	_, err := db.Exec("UPDATE mail_template SET name=?, subject=?, body=? WHERE id=?",
		t.Name, t.Subject, t.Body, t.ID)
	return err
}

func deleteTemplate(id int) error {
	_, err := db.Exec("DELETE FROM mail_template WHERE id=?", id)
	return err
}

func insertAuditLog(timeUnix int64, ip, username, action, target, detail, result string) {
	query := `INSERT INTO audit_log (time_unix, ip, username, action, target, detail, result)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, timeUnix, ip, username, action, target, detail, result)
	if err != nil {
		log.Printf("Failed to insert audit log: %v", err)
	}
}

func listAuditLogs(limit, offset int) ([]AuditLog, error) {
	query := `SELECT id, time_unix, ip, username, action, target, detail, result, created_at FROM audit_log
			  ORDER BY id DESC LIMIT ? OFFSET ?`
	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]AuditLog, 0)
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(&l.ID, &l.TimeUnix, &l.Ip, &l.Username, &l.Action, &l.Target, &l.Detail, &l.Result, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func insertMailLog(ctx context.Context, rec RecordParams) error {
	query := `INSERT INTO mail_log (time_unix, ip, user, subject, body, altbody, team_name, mail_id, result, duration_ms)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.ExecContext(ctx, query,
		rec.TimeUnix, rec.Ip, rec.User,
		sanitizeSensitive(rec.Subject),
		sanitizeSensitive(rec.Body),
		sanitizeSensitive(rec.Altbody),
		rec.Tname, rec.MailID, rec.Result, rec.DurationMs)
	return err
}

func listMails() ([]Mail, error) {
	rows, err := db.Query("SELECT id, domain, port, sender, password FROM mail ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mails := make([]Mail, 0)
	for rows.Next() {
		var m Mail
		if err := rows.Scan(&m.ID, &m.Domain, &m.Port, &m.Sender, &m.Password); err != nil {
			return nil, err
		}
		mails = append(mails, m)
	}
	return mails, rows.Err()
}

func createMail(m *Mail) error {
	result, err := db.Exec("INSERT INTO mail (domain, port, sender, password) VALUES (?, ?, ?, ?)",
		m.Domain, m.Port, m.Sender, m.Password)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	m.ID = int(id)
	rdb.Del(context.Background(), "mail_count")
	return nil
}

func updateMail(m *Mail) error {
	_, err := db.Exec("UPDATE mail SET domain=?, port=?, sender=?, password=? WHERE id=?",
		m.Domain, m.Port, m.Sender, m.Password, m.ID)
	if err != nil {
		return err
	}
	cacheKey := fmt.Sprintf("mailconfig:%d", m.ID)
	if err := rdb.Del(context.Background(), cacheKey).Err(); err != nil {
		log.Printf("Failed to clear redis cache for %s: %v", cacheKey, err)
	}
	return nil
}

func deleteMail(id int) error {
	_, err := db.Exec("DELETE FROM mail WHERE id=?", id)
	if err != nil {
		return err
	}
	cacheKey := fmt.Sprintf("mailconfig:%d", id)
	if err := rdb.Del(context.Background(), cacheKey).Err(); err != nil {
		log.Printf("Failed to clear redis cache for %s: %v", cacheKey, err)
	}
	rdb.Del(context.Background(), "mail_count")
	return nil
}

func listLogs(limit, offset int, keyword, status string) ([]MailLog, error) {
	var query string
	var args []interface{}
	hasKeyword := keyword != ""
	hasStatus := status == "success" || status == "failed"
	if hasKeyword && hasStatus {
		kw := "%" + keyword + "%"
		if status == "success" {
			query = `SELECT id, time_unix, ip, user, subject, body, altbody, team_name, mail_id, result, duration_ms, created_at FROM mail_log
				 WHERE (user LIKE ? OR subject LIKE ? OR ip LIKE ? OR result LIKE ?) AND result LIKE '%Message has been sent%'
				 ORDER BY id DESC LIMIT ? OFFSET ?`
		} else {
			query = `SELECT id, time_unix, ip, user, subject, body, altbody, team_name, mail_id, result, duration_ms, created_at FROM mail_log
				 WHERE (user LIKE ? OR subject LIKE ? OR ip LIKE ? OR result LIKE ?) AND result NOT LIKE '%Message has been sent%'
				 ORDER BY id DESC LIMIT ? OFFSET ?`
		}
		args = append(args, kw, kw, kw, kw, limit, offset)
	} else if hasKeyword {
		kw := "%" + keyword + "%"
		query = `SELECT id, time_unix, ip, user, subject, body, altbody, team_name, mail_id, result, duration_ms, created_at FROM mail_log
				 WHERE user LIKE ? OR subject LIKE ? OR ip LIKE ? OR result LIKE ?
				 ORDER BY id DESC LIMIT ? OFFSET ?`
		args = append(args, kw, kw, kw, kw, limit, offset)
	} else if hasStatus {
		if status == "success" {
			query = `SELECT id, time_unix, ip, user, subject, body, altbody, team_name, mail_id, result, duration_ms, created_at FROM mail_log
				 WHERE result LIKE '%Message has been sent%' ORDER BY id DESC LIMIT ? OFFSET ?`
		} else {
			query = `SELECT id, time_unix, ip, user, subject, body, altbody, team_name, mail_id, result, duration_ms, created_at FROM mail_log
				 WHERE result NOT LIKE '%Message has been sent%' ORDER BY id DESC LIMIT ? OFFSET ?`
		}
		args = append(args, limit, offset)
	} else {
		query = "SELECT id, time_unix, ip, user, subject, body, altbody, team_name, mail_id, result, duration_ms, created_at FROM mail_log ORDER BY id DESC LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]MailLog, 0)
	for rows.Next() {
		var l MailLog
		if err := rows.Scan(&l.ID, &l.TimeUnix, &l.Ip, &l.User, &l.Subject, &l.Body, &l.Altbody, &l.TeamName, &l.MailID, &l.Result, &l.DurationMs, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func listLogsAll(keyword, status string) ([]MailLog, error) {
	var query string
	var args []interface{}
	hasKeyword := keyword != ""
	hasStatus := status == "success" || status == "failed"
	if hasKeyword && hasStatus {
		kw := "%" + keyword + "%"
		if status == "success" {
			query = `SELECT id, time_unix, ip, user, subject, body, altbody, team_name, mail_id, result, duration_ms, created_at FROM mail_log
				 WHERE (user LIKE ? OR subject LIKE ? OR ip LIKE ? OR result LIKE ?) AND result LIKE '%Message has been sent%'`
		} else {
			query = `SELECT id, time_unix, ip, user, subject, body, altbody, team_name, mail_id, result, duration_ms, created_at FROM mail_log
				 WHERE (user LIKE ? OR subject LIKE ? OR ip LIKE ? OR result LIKE ?) AND result NOT LIKE '%Message has been sent%'`
		}
		args = append(args, kw, kw, kw, kw)
	} else if hasKeyword {
		kw := "%" + keyword + "%"
		query = `SELECT id, time_unix, ip, user, subject, body, altbody, team_name, mail_id, result, duration_ms, created_at FROM mail_log
				 WHERE user LIKE ? OR subject LIKE ? OR ip LIKE ? OR result LIKE ?`
		args = append(args, kw, kw, kw, kw)
	} else if hasStatus {
		if status == "success" {
			query = `SELECT id, time_unix, ip, user, subject, body, altbody, team_name, mail_id, result, duration_ms, created_at FROM mail_log
				 WHERE result LIKE '%Message has been sent%'`
		} else {
			query = `SELECT id, time_unix, ip, user, subject, body, altbody, team_name, mail_id, result, duration_ms, created_at FROM mail_log
				 WHERE result NOT LIKE '%Message has been sent%'`
		}
	} else {
		query = "SELECT id, time_unix, ip, user, subject, body, altbody, team_name, mail_id, result, duration_ms, created_at FROM mail_log"
	}
	query += " ORDER BY id DESC"
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]MailLog, 0)
	for rows.Next() {
		var l MailLog
		if err := rows.Scan(&l.ID, &l.TimeUnix, &l.Ip, &l.User, &l.Subject, &l.Body, &l.Altbody, &l.TeamName, &l.MailID, &l.Result, &l.DurationMs, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
func countLogs(keyword, status string) (int, error) {
	var query string
	var args []interface{}
	hasKeyword := keyword != ""
	hasStatus := status == "success" || status == "failed"
	if hasKeyword && hasStatus {
		kw := "%" + keyword + "%"
		if status == "success" {
			query = `SELECT COUNT(*) FROM mail_log
				 WHERE (user LIKE ? OR subject LIKE ? OR ip LIKE ? OR result LIKE ?) AND result LIKE '%Message has been sent%'`
		} else {
			query = `SELECT COUNT(*) FROM mail_log
				 WHERE (user LIKE ? OR subject LIKE ? OR ip LIKE ? OR result LIKE ?) AND result NOT LIKE '%Message has been sent%'`
		}
		args = append(args, kw, kw, kw, kw)
	} else if hasKeyword {
		kw := "%" + keyword + "%"
		query = `SELECT COUNT(*) FROM mail_log
				 WHERE user LIKE ? OR subject LIKE ? OR ip LIKE ? OR result LIKE ?`
		args = append(args, kw, kw, kw, kw)
	} else if hasStatus {
		if status == "success" {
			query = `SELECT COUNT(*) FROM mail_log
				 WHERE result LIKE '%Message has been sent%'`
		} else {
			query = `SELECT COUNT(*) FROM mail_log
				 WHERE result NOT LIKE '%Message has been sent%'`
		}
	} else {
		query = "SELECT COUNT(*) FROM mail_log"
	}
	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	return count, err
}

func getLogByID(id int) (*MailLog, error) {
	var l MailLog
	err := db.QueryRow("SELECT id, time_unix, ip, user, subject, body, altbody, team_name, mail_id, result, duration_ms, created_at FROM mail_log WHERE id = ?", id).Scan(
		&l.ID, &l.TimeUnix, &l.Ip, &l.User, &l.Subject, &l.Body, &l.Altbody, &l.TeamName, &l.MailID, &l.Result, &l.DurationMs, &l.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &l, nil
}
