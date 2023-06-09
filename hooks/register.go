package hooks

import "fmt"

type Hook interface {
	Init()
	Perform(msg Message, status chan<- Message)
	Destroy()
	GetName() string
}

var hooks map[string]Hook = make(map[string]Hook)

func register(h Hook) {
	hook := h.GetName()
	fmt.Println(hook)
	hooks[hook] = h
}

func getHooks() map[string]Hook {
	return hooks
}
