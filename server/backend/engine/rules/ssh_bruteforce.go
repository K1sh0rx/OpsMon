package rules

import "strings"

func IsSSHBruteForce(logDoc map[string]interface{}) bool {

	process := strings.ToLower(getStr(logDoc, "process"))
	exe := strings.ToLower(getStr(logDoc, "exe"))
	msg := strings.ToLower(getStr(logDoc, "message"))

	// Check SSH process
	if !(strings.Contains(process, "ssh") ||
		strings.Contains(exe, "ssh")) {
		return false
	}

	// Check failed login indicators
	if strings.Contains(msg, "failed password"){
		return true
	}

	return false
}


