package testutil

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnMarshallCall(t *testing.T) {
	out := &bytes.Buffer{}

	call1 := Call{
		Args: "a b c",
		Env:  []string{"a=b"},
	}
	call2 := Call{
		Args: "a b c",
		Env:  []string{"a=b"},
	}

	startCall(out)

	writeCall(call1, out)
	willAppendCall(out)
	writeCall(call2, out)

	c, err := GetCalls(out)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(c.Calls))
	assert.Contains(t, c.Calls, call1)
	assert.Contains(t, c.Calls, call2)
}

func TestMarshallCalls(t *testing.T) {
	c := Calls{Calls: []Call{
		Call{
			Args: "a b c",
			Env:  []string{"a=b"},
			Out:  "output",
		},
	}}

	b, err := json.Marshal(c)

	assert.Nil(t, err)
	assert.Equal(t, "{\"Calls\":[{\"Args\":\"a b c\",\"Env\":[\"a=b\"],\"Out\":\"output\"}]}", string(b))
}
