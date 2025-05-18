package config

import "fmt"

const ConfirmSubject = "Confirm your subscription"

func BuildConfirmBody(baseURL, token string) string {
	return fmt.Sprintf(
		`<p>Click <a href="%s/api/confirm/%s">here</a> to confirm your subscription.</p>`,
		baseURL, token,
	)
}
