package config



subject := "Confirm your subscription"
body := fmt.Sprintf(<p>Click <a href="%s/api/subscription/confirm/%s">here</a> to confirm your subscription.</p>, c.BaseURL, token)

msg := []byte("To: " + to + "\r\n" +
"Subject: " + subject + "\r\n" +
"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
"\r\n" + body)
