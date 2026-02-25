package rules

import "strings"

func IsWebCommandInjection(logDoc map[string]interface{}) bool {

	msg := strings.ToLower(getStr(logDoc, "message"))

	// Detect command injection payloads
	payloads := []string{
		"cmd",
		"exec",
		"system",
		"bash",
		"nc",
		"wget",
		"curl",
		"whoami",
		"passwd",
		"php",
		"select",
	}

	for _, p := range payloads {
		if strings.Contains(msg, p) {
			return true
		}
	}

	return false
}

func getStr(m map[string]interface{}, k string) string {
	if v, ok := m[k].(string); ok {
		return v
	}
	return ""
}
