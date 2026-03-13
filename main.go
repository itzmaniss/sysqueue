package main

import (
	"context"
	"fmt"
	"os/signal"
	"sync"
	"syscall"

	"github.com/itzmaniss/sysqueue/queue"
	"github.com/itzmaniss/sysqueue/server"
	"github.com/itzmaniss/sysqueue/worker"
)

func main() {
	var wg sync.WaitGroup
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	q := queue.NewQueue()

	for i := range 3 {
		w := worker.NewWorker(q.JobQueue, q.Jobs, &q.Lock)
		wg.Add(1)
		go w.Start(ctx, &wg, i+1)
	}

	server := server.NewServer(q)
	wg.Add(1)
	go server.Start(ctx, &wg)

	q.Enqueue("echo job one")
	q.Enqueue("echo job two")
	q.Enqueue("echo job three")

	<-ctx.Done()
	wg.Wait()
	fmt.Println("\nProcesses stopped!")
}
