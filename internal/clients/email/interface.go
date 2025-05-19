package email

import "context"

// Client is an interface that defines the methods for sending emails
type Client interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}
