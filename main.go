package main

import (
	"fmt"
	"time"
)

func main() {
	q := NewQueue(3)
	fmt.Println(q)
	for _, worker := range q.Workers {
		go worker.Start()
	}
	q.Enqueue("echo job one")
	q.Enqueue("echo job two")
	q.Enqueue("echo job three")
	time.Sleep(time.Second * 2)
}
