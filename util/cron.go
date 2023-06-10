package util

import (
	"time"

	"github.com/gorhill/cronexpr"
)

func CalculateNextOccurence(cron string) (time.Time, error) {
	expression, err := cronexpr.Parse(cron)

	if err != nil {
		return time.Now(), err
	}

	return expression.Next(time.Now()), nil
}

func IsValidCron(cron string) error {
	_, err := cronexpr.Parse(cron)

	if err != nil {
		return err
	}
	return nil
}
