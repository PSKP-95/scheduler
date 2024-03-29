package hooks

import db "github.com/PSKP-95/scheduler/db/sqlc"

type MsgType int

const (
	TRIGGER MsgType = iota
	SCHEDULED
	FAILED
	SUCCESS
)

type Message struct {
	Type      MsgType
	Occurence db.NextOccurence
	Schedule  db.Schedule
	Details   string
}
