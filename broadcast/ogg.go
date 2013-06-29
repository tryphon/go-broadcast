package broadcast

import (
	"fmt"
	"github.com/grd/ogg"
	"io"
)

type OggDecoder struct {
	handler OggHandler

	oy  ogg.SyncState
	oss ogg.StreamState // take physical pages, weld into a logical stream of packets
	og  ogg.Page        // one Ogg bitstream page. Vorbis packets are inside
	op  ogg.Packet      // one raw packet of data for decode
}

func (decoder *OggDecoder) Reset() {
	decoder.oy.Reset()
	decoder.oss.Reset()
	decoder.og.Reset()
}

type OggHandler interface {
	NewStream(serialNo int32)
	PacketOut(packet *ogg.Packet)
}

func (decoder *OggDecoder) SetHandler(handler OggHandler) {
	decoder.handler = handler
}

func (decoder *OggDecoder) Read(reader io.Reader) bool {
	buffer := decoder.oy.Buffer(4096)

	readLength, err := reader.Read(buffer)
	if err != nil {
		return false
	}

	decoder.oy.Wrote(readLength)

	for decoder.oy.PageOut(&decoder.og) == 1 {
		if decoder.oss.SerialNo != decoder.og.SerialNo() {
			fmt.Printf("Init Ogg Stream State %d\n", decoder.og.SerialNo())
			decoder.oss.Init(decoder.og.SerialNo())

			decoder.handler.NewStream(decoder.og.SerialNo())
		}

		err = decoder.oss.PageIn(&decoder.og)
		if err != nil {
			return false
		}

		for decoder.oss.PacketOut(&decoder.op) == 1 {
			// fmt.Printf("PacketOut\n");

			// if result < 1 {
			// 	fmt.Printf("Error reading next packet.\n");
			// 	os.Exit(1)
			// }

			decoder.handler.PacketOut(&decoder.op)
		}
	}

	return true
}
