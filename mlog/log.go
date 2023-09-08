package mlog

import "log"

type Log struct {
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}
