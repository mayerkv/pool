package pool

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type Pool struct {
	in       chan interface{}
	done     chan struct{}
	dead     chan error
	needCnt  int
	cnt      int32
	f        WorkerFunc
	logger   *log.Logger
	duration time.Duration
	once     sync.Once
}

type WorkerFunc func(b *Bus)

type Bus struct {
	In     chan interface{}
	Done   chan struct{}
	Logger *log.Logger
	Dead   chan error
}

type Config struct {
	Cnt        int
	Logger     *log.Logger
	Duration   time.Duration
	Func       WorkerFunc
	BufferSize int
}

func (p *Pool) panicRecover() {
	if r := recover(); r != nil {
		p.dead <- fmt.Errorf("worker recover with - %v", r)
	}
}

func (p *Pool) add() {
	go p.work()
	p.logger.Println("add worker")
	atomic.AddInt32(&p.cnt, 1)
}

func (p *Pool) kill() {
	p.done <- struct{}{}
	p.logger.Println("kill worker")
	atomic.AddInt32(&p.cnt, -1)
}

func (p *Pool) work() {
	defer p.panicRecover()
	p.f(p.Bus())
}

func New(c Config) (b *Pool) {
	return &Pool{
		in:       make(chan interface{}, c.BufferSize),
		done:     make(chan struct{}, c.Cnt),
		dead:     make(chan error, c.Cnt),
		needCnt:  c.Cnt,
		cnt:      0,
		logger:   c.Logger,
		duration: c.Duration,
		f:        c.Func,
	}
}

func (p *Pool) Bus() *Bus {
	return &Bus{
		In:     p.in,
		Done:   p.done,
		Logger: p.logger,
		Dead:   p.dead,
	}
}

func (p *Pool) Run() {
	p.once.Do(func() {
		go p.run()
	})
}

func (p *Pool) run() {
	for i := 0; i < p.needCnt; i++ {
		p.add()
	}

	go p.watchDead()

	ticker := time.NewTicker(p.duration)
	defer ticker.Stop()

	for range ticker.C {
		cnt := atomic.LoadInt32(&p.cnt)
		abs := p.needCnt - int(cnt)

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

func (p *Pool) watchDead() {
	for err := range p.dead {
		p.logger.Println("worker is dead:", err.Error())
		atomic.AddInt32(&p.cnt, -1)
	}
}
