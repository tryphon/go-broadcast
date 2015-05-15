package command

import (
	"testing"
)

func TestPlayConfig_IsEmpty(t *testing.T) {
	config := PlayConfig{}

	if !config.IsEmpty() {
		t.Errorf("Fresh PlayConfig should be empty")
	}

	config.Http.Url = "dummy"
	if config.IsEmpty() {
		t.Errorf("PlayConfig with an http url shouldn't be empty")
	}
}
