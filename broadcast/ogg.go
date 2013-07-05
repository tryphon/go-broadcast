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

	decoder.oss.SerialNo = 0
}

type OggHandler interface {
	NewStream(serialNo int32)
	PacketOut(packet *ogg.Packet) bool
}

func (decoder *OggDecoder) SetHandler(handler OggHandler) {
	decoder.handler = handler
}

func (decoder *OggDecoder) Read(reader io.Reader) (result bool) {
	result = false

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Exception occured: ", err)
			result = false
		}
	}()

	buffer := decoder.oy.Buffer(4096)

	readLength, err := reader.Read(buffer)

	if err != nil {
		return
	}

	decoder.oy.Wrote(readLength)

	for decoder.oy.PageOut(&decoder.og) == 1 {
		if decoder.oss.SerialNo != decoder.og.SerialNo() {
			// fmt.Printf("Init Ogg Stream State %d\n", decoder.og.SerialNo())
			decoder.oss.Init(decoder.og.SerialNo())

			decoder.handler.NewStream(decoder.og.SerialNo())
		}

		err = decoder.oss.PageIn(&decoder.og)
		if err != nil {
			return
		}

		var packetOutResult = 1
		for packetOutResult == 1 {
			packetOutResult = decoder.oss.PacketOut(&decoder.op)
			if packetOutResult == 1 {
				if !decoder.handler.PacketOut(&decoder.op) {
					return
				}
			}
		}

		if packetOutResult < 0 {
			fmt.Printf("PacketOutResult: %d\n", packetOutResult)
			// the second page of a Ogg stream seems to return a nice -1 ...
			// 	return false
		}
	}

	result = true
	return
}
