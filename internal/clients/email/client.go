package email

import (
	"Weather-API-Application/internal/logger"
	"context"
	"net/smtp"
)

type SMTPClient struct {
	From     string
	Password string
	Host     string
	Port     string
}

func NewSMTPClient(from, password, host, port string) *SMTPClient {
	return &SMTPClient{From: from, Password: password, Host: host, Port: port}
}

// SendEmail sends an email using the SMTP client
func (c *SMTPClient) SendEmail(ctx context.Context, to, subject, body string) error {
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
		"\r\n" + body)

	auth := smtp.PlainAuth("", c.From, c.Password, c.Host)

	// Send the email
	err := smtp.SendMail(c.Host+":"+c.Port, auth, c.From, []string{to}, msg)
	if err != nil {
		return err
	}
	logger.Info(ctx, "Email sent successfully to "+to)

	return nil
}
