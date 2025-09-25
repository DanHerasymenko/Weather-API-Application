package validate

import (
	"regexp"
	"strings"
)

func IsValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func IsValidCity(city string) bool {
	return strings.TrimSpace(city) != ""
}

func IsValidFrequency(frequency string) bool {
	freq := strings.ToLower(strings.TrimSpace(frequency))
	return freq == "hourly" || freq == "daily"
}
