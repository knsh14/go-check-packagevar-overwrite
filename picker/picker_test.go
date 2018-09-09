package picker

import (
	"testing"
)

func TestGetComment(t *testing.T) {
	tests := []struct {
		packageName string
		valName     string
	}{
		{"bytes", "ErrTooLarge"},
	}

	for _, tt := range tests {
		_, err := GetComment(tt.packageName, tt.valName)
		if err != nil {
			t.Fatal(err)
		}
	}
}
