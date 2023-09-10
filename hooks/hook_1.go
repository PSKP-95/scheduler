//go:build hook_1
// +build hook_1

package hooks

import (
	"fmt"

	"github.com/rs/zerolog/log"
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
	log.Info().Msg("Performing MyHook1")
	log.Info().Msgf("%v", msg)
	msg.Details = "Done working..."
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
