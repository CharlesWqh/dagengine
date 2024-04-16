package graph

import "time"

var defaultEventChan = make(chan *Event, 10000)

// GetDefaultEventChan GetDefaultEventChan
func GetDefaultEventChan() chan *Event {
	return defaultEventChan
}

// Event stat processor execute status
type Event struct {
	Processor string
	Duration  time.Duration
}

// AddEvent add event no block
func AddEvent(e *Event) {
	select {
	case defaultEventChan <- e:
	default:
	}
}
