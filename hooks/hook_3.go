//go:build hook_3
// +build hook_3

package hooks

import (
	"fmt"
	"time"
)

type MyHook3 struct {
	Name string
}

func (m *MyHook3) Init() {
	fmt.Println("Initializing Hook 2")
}

func (m *MyHook3) GetName() string {
	return m.Name
}

func (m *MyHook3) Perform(msg Message, statusChan chan<- Message) {
	fmt.Println("Performing Hook 3")
	fmt.Println(msg)
	msg.Details = msg.Schedule.Data
	msg.Type = SUCCESS
	time.Sleep(30 * time.Second)
	statusChan <- msg
}

func (m *MyHook3) Destroy() {
	fmt.Println("Destroying Hook 3")
}

func init() {
	fmt.Println("Registering hook 3")
	register(&MyHook3{Name: "MyHook3"})
}
