package client

import (
	"Weather-API-Application/internal/config"
	"Weather-API-Application/internal/logger"
	"Weather-API-Application/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
)

type EmailClient struct {
	From     string
	Password string
	Host     string
	Port     string
}

func NewEmailClient(cfg *config.Config) *EmailClient {
	return &EmailClient{From: cfg.EmailClientFrom, Password: cfg.EmailClientPassword, Host: cfg.EmailClientHost, Port: cfg.EmailClientPort}
}

// Client is an interface that defines the methods for sending emails
type Client interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

// SendEmail sends an email_service using the SMTP client
func (c *EmailClient) SendEmail(ctx context.Context, to, subject, body string) error {
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
		"\r\n" + body)

	auth := smtp.PlainAuth("", c.From, c.Password, c.Host)

	err := smtp.SendMail(c.Host+":"+c.Port, auth, c.From, []string{to}, msg)
	if err != nil {
		return err
	}
	logger.Info(ctx, "Email sent successfully to "+to)

	return nil
}

// sendUpdate fetches the current weather data for the given subscription's city,
// formats it into a plain text message, and sends it via email_client to the subscriber.
// Returns an error if any of the steps fail (HTTP request, JSON parsing, or email_service sending).
func SendUpdate(ctx context.Context, apiKey string, sub *model.Subscription, emailClient EmailClient) error {

	// Fetch the weather data for the city
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

	// Prepare the email_service content
	weatherMailText := fmt.Sprintf(`Weather for %s:<br>- temperature: %.1f°C<br>- humidity: %.0f%%<br>- description: %s`,
		sub.City, weatherApiResp.Current.TempC, weatherApiResp.Current.Humidity, weatherApiResp.Current.Condition.Text)
	subject := fmt.Sprintf("%s forecast", sub.City)

	// Send the email_service
	if err := emailClient.SendEmail(ctx, sub.Email, subject, weatherMailText); err != nil {
		return fmt.Errorf("failed to send email_service to %s for city %s: %w", sub.Email, sub.City, err)
	}

	return nil
}
