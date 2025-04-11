package clock

import "time"

var GlobalClock Clock = NewClock()

type Clock interface {
	CurrentTime() time.Time
}

type ClockImpl struct{}

func NewClock() Clock {
	return &ClockImpl{}
}

func (c *ClockImpl) CurrentTime() time.Time {
	return time.Now().UTC()
}

func Now() time.Time {
	return GlobalClock.CurrentTime()
}
