package loadbalancer

import (
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	log "github.com/sirupsen/logrus"
)

const (
	//Attempts is self explanatory
	Attempts int = iota
	//Retry is number of retries.
	Retry
)

// Server holds the data about a server
type Server struct {
	UID          uint32
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

// SetAlive for this server
func (s *Server) SetAlive(alive bool) {
	s.mux.Lock()
	s.Alive = alive
	s.mux.Unlock()
}

// IsAlive returns true when server is alive
func (s *Server) IsAlive() (alive bool) {
	s.mux.RLock()
	alive = s.Alive
	s.mux.RUnlock()
	return
}

//ServerPool isolates server management.
type ServerPool struct {
	Registered []*Server
	current    uint64
	mux        sync.RWMutex
	muxID      sync.RWMutex
	//Set of active server IDs.
	active map[uint32]struct{}
	//Maps server ID to Server Array.
	idMap map[uint32]int
}

func (sp *ServerPool) register(s *Server) {
	sp.Registered = append(sp.Registered, s)
	sp.idMap[s.UID] = len(sp.Registered) - 1
}
func (sp *ServerPool) kill(serverID uint32) {
	sp.muxID.RLock()
	index, _ := sp.idMap[serverID]
	sp.muxID.RUnlock()
	sp.Registered[index].SetAlive(false)
	sp.mux.Lock()
	delete(sp.active, serverID)
	sp.mux.Unlock()

}
func (sp *ServerPool) start(serverID uint32) {
	sp.muxID.RLock()
	index, _ := sp.idMap[serverID]
	sp.muxID.RUnlock()
	sp.Registered[index].SetAlive(true)
	sp.mux.Lock()
	sp.active[serverID] = struct{}{}
	sp.mux.Unlock()
}

//ServerManager holds information about reachable servers
type ServerManager struct {
	pool      *ServerPool
	scheduler AlgoScheduler
}

//NewServerManager is a factory for ServerManager structs
func NewServerManager(algoType string) (*ServerManager, error) {
	switch algoType {
	case "cycle":
		return &ServerManager{
			scheduler: &Cycle{},
			pool: &ServerPool{idMap: make(map[uint32]int),
				active: make(map[uint32]struct{})},
		}, nil
	default:
		return nil, errors.New("Invalid algorithm type")
	}
}

// AddServer to the server mananager
func (sm *ServerManager) AddServer(s *Server) {
	sm.pool.register(s)
}

//Schedule self-explanatory
func (sm *ServerManager) Schedule() *Server {
	return sm.scheduler.Schedule(sm)
}

// MarkServerStatus changes a status of a server
func (sm *ServerManager) MarkServerStatus(serverUID uint32, alive bool) {
	_, ok := sm.pool.active[serverUID]
	if !ok {
		return
	}
	if alive {
		sm.pool.kill(serverUID)
	} else {
		sm.pool.start(serverUID)
	}
}

// GetAttemptsFromContext returns the attempts for request
func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(Attempts).(int); ok {
		return attempts
	}
	return 1
}

// GetRetryFromContext returns the retries for request
func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(Retry).(int); ok {
		return retry
	}
	return 0
}

// Handle load balances the incoming request
func Handle(w http.ResponseWriter, r *http.Request, sm *ServerManager) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.WithFields(log.Fields{"address": r.RemoteAddr,
			"path": r.URL.Path,
		}).Error("Max attempts reached, terminating")
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	peer := sm.Schedule() //needs to be replaced
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}
