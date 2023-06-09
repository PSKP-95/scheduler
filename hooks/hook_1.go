//go:build hook_1
// +build hook_1

package hooks

import (
	"fmt"
)

type MyHook1 struct {
	Name string
}

func (m *MyHook1) Init() {
	fmt.Println("Initializing Hook 1")
}

func (m *MyHook1) GetName() string {
	return m.Name
}

func (m *MyHook1) Perform(msg Message, statusChan chan<- Message) {
	fmt.Println("Performing Hook 1")
	fmt.Println(msg)
	msg.Type = SUCCESS
	statusChan <- msg
}

func (m *MyHook1) Destroy() {
	fmt.Println("Destroying Hook 1")
}

func init() {
	fmt.Println("Registering hook 1")
	register(&MyHook1{Name: "MyHook1"})
}
