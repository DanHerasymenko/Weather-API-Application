package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/smtp"

	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/logger"
	"Weather-API-Application/internal/model"
)

// SmtpSender abstracts smtp.SendMail for testability.
type SmtpSender interface {
	SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

type smtpSender struct{}

func (smtpSender) SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	return smtp.SendMail(addr, a, from, to, msg)
}

type EmailClient struct {
	From     string
	Password string
	Host     string
	Port     string
	sender   SmtpSender
}

func NewEmailClient(cfg *config.Config) *EmailClient {
	return &EmailClient{From: cfg.EmailClientFrom, Password: cfg.EmailClientPassword, Host: cfg.EmailClientHost, Port: cfg.EmailClientPort, sender: smtpSender{}}
}

// NewEmailClientWithSender allows injecting a custom SmtpSender (useful for tests).
func NewEmailClientWithSender(cfg *config.Config, sender SmtpSender) *EmailClient {
	return &EmailClient{From: cfg.EmailClientFrom, Password: cfg.EmailClientPassword, Host: cfg.EmailClientHost, Port: cfg.EmailClientPort, sender: sender}
}

// Client defines methods for sending emails (used by services).
type Client interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

// SendEmail sends an email using SMTP.
func (c *EmailClient) SendEmail(ctx context.Context, to, subject, body string) error {
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
		"\r\n" + body)

	auth := smtp.PlainAuth("", c.From, c.Password, c.Host)

	if err := c.sender.SendMail(c.Host+":"+c.Port, auth, c.From, []string{to}, msg); err != nil {
		return err
	}

	logger.Info(ctx, "Email sent successfully",
		slog.String("to", to),
		slog.String("subject", subject))
	return nil
}

// SendUpdate fetches current weather for the subscription city and emails the user.
func SendUpdate(ctx context.Context, apiKey string, sub *model.Subscription, emailClient Client) error {
	if apiKey == "" {
		return fmt.Errorf("weather API key is missing in config")
	}

	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s&aqi=no", apiKey, sub.City)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("invalid request: failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("city not found: failed to fetch weather data: %s", resp.Status)
	}

	var weatherApiResp model.WeatherAPIResponse
	if err = json.NewDecoder(resp.Body).Decode(&weatherApiResp); err != nil {
		return fmt.Errorf("failed to decode weather data: %w", err)
	}

	weatherMailText := fmt.Sprintf(`Weather for %s:<br>- temperature: %.1f°C<br>- humidity: %.0f%%<br>- description: %s`,
		sub.City, weatherApiResp.Current.TempC, weatherApiResp.Current.Humidity, weatherApiResp.Current.Condition.Text)
	subject := fmt.Sprintf("%s forecast", sub.City)

	if err := emailClient.SendEmail(ctx, sub.Email, subject, weatherMailText); err != nil {
		return fmt.Errorf("failed to send email to %s for city %s: %w", sub.Email, sub.City, err)
	}
	return nil
}
