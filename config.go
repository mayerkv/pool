package pool

import "time"

type Config struct {
	Count      int
	Logger     LoggerInterface
	Duration   time.Duration
	Func       WorkerFunc
	BufferSize int
}

type WorkerFunc func(in <-chan interface{}, done <-chan struct{}, l LoggerInterface)

type LoggerInterface interface {
	Info(interface{})
	Error(interface{})
	CloseWriter()
}
