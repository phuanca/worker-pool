# Worker Pool

This is a package that provides an worker pool executor to run any task that implement the interface.

> This project is for **testing** purpose.

```go
    type Task interface{
        Execute()
    }    
```

## Installation

To install worker-pool, use `go get`:

```go
    go get github.com/phuanca/worker-pool
```

## Example

```go
package your_package

import (
 "log"

 workerpool "github.com/phuanca/worker-pool"
)

// Define struct that implement the executor.Task interface
type sendMessage struct {
 id  int
 msg string
}

func (s sendMessage) Execute() {
 start := time.Now()
 log.Printf("msg_id: %v, doing some work", s.id)
 r := rand.Intn(5000)
 time.Sleep(time.Duration(r) * time.Millisecond)
 log.Printf("msg_id: %v, %v in: %v", s.id, s.msg, time.Since(start))
}

func main() {
 workers := 2 // workers
 maxQueueDepth := 2 // queue size
 
 // create an executor
 wp := workerpool.NewService(workers, maxQueueDepth)
 defer wp.Exit() // you can stop all worker when needed.

 // make executor ready to receive messages (work).
 wp.Run()

 for i := 0; i < 10; i++ {
  number := i
  task := sendMessage{id: number, msg: "snapshot saved"}

  // send task to executor.
  wp.Send(task)
 }
 log.Printf("====== all task sended! ====== ")
 time.Sleep(10 * time.Second)
}

```
