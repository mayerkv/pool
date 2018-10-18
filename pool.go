package pool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Pool struct {
	in    chan interface{}
	done  chan struct{}
	dead  chan error
	count int32

	once   sync.Once
	config Config
}

//создает экземпляр пула воркеров
func New(c Config) *Pool {
	return &Pool{
		config: c,
		in:     make(chan interface{}, c.BufferSize),
		done:   make(chan struct{}, c.Count),
		dead:   make(chan error, c.Count),
		count:  0,
	}
}

//запуск пула
func (p *Pool) Run() {
	p.once.Do(func() {
		go p.run()
	})
}

func (p *Pool) Push(msg interface{}) {
	p.in <- msg
}

//при перехвате паники говорит что воркер сдох
func (p *Pool) panicRecover() {
	if r := recover(); r != nil {
		p.dead <- fmt.Errorf("worker recover with - %v", r)
	}
}

//добавляет воркера
func (p *Pool) add() {
	go p.work()
	p.config.Logger.Info("add worker")
	atomic.AddInt32(&p.count, 1)
}

//говорит воркеру что ему нужно сдохнуть
func (p *Pool) kill() {
	p.done <- struct{}{}
	p.config.Logger.Info("kill worker")
	atomic.AddInt32(&p.count, -1)
}

//обертка над функцией воркера для перехвата паники
func (p *Pool) work() {
	defer p.panicRecover()
	p.config.Func(p.in, p.done, p.config.Logger)
}

//слушает канал для понимания что воркер сдох
func (p *Pool) watchDead() {
	for err := range p.dead {
		p.config.Logger.Error(fmt.Sprintf("worker is dead: %v", err.Error()))
		atomic.AddInt32(&p.count, -1)
	}
}

//регулирует количество воркеров
func (p *Pool) run() {
	for i := 0; i < p.config.Count; i++ {
		p.add()
	}

	go p.watchDead()

	ticker := time.NewTicker(p.config.Duration)
	defer ticker.Stop()

	for range ticker.C {
		cnt := atomic.LoadInt32(&p.count)
		abs := p.config.Count - int(cnt)

		switch {
		case abs < 0:
			for i := 0; i < -abs; i++ {
				p.kill()
			}
		case abs > 0:
			for i := 0; i < abs; i++ {
				p.add()
			}
		}
	}
}
