package worker

import (
	"fmt"
	"time"
)

func (worker *Worker) Work() {
	for range time.Tick(10 * time.Second) {
		fmt.Println("Hi from worker")
	}
}
