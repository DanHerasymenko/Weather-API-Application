package email

// Client is an interface that defines the methods for sending emails
type Client interface {
	SendEmail(to, subject, body string) error
}
