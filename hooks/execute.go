package hooks

import (
	"context"
	"fmt"

	db "github.com/PSKP-95/schedular/db/sqlc"
	"github.com/PSKP-95/schedular/util"
)

type Executor struct {
	config util.Config
	store  db.Store
	hooks  map[string]Hook
	exChan chan Message
}

func NewExecutor(config util.Config, store db.Store, exChan chan Message) (*Executor, error) {
	ex := &Executor{
		config: config,
		store:  store,
		hooks:  getHooks(),
		exChan: exChan,
	}

	return ex, nil
}

func (ex *Executor) GetHooks() map[string]Hook {
	return ex.hooks
}

func (ex *Executor) Submit(msg Message) {
	ex.exChan <- msg
}

func (ex *Executor) Execute() {
	for {
		msg := <-ex.exChan
		switch msg.Type {
		case TRIGGER:
			schedule, err := ex.store.GetSchedule(context.Background(), msg.Occurence.Schedule)
			if err != nil {
				fmt.Println("Processing failed")
			}
			go ex.hooks[schedule.Hook].Perform(msg, ex.exChan)
		case SUCCESS:
			fmt.Println("Processing Success")
		}
	}
}
