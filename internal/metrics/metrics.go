package metrics


import (
	"time"
	"sync"
)


type Metrics struct {
	// Stores when the server started
	startTime          time.Time
	// How many connections have been handled in total.
	handledConnections uint64
	// How many tokens have been generated in total.
	generatedTokens    uint64

	handConnMutex      sync.RWMutex
	genTokMutex        sync.RWMutex
}


// InitMetrics initializes all the metrics to their starting values.
func InitMetrics() *Metrics {
	return &Metrics{
		startTime:          time.Now(),
		handledConnections: 0,
		generatedTokens:    0,
	}
}


// GetStartTime returns when the server started running.
func (m *Metrics) GetStartTime() time.Time {
	return m.startTime
}


// AddConn adds one to the connection counter.
func (m *Metrics) AddConn() {
	m.handConnMutex.Lock()
	defer m.handConnMutex.Unlock()

	m.handledConnections++
}


// GetConn returns the amount of handled connections.
func (m *Metrics) GetConn() uint64 {
	m.handConnMutex.RLock()
	defer m.handConnMutex.RUnlock()

	return m.handledConnections
}


// AddToken adds one to the token counter.
func (m *Metrics) AddToken() {
	m.genTokMutex.Lock()
	defer m.handConnMutex.Unlock()

	m.generatedTokens++
}


// GetToken returns the amount of generated tokens.
func (m *Metrics) GetTokens() uint64 {
	m.genTokMutex.RLock()
	defer m.genTokMutex.RUnlock()

	return m.generatedTokens
}