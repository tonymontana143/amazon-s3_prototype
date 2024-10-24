package handlers

import (
	"net"
	"regexp"
)

func ValidateBucketName(s string) bool {
	if len(s) < 3 || len(s) > 63 {
		return false
	}

	regex := `^[a-z0-9]([a-z0-9.-]{1,61}[a-z0-9])?$`
	re := regexp.MustCompile(regex)

	if !re.MatchString(s) {
		return false
	}

	for i := 0; i < len(s)-1; i++ {
		if (s[i] == '.' && s[i+1] == '.') || (s[i] == '-' && s[i+1] == '-') {
			return false
		}
	}

	if net.ParseIP(s) != nil {
		return false
	}

	return true
}
