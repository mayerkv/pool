# pool
Worker pool

#### Example

```
package main

import (
	"fmt"
	"github.com/mayerkv/pool"
	"log"
	"math/rand"
	"os"
	"time"
)

func work(b *pool.Bus) {
	fmt.Println("work!")

	for {
		select {
		case <-b.Done:
			return
		case msg := <-b.In:
			b.Logger.Println("getting message from channel")
			b.Logger.Println(msg)

			i, ok := msg.(int)
			if !ok {
				b.Logger.Println("bad message type")
				continue
			}

			b.Logger.Printf("origin = %v, double = %v\n", i, i*2)
		}
	}
}

func main() {
	c := pool.Config{
		Cnt:        3, // count of workers
		Logger:     log.New(os.Stdout, "[MAIN POOL] ", log.Ldate|log.Ltime|log.Lshortfile), // logger
		Duration:   2 * time.Second, // frequency of adjusting the number of workers
		Func:       work, // worker function
		BufferSize: 100, // size of channel buffer size
	}

	p := pool.New(c) // create new worker pool

	p.Run() // start worker pool

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		p.Bus().In <- rand.Intn(100) // send message to pool.bus.In channel
	}

}

```

#### We get the following:

```bash
$ go run main.go
work!
[MAIN POOL] 2018/10/09 19:17:58 pool.go:48: add worker
[MAIN POOL] 2018/10/09 19:17:58 pool.go:48: add worker
work!
[MAIN POOL] 2018/10/09 19:17:58 pool.go:48: add worker
work!
[MAIN POOL] 2018/10/09 19:17:58 t.go:20: getting message from channel
[MAIN POOL] 2018/10/09 19:17:58 t.go:21: 81
[MAIN POOL] 2018/10/09 19:17:58 t.go:29: origin = 81, double = 162
[MAIN POOL] 2018/10/09 19:17:58 t.go:20: getting message from channel
[MAIN POOL] 2018/10/09 19:17:58 t.go:21: 87
[MAIN POOL] 2018/10/09 19:17:58 t.go:29: origin = 87, double = 174
[MAIN POOL] 2018/10/09 19:17:58 t.go:20: getting message from channel
[MAIN POOL] 2018/10/09 19:17:58 t.go:21: 47
...
```
