package sender

import (
	"net"
	"time"
)

// checkServerReachable performs TCP reachability check
// before retrying held batch after SOC backpressure
func checkServerReachable() bool {

	conn, err := net.DialTimeout(
		"tcp",
		baseURLHost()+":8080",
		3*time.Second,
	)

	if err != nil {
		return false
	}

	conn.Close()
	return true
}
