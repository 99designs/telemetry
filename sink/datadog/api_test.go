package datadog

import (
	"encoding/json"
	"testing"
)

func TestPointEncode(t *testing.T) {
	p := Points{
		12345678: 32,
		12345679: 44,
	}

	b, err := json.Marshal(p)
	if err != nil {
		t.Error("unexpected error: " + err.Error())
	}

	if string(b) != "[[12345678,32],[12345679,44]]" {
		t.Error("unexpected encoding: " + string(b))
	}
}
