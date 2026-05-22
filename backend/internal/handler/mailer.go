package handler

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"blogapp/backend/internal/config"
)

type verificationMailer struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewVerificationMailer(cfg config.Config) *verificationMailer {
	host := strings.TrimSpace(cfg.SMTPHost)
	username := strings.TrimSpace(cfg.SMTPUsername)
	password := strings.TrimSpace(cfg.SMTPPassword)
	if host == "" || username == "" || password == "" {
		return nil
	}
	port := strings.TrimSpace(cfg.SMTPPort)
	if port == "" {
		port = "465"
	}
	from := strings.TrimSpace(cfg.SMTPFrom)
	if from == "" {
		from = username
	}
	return &verificationMailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (m *verificationMailer) SendVerificationCode(to, purpose, code string) error {
	if m == nil || m.host == "" || m.username == "" || m.password == "" || m.from == "" {
		return errors.New("smtp is not configured")
	}
	label := map[string]string{
		"login":    "登录",
		"register": "注册",
	}[purpose]
	if label == "" {
		label = "验证"
	}
	body := fmt.Sprintf("你好：\n\n你的 xiaoli 博客%s验证码是：%s\n验证码 5 分钟内有效，请尽快使用。\n\n如果不是你本人操作，请忽略这封邮件。\n", label, code)
	message := buildVerificationMessage(m.from, to, body)
	addr := net.JoinHostPort(m.host, m.port)
	if m.port == "465" {
		return m.sendImplicitTLS(addr, to, message)
	}
	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	return smtp.SendMail(addr, auth, m.from, []string{to}, message)
}

func (m *verificationMailer) sendImplicitTLS(addr, to string, message []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: m.host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		return err
	}
	defer client.Close()

	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	if err := client.Auth(auth); err != nil {
		return err
	}
	if err := client.Mail(m.from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func buildVerificationMessage(from, to, body string) []byte {
	headers := []string{
		"From: Xiaoli Blog <" + from + ">",
		"To: " + to,
		"Subject: Xiaoli Blog verification code",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
	}
	return []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + body)
}
