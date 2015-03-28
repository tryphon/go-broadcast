package broadcast

import (
	"testing"
)

func TestHttpStreamOutputs_Create_Identifier(t *testing.T) {
	streams := HttpStreamOutputs{}
	stream := streams.Create(nil)

	if stream.Identifier == "" {
		t.Errorf("Created stream should have an identifier :\n got: '%s'", stream.Identifier)
	}

	secondStream := streams.Create(nil)
	if secondStream.Identifier == stream.Identifier {
		t.Errorf("Stream identifier should be uniq: got %s and %s", secondStream.Identifier, stream.Identifier)
	}

	thirdStream := streams.Create(&BufferedHttpStreamOutputConfig{Identifier: stream.Identifier})
	if thirdStream.Identifier == stream.Identifier {
		t.Errorf("Stream identifier should be uniq: got %s and %s", thirdStream.Identifier, stream.Identifier)
	}

	streamWithDefaultIdentifier := streams.Create(&BufferedHttpStreamOutputConfig{HttpStreamOutputConfig: HttpStreamOutputConfig{Target: "dummy"}})
	streamWithSameTarget := streams.Create(&BufferedHttpStreamOutputConfig{HttpStreamOutputConfig: HttpStreamOutputConfig{Target: streamWithDefaultIdentifier.output.Target}})

	if streamWithSameTarget.Identifier == streamWithDefaultIdentifier.Identifier {
		t.Errorf("Stream identifier should be uniq: got %s and %s", streamWithSameTarget.Identifier, streamWithDefaultIdentifier.Identifier)
	}
}
