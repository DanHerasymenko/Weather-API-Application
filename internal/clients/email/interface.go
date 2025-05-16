package email

type Client interface {
	SendEmail(to, subject, body string) error
}
