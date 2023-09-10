//go:build hook_2
// +build hook_2

package hooks

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

type MyHook2 struct {
	Name string
}

func (m *MyHook2) Init() {
	fmt.Println("Initializing Hook 2")
}

func (m *MyHook2) GetName() string {
	return m.Name
}

func (m *MyHook2) Perform(msg Message, statusChan chan<- Message) {
	log.Info().Msg("Performing Hook 2")
	log.Info().Msgf("%v", msg)
	msg.Details = msg.Schedule.Data
	msg.Type = SUCCESS
	statusChan <- msg
}

func (m *MyHook2) Destroy() {
	fmt.Println("Destroying Hook 2")
}

func init() {
	fmt.Println("Registering hook 2")
	register(&MyHook2{Name: "MyHook2"})
}
