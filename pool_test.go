package pool

import (
	"fmt"
	"testing"
	"time"
)

type Logger struct {
}

func (l Logger) Error(i interface{}) {
	fmt.Println("error:", i)
}

func (l Logger) Info(i interface{}) {
	fmt.Println("info", i)
}

func workFunc(in <-chan interface{}, done <-chan struct{}, l LoggerInterface) {
	for {
		select {
		case <-done:
			return
		case msg := <-in:
			l.Info(msg)
		}
	}
}

func TestNew(t *testing.T) {
	p := New(Config{
		Count:      10,
		Logger:     Logger{},
		Duration:   time.Second,
		Func:       workFunc,
		BufferSize: 100,
	})

	p.Run()

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	i := 0
loop:
	for {
		select {
		case <-timer.C:
			break loop
		case <-ticker.C:
			p.Push(i)
			i++
		}
	}
}
