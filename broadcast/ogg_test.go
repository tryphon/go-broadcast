package broadcast

import (
	"fmt"
	ogg "github.com/tryphon/go-ogg"
	"os"
	"testing"
)

type MockOggHandler struct {
	SerialNo int32
	Packets  []*ogg.Packet
}

func (mock *MockOggHandler) NewStream(serialNo int32) {
	mock.SerialNo = serialNo
}

func (mock *MockOggHandler) PacketOut(packet *ogg.Packet) bool {
	mock.Packets = append(mock.Packets, packet)
	return true
}

func (mock *MockOggHandler) LastPacket() (packet *ogg.Packet) {
	if len(mock.Packets) > 0 {
		return mock.Packets[len(mock.Packets)-1]
	} else {
		return nil
	}
}

func TestOggDecoder_Reset(t *testing.T) {
	decoder := new(OggDecoder)
	decoder.oss.SerialNo = 123

	decoder.Reset()

	if decoder.oss.SerialNo != 0 {
		t.Errorf("Should reset oss.SerialNo")
	}
}

func TestOggDecoder_Read(t *testing.T) {
	handler := new(MockOggHandler)
	decoder := OggDecoder{handler: handler}

	expectedPacketNumbers := []int64{0, 2, 70, 139, 208, 277, 346, 415, 483, 551, 619, 652}

	for oggPageIndex, expectedPacketNo := range expectedPacketNumbers {
		oggPageFileName := fmt.Sprintf("testdata/ogg_page_%04d", oggPageIndex)

		file, err := os.Open(oggPageFileName)
		if err != nil {
			t.Fatal(err)
		}

		decoder.Read(file)

		if handler.LastPacket().PacketNo != expectedPacketNo {
			t.Errorf("#%d: Wrong PacketNo:\n got: %d\nwant: %d", oggPageIndex, handler.LastPacket().PacketNo, expectedPacketNo)
		}
	}
}
