package loadbalancer

import (
	"net"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
)

// HealthCheckRun runs a routine for check status of the servers every 2 mins
func HealthCheckRun(sm *ServerManager) {
	t := time.NewTicker(time.Minute * 2)
	for {
		select {
		case <-t.C:
			log.Info("Starting health check...")
			runAll(sm)
			log.Info("Health check completed")
		}
	}
}

func runAll(sm *ServerManager) {
	for _, s := range sm.pool.Registered {
		status := "up"
		alive := isServerAlive(s.URL)
		sm.MarkServerStatus(s.UID, alive)
		if !alive {
			status = "down"
		}
		log.WithFields(log.Fields{
			"url":    s.URL,
			"status": status,
		}).Info()
	}
}

// isAlive checks whether a server is Alive by establishing a TCP connection
func isServerAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.WithFields(log.Fields{"error": err.Error()}).Error("Site unreachable")
		return false
	}
	_ = conn.Close()
	return true
}
