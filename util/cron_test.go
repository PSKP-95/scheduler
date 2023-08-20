package util

import (
	"testing"
)

func TestIsValidCron(t *testing.T) {
	testCases := []struct {
		expr    string
		isValid bool
	}{
		{
			expr:    "* * * * *",
			isValid: true,
		},
		{
			expr:    "** * * *",
			isValid: false,
		},
		{
			expr:    "*/2 * * * *",
			isValid: true,
		},
		{
			expr:    "* * * * 1-5",
			isValid: true,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.expr, func(t *testing.T) {
			err := IsValidCron(tc.expr)

			if (err != nil && tc.isValid) || (err == nil && !tc.isValid) {
				t.Failed()
			}
		})
	}
}
