package main

import (
	"context"
	"fmt"
	"os/signal"
	"sync"
	"syscall"
)

func main() {

	var wg sync.WaitGroup
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	q := NewQueue(3)
	var count int = 0
	fmt.Println(q)
	for _, worker := range q.Workers {
		wg.Add(1)
		count++
		go worker.Start(ctx, &wg, count)
	}
	q.Enqueue("echo job one")
	q.Enqueue("echo job two")
	q.Enqueue("echo job three")

	<-ctx.Done()
	wg.Wait()
	fmt.Println("\nProcesses stopped!")
}
