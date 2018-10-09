package pool

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
)

func work(b *Bus) {
	for {
		select {
		case <-b.Done:
			return
		case msg := <-b.In:
			fmt.Println(msg)
			b.Logger.Println("log")
			r := rand.Intn(7)
			if r%7 == 1 {
				panic("is dead")
			}
		}
	}
}

func TestNew(t *testing.T) {
	c := Config{
		Cnt:        5,
		Logger:     log.New(os.Stdout, "POOL ", log.Ldate|log.Ltime),
		Duration:   time.Duration(2 * time.Second),
		Func:       work,
		BufferSize: 100,
	}

	b := New(c)
	b.Run()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		b.Bus().In <- "MESSAGE"
	}

}
