package broadcast

import (
	"encoding/binary"
	"io"
	"math"
)

type SampleFormat interface {
	Write(writer io.Writer, sample float32) error
	Read(reader io.Reader) (float32, error)
	Name() string
	SampleSize() int
}

func ParseSampleFormat(name string) SampleFormat {
	switch name {
	case Sample16bLittleEndian.Name():
		return Sample16bLittleEndian
	case Sample32bLittleEndian.Name():
		return Sample32bLittleEndian
	}
	return nil
}

type sample32bLittleEndian struct{}

var Sample32bLittleEndian sample32bLittleEndian

func (format sample32bLittleEndian) Write(writer io.Writer, sample float32) error {
	return binary.Write(writer, binary.LittleEndian, format.ToInt32(sample))
}

func (format sample32bLittleEndian) Read(reader io.Reader) (sample float32, err error) {
	var value int32

	err = binary.Read(reader, binary.LittleEndian, &value)
	if err != nil {
		return 0, err
	}

	return format.ToFloat(value), nil
}

func (format sample32bLittleEndian) Name() string {
	return "s32le"
}

func (format sample32bLittleEndian) SampleSize() int {
	return 4
}

func (sample32bLittleEndian) ToInt32(sample float32) int32 {
	if sample >= 1.0 {
		return math.MaxInt32
	} else if sample <= -1.0 {
		return math.MinInt32
	} else {
		return int32(sample * math.MaxInt32)
	}
}

func (sample32bLittleEndian) ToFloat(value int32) float32 {
	return float32(value) / math.MaxInt32
}

type sample16bLittleEndian struct{}

var Sample16bLittleEndian sample16bLittleEndian

func (format sample16bLittleEndian) Write(writer io.Writer, sample float32) error {
	return binary.Write(writer, binary.LittleEndian, format.ToInt16(sample))
}

func (format sample16bLittleEndian) Read(reader io.Reader) (sample float32, err error) {
	var value int16

	err = binary.Read(reader, binary.LittleEndian, &value)
	if err != nil {
		return 0, err
	}

	return format.ToFloat(value), nil
}

func (format sample16bLittleEndian) Name() string {
	return "s16le"
}

func (format sample16bLittleEndian) SampleSize() int {
	return 2
}

func (sample16bLittleEndian) ToFloat(value int16) float32 {
	return float32(value) / math.MaxInt16
}

func (sample16bLittleEndian) ToInt16(sample float32) int16 {
	if sample >= 1.0 {
		return math.MaxInt16
	} else if sample <= -1.0 {
		return math.MinInt16
	} else {
		return int16(sample * math.MaxInt16)
	}
}
