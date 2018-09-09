package checker

import (
	"testing"
)

func TestCheck(t *testing.T) {
	err := Check("testdata.go")
	if err != nil {
		t.Error(err)
	}
}
