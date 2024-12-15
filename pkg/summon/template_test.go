package summon

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

type testPrompter struct {
	b         *bytes.Buffer
	choice    int
	userInput string
}

func (tpr *testPrompter) NewPrompt(userPrompt string) {
	tpr.b = bytes.NewBufferString(userPrompt)
}

func (tpr *testPrompter) Choose(choices []string) (string, error) {
	return choices[tpr.choice], nil
}

func (tpr *testPrompter) Input(defaultVal string) (string, error) {
	if tpr.userInput == "" {
		return defaultVal, nil
	}
	return tpr.userInput, nil
}

func TestPrompt(t *testing.T) {
	testFs := fstest.MapFS{}
	testFs["text.txt"] = &fstest.MapFile{Data: []byte("this is a text")}

	tpr := &testPrompter{
		choice: 1,
	}
	// create a summoner to summon text.txt at
	s, err := New(testFs, WithPrompter(tpr))
	assert.NoError(t, err)

	ret, err := s.renderTemplate(`{{ prompt "slotA" "What is your name" "David" }}`)
	assert.NoError(t, err)

	// check default value
	assert.Equal(t, "What is your name", tpr.b.String())
	assert.Equal(t, "David", ret)

	promptVal, err := s.renderTemplate(`{{ promptValue "slotA" }}`)
	assert.NoError(t, err)
	assert.Equal(t, "David", promptVal)

	// test selection (choice is B by the testPrompter config)
	ret, err = s.renderTemplate(`{{ prompt "continue" "Select One" (list "A" "B" "C") }}`)
	assert.NoError(t, err)

	assert.Equal(t, "B", ret)
	promptVal, err = s.renderTemplate(`{{ promptValue "continue" }}`)
	assert.NoError(t, err)
	assert.Equal(t, "B", promptVal)
}
