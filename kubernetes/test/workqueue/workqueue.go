package main

import (
	"fmt"
	"sync"
	"time"

	"k8s.io/client-go/util/workqueue"
)

func main() {
	// wq()
	// dwq()
	rdwq()
}
func rdwq() {
	queue := workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(500*time.Millisecond, time.Second))

	for i := 0; i < 10; i++ {
		queue.AddRateLimited(i)
	}
	for {
		item, _ := queue.Get()
		fmt.Printf("%s\n", item)
	}
}
func dwq() {
	queue := workqueue.NewDelayingQueue()
	queue.AddAfter("hello", time.Second)
	item, quit := queue.Get()
	if quit {
		fmt.Printf("quit\n")
	}
	fmt.Printf("item handle begin: %s\n", item)
	queue.Done(item)
	fmt.Printf("item handle done: %s\n", item)
}
func wq() {
	tests := []struct {
		queue         *workqueue.Type
		queueShutDown func(workqueue.Interface)
	}{
		{
			queue:         workqueue.New(),
			queueShutDown: workqueue.Interface.ShutDown,
		},
		{
			queue:         workqueue.New(),
			queueShutDown: workqueue.Interface.ShutDownWithDrain,
		},
	}

	for _, test := range tests {
		const producers = 50
		producerWG := sync.WaitGroup{}
		producerWG.Add(producers)
		for i := 0; i < producers; i++ {
			go func(i int) {
				defer producerWG.Done()
				for j := 0; j < 50; j++ {
					test.queue.Add(i)
					time.Sleep(time.Millisecond)
				}
			}(i)
		}

		const consumers = 10
		consumerWG := sync.WaitGroup{}
		consumerWG.Add(consumers)
		for i := 0; i < consumers; i++ {
			go func(i int) {
				defer consumerWG.Done()
				for {
					item, quit := test.queue.Get()
					if item == "added after shutdown!" {
						fmt.Printf("Got an item added after shutdown.\n")
					}
					if quit {
						return
					}
					// fmt.Printf("Worker %v: begin processing %v\n", i, item)
					time.Sleep(3 * time.Millisecond)
					// fmt.Printf("Worker %v: done processing %v\n", i, item)
					test.queue.Done(item)
				}
			}(i)
		}

		producerWG.Wait()
		test.queueShutDown(test.queue)
		test.queue.Add("added after shutdown!")
		consumerWG.Wait()
		if test.queue.Len() != 0 {
			fmt.Printf("Expected the queue to be empty, had: %v items", test.queue.Len())
		}
	}
}
