package hooks

import (
	"fmt"

	"github.com/PSKP-95/scheduler/mlog"
)

type Hook interface {
	Init()
	Perform(msg Message, status chan<- Message, log *mlog.Log)
	Destroy()
	GetName() string
}

var hooks map[string]Hook = make(map[string]Hook)

func register(h Hook) {
	hook := h.GetName()
	fmt.Println(hook)
	h.Init()
	hooks[hook] = h
}

func getHooks() map[string]Hook {
	return hooks
}
