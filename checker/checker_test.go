package checker

import (
	"testing"

	"github.com/k0kubun/pp"
)

func TestCheck(t *testing.T) {
	msgs, err := Check("testdata.go")
	if err != nil {
		t.Error(err)
	}
	pp.Println(msgs)
}
