package loadbalancer

import "sync/atomic"

//AlgoScheduler interface helps implement strategy pattern for algorithms.
type AlgoScheduler interface {
	Schedule(sm *ServerManager) *Server
}

// Cycle implements a scheduling algorithm that iterates cyclically through alive servers.
type Cycle struct {
}

// WRR implements Weighted Round-Robin scheduling through alive servers.
type WRR struct {
}

// IWRR implements interleaved Weighted Round-Robin scheduling through alive servers.
type IWRR struct {
}

//Schedule method
func (algo *Cycle) Schedule(sm *ServerManager) *Server {
	// loop entire servers to find out an Alive server
	next := int(atomic.AddUint64(&sm.pool.current, uint64(1)) % uint64(len(sm.pool.Registered)))
	l := len(sm.pool.Registered) + next // start from next and move a full cycle
	for i := next; i < l; i++ {
		idx := i % len(sm.pool.Registered)     // take an index by modding
		if sm.pool.Registered[idx].IsAlive() { // if we have an alive server, use it and store if its not the original one
			if i != next {
				atomic.StoreUint64(&sm.pool.current, uint64(idx))
			}
			return sm.pool.Registered[idx]
		}
	}
	return nil
}
