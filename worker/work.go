package worker

import (
	"context"
	"fmt"
	"time"
)

func (worker *Worker) Work() {
	for range time.Tick(10 * time.Second) {
		fmt.Println("Hi from worker")
		worker.removeDeadBodies()
		worker.punchCard()
	}
}

func (worker *Worker) removeDeadBodies() {
	err := worker.store.RemoveDeadWorkers(context.Background())
	fmt.Println(err)
}

func (worker *Worker) punchCard() {
	err := worker.store.ProveLiveliness(context.Background(), worker.id)
	fmt.Println(err)
}
