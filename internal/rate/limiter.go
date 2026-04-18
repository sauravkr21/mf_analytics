package rate

import "time"

type Limiter struct {
	sec  <-chan time.Time
	min  <-chan time.Time
	hour <-chan time.Time
}

func NewLimiter() *Limiter {
	return &Limiter{
		sec:  time.Tick(time.Second / 2),
		min:  time.Tick(time.Minute / 50),
		hour: time.Tick(time.Hour / 300),
	}
}

func (l *Limiter) Wait() {
	<-l.sec
	<-l.min
	<-l.hour
}