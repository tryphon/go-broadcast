package broadcast

import (
	"errors"
	ogg "github.com/tryphon/go-ogg"
	"io"
	"math/rand"
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

	if resettable, ok := decoder.handler.(Resettable); ok {
		resettable.Reset()
	}
}

type OggPacketHandler interface {
	PacketAvailable(packet *ogg.Packet)
}

type OggHandler interface {
	NewStream(serialNo int32)
	PacketOut(packet *ogg.Packet) bool
}

func (decoder *OggDecoder) SetHandler(handler OggHandler) {
	decoder.handler = handler
}

func (decoder *OggDecoder) SetAudioHandler(audioHandler AudioHandler) {
	if audioHandlerSupport, ok := decoder.handler.(AudioHandlerSupport); ok {
		audioHandlerSupport.SetAudioHandler(audioHandler)
	}
}

func (decoder *OggDecoder) Init() error {
	return nil
}

func (decoder *OggDecoder) Read(reader io.Reader) error {
	// defer func() {
	// 	if err := recover(); err != nil {
	// 		Log.Printf("Exception occured in Ogg/Vorbis decoder : %s", err)
	// 		return err
	// 	}
	// }()

	buffer := decoder.oy.Buffer(4096)

	readLength, err := reader.Read(buffer)

	if err != nil {
		return err
	}

	decoder.oy.Wrote(readLength)

	for decoder.oy.PageOut(&decoder.og) == 1 {
		if decoder.oss.SerialNo != decoder.og.SerialNo() {
			// Log.Debugf("Init Ogg Stream State %d", decoder.og.SerialNo())
			decoder.oss.Init(decoder.og.SerialNo())

			decoder.handler.NewStream(decoder.og.SerialNo())
		}

		err = decoder.oss.PageIn(&decoder.og)
		if err != nil {
			return err
		}

		var packetOutResult = 1
		for packetOutResult == 1 {
			packetOutResult = decoder.oss.PacketOut(&decoder.op)
			if packetOutResult == 1 {
				if !decoder.handler.PacketOut(&decoder.op) {
					return errors.New("Can't decode packet")
				}
			}
		}

		if packetOutResult < 0 {
			Log.Debugf("PacketOutResult: %d", packetOutResult)
			// the second page of a Ogg stream seems to return a nice -1 ...
			// 	return false
		}
	}

	return nil
}

type OggEncoder struct {
	Writer  io.Writer
	Encoder *VorbisEncoder

	oss ogg.StreamState // take physical pages, weld into a logical stream of packets
}

func (encoder *OggEncoder) Init() error {
	encoder.Encoder.PacketHandler = encoder

	encoder.oss.Init(rand.Int31())
	err := encoder.Encoder.Init()
	if err != nil {
		return err
	}

	encoder.Flush()
	return nil
}

func (encoder *OggEncoder) PacketAvailable(packet *ogg.Packet) {
	encoder.oss.PacketIn(packet)
}

func (encoder *OggEncoder) Flush() {
	var page ogg.Page
	for encoder.oss.Flush(&page) {
		encoder.write(&page)
	}
}

func (encoder *OggEncoder) writeAvailablePages() {
	var page ogg.Page
	for encoder.oss.PageOut(&page) {
		encoder.write(&page)
	}
}

func (encoder *OggEncoder) write(page *ogg.Page) {
	if encoder.Writer != nil {
		encoder.Writer.Write(page.Header)
		encoder.Writer.Write(page.Body)
	}
}

func (encoder *OggEncoder) AudioOut(audio *Audio) {
	encoder.Encoder.AudioOut(audio)
	encoder.writeAvailablePages()
}

func (encoder *OggEncoder) Close() {
	encoder.Encoder.Close()
	encoder.Encoder.PacketHandler = nil
	encoder.Flush()
}
