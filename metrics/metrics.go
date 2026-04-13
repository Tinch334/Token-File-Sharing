package metrics


import (
	"time"
)


type Metrics struct {
	// Stores when the server started
	startTime time.Time
	// How many connections have been handled in total.
	handledConnections uint64
}


// InitMetrics initializes all the metrics to their starting values.
func InitMetrics() *Metrics {
	return &Metrics{
		startTime:          time.Now(),
		handledConnections: 0,
	}
}


// GetStartTime returns when the server started running.
func (m *Metrics) GetStartTime() time.Time {
	return m.startTime
}


// AddConn adds one to the connection counter.
func (m *Metrics) AddConn() {
	m.handledConnections++
}


// GetConn returns the amount of handled connections.
func (m *Metrics) GetConn() uint64 {
	return m.handledConnections
}