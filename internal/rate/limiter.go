package rate

import (
	"errors"
	"time"
)

type Limiter struct {
	sec  <-chan time.Time
	min  <-chan time.Time
	hour <-chan time.Time

	hourCount int
	minCount  int
	lastMin   time.Time
}

// ONLY FOR TESTING
func (l *Limiter) SetHourCount(n int) {
	l.hourCount = n
}

func (l *Limiter) SetMinCount(n int) {
	l.minCount = n
}

func (l *Limiter) SetLastMin(t time.Time) {
	l.lastMin = t
}

func NewLimiter() *Limiter {
	return &Limiter{
		sec:     time.Tick(time.Second / 2),
		min:     time.Tick(time.Minute / 50),
		hour:    time.Tick(time.Hour / 300),
		lastMin: time.Now(),
	}
}

func (l *Limiter) Wait() error {

	// Hour quota check
	if l.hourCount >= 300 {
		return errors.New("quota exhausted")
	}

	// Minute block logic
	if time.Since(l.lastMin) < time.Minute && l.minCount >= 50 {
		time.Sleep(5 * time.Minute)
		l.minCount = 0
		l.lastMin = time.Now()
	}

	<-l.sec
	<-l.min
	<-l.hour

	l.hourCount++
	l.minCount++

	return nil
}