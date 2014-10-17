package broadcast

import (
	"testing"
)

func TestBufferedHttpStreamOutput_defaultIdentifier(t *testing.T) {
	output := NewBufferedHttpStreamOutput()
	output.output.Target = "http://source:secret@stream-in.tryphon.eu:8000/test.ogg"
	if output.defaultIdentifier() != "42336522" {
		t.Errorf("Wrong default identifier :\n got: %v\nwant: %v", output.defaultIdentifier(), "42336522")
	}
}
