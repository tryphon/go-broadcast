package broadcast

import (
	"bytes"
	"math"
	"testing"
)

func TestSample32bLittleEndian_Write(t *testing.T) {
	var conditions = []struct {
		sample float32
		output []byte
	}{
		{1, []byte{255, 255, 255, 127}},
		{0, []byte{0, 0, 0, 0}},
		{-1, []byte{0, 0, 0, 128}},
	}

	buffer := &bytes.Buffer{}
	for _, condition := range conditions {
		Sample32bLittleEndian.Write(buffer, condition.sample)
		output := buffer.Bytes()

		if !bytes.Equal(output, condition.output) {
			t.Errorf("Wrong Write output for sample %v:\n got: %v\nwant: %v", condition.sample, output, condition.output)
		}

		buffer.Reset()
	}
}

func TestSample32bLittleEndian_Read(t *testing.T) {
	var conditions = []struct {
		sample float32
		input  []byte
	}{
		{1, []byte{255, 255, 255, 127}},
		{0, []byte{0, 0, 0, 0}},
		{-1, []byte{0, 0, 0, 128}},
	}

	buffer := &bytes.Buffer{}
	for _, condition := range conditions {
		buffer.Write(condition.input)
		sample, _ := Sample32bLittleEndian.Read(buffer)

		if sample != condition.sample {
			t.Errorf("Wrong Read value for input %v:\n got: %v\nwant: %v", condition.input, sample, condition.sample)
		}

		buffer.Reset()
	}
}

func TestSample32bLittleEndian_ToInt32(t *testing.T) {
	var conditions = []struct {
		sample float32
		value  int32
	}{
		{1.0, math.MaxInt32},
		{0.5, math.MaxInt32 / 2},
		{0.1, math.MaxInt32 / 10},
		{0.0, 0},
		{-0.5, math.MinInt32 / 2},
		{-1.0, math.MinInt32},
	}

	for _, condition := range conditions {
		value := Sample32bLittleEndian.ToInt32(condition.sample)
		// float64 * int32 doesn't give exact result ...
		if !sameInt32(value, condition.value, 4) {
			t.Errorf("Wrong value for sample %v:\n got: %d\nwant: %d", condition.sample, value, condition.value)
		}
	}
}

func TestSample32bLittleEndian_ToFloat(t *testing.T) {
	var conditions = []struct {
		sample float32
		value  int32
	}{
		{1.0, math.MaxInt32},
		{0.5, math.MaxInt32 / 2},
		{0.1, math.MaxInt32 / 10},
		{0.0, 0},
		{-0.5, math.MinInt32 / 2},
		{-1.0, math.MinInt32},
	}

	for _, condition := range conditions {
		sample := Sample32bLittleEndian.ToFloat(condition.value)
		if !sameFloat32(sample, condition.sample, 0.0000001) {
			t.Errorf("Wrong sample for value %v:\n got: %v\nwant: %v", condition.value, sample, condition.sample)
		}
	}
}

func sameFloat32(value1 float32, value2 float32, tolerance float32) bool {
	return math.Abs(float64(value1-value2)) <= float64(tolerance)
}

func sameInt32(value1 int32, value2 int32, tolerance int32) bool {
	return math.Abs(float64(value1-value2)) <= float64(tolerance)
}

func TestSample16bLittleEndian_Write(t *testing.T) {
	var conditions = []struct {
		sample float32
		output []byte
	}{
		{1, []byte{255, 127}},
		{0, []byte{0, 0}},
		{-1, []byte{0, 128}},
	}

	buffer := &bytes.Buffer{}
	for _, condition := range conditions {
		Sample16bLittleEndian.Write(buffer, condition.sample)
		output := buffer.Bytes()

		if !bytes.Equal(output, condition.output) {
			t.Errorf("Wrong Write output for sample %v:\n got: %v\nwant: %v", condition.sample, output, condition.output)
		}

		buffer.Reset()
	}
}

func TestSample16bLittleEndian_Read(t *testing.T) {
	var conditions = []struct {
		sample float32
		input  []byte
	}{
		{1, []byte{255, 127}},
		{0, []byte{0, 0}},
		{-1, []byte{0, 128}},
	}

	buffer := &bytes.Buffer{}
	for _, condition := range conditions {
		buffer.Write(condition.input)
		sample, _ := Sample16bLittleEndian.Read(buffer)

		if !sameFloat32(sample, condition.sample, 0.0001) {
			t.Errorf("Wrong Read value for input %v:\n got: %v\nwant: %v", condition.input, sample, condition.sample)
		}

		buffer.Reset()
	}
}

func TestSample16bLittleEndian_ToInt16(t *testing.T) {
	var conditions = []struct {
		sample float32
		value  int16
	}{
		{1.0, math.MaxInt16},
		{0.5, math.MaxInt16 / 2},
		{0.1, math.MaxInt16 / 10},
		{0.0, 0},
		{-0.5, math.MinInt16 / 2},
		{-1.0, math.MinInt16},
	}

	for _, condition := range conditions {
		value := Sample16bLittleEndian.ToInt16(condition.sample)
		// float64 * int16 doesn't give exact result ...
		if !sameInt16(value, condition.value, 1) {
			t.Errorf("Wrong value for sample %v:\n got: %d\nwant: %d", condition.sample, value, condition.value)
		}
	}
}

func TestSample16bLittleEndian_ToFloat(t *testing.T) {
	var conditions = []struct {
		sample float32
		value  int16
	}{
		{1.0, math.MaxInt16},
		{0.5, math.MaxInt16 / 2},
		{0.1, math.MaxInt16 / 10},
		{0.0, 0},
		{-0.5, math.MinInt16 / 2},
		{-1.0, math.MinInt16},
	}

	for _, condition := range conditions {
		sample := Sample16bLittleEndian.ToFloat(condition.value)
		if !sameFloat32(sample, condition.sample, 0.0001) {
			t.Errorf("Wrong sample for value %v:\n got: %v\nwant: %v", condition.value, sample, condition.sample)
		}
	}
}

func sameInt16(value1 int16, value2 int16, tolerance int16) bool {
	return math.Abs(float64(value1-value2)) <= float64(tolerance)
}

func TestParseSampleFormat(t *testing.T) {
	var conditions = []struct {
		name         string
		sampleFormat SampleFormat
	}{
		{"s32le", Sample32bLittleEndian},
		{"s16le", Sample16bLittleEndian},
	}

	for _, condition := range conditions {
		sampleFormat := ParseSampleFormat(condition.name)

		if sampleFormat != condition.sampleFormat {
			t.Errorf("Wrong sample for name %v:\n got: %v\nwant: %v", condition.name, sampleFormat, condition.sampleFormat)
		}
	}
}
