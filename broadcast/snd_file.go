package broadcast

/*
#cgo LDFLAGS: -lsndfile
#include <stdio.h>
#include <stdlib.h>
#include <sndfile.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

// sf_count_t	frames ;
// int			samplerate ;
// int			channels ;
// int			format ;
// int			sections ;
// int			seekable ;
type Info C.SF_INFO

func (info *Info) Frames() int64 {
	return int64(info.frames)
}

func (info *Info) SampleRate() int {
	return int(info.samplerate)
}

func (info *Info) Channels() int {
	return int(info.channels)
}

func (info *Info) SetSampleRate(sampleRate int) {
	info.samplerate = C.int(sampleRate)
}

func (info *Info) SetChannels(channels int) {
	info.channels = C.int(channels)
}

func (info *Info) SetFormat(format int) {
	info.format = C.int(format)
}

type SndFile struct {
	path   string
	info   Info
	handle *C.SNDFILE
}

func (file *SndFile) Info() *Info {
	return &file.info
}

func (file *SndFile) Path() string {
	return file.path
}

func (file *SndFile) Close() error {
	file.WriteSync()

	errorCode := C.sf_close(file.handle)
	if errorCode != 0 {
		return errors.New("Can't close file")
	} else {
		file.handle = nil
		return nil
	}
}

func (file *SndFile) IsClosed() bool {
	return file.handle == nil
}

func (file *SndFile) WriteFloat(samples []float32) int64 {
	return int64(C.sf_write_float(file.handle, (*C.float)(unsafe.Pointer(&samples[0])), C.sf_count_t(len(samples))))
}

func (file *SndFile) ReadFloat(samples []float32) int64 {
	return int64(C.sf_read_float(file.handle, (*C.float)(unsafe.Pointer(&samples[0])), C.sf_count_t(len(samples))))
}

func (file *SndFile) WriteSync() {
	C.sf_write_sync(file.handle)
}

const (
	O_RDONLY int = C.SFM_READ
	O_WRONLY int = C.SFM_WRITE
	O_RDWR   int = C.SFM_RDWR

	/* Major formats. */
	FORMAT_WAV   int = C.SF_FORMAT_WAV   /* Microsoft WAV format (little endian). */
	FORMAT_AIFF  int = C.SF_FORMAT_AIFF  /* Apple/SGI AIFF format (big endian). */
	FORMAT_AU    int = C.SF_FORMAT_AU    /* Sun/NeXT AU format (big endian). */
	FORMAT_RAW   int = C.SF_FORMAT_RAW   /* RAW PCM data. */
	FORMAT_PAF   int = C.SF_FORMAT_PAF   /* Ensoniq PARIS file format. */
	FORMAT_SVX   int = C.SF_FORMAT_SVX   /* Amiga IFF / SVX8 / SV16 format. */
	FORMAT_NIST  int = C.SF_FORMAT_NIST  /* Sphere NIST format. */
	FORMAT_VOC   int = C.SF_FORMAT_VOC   /* VOC files. */
	FORMAT_IRCAM int = C.SF_FORMAT_IRCAM /* Berkeley/IRCAM/CARL */
	FORMAT_W64   int = C.SF_FORMAT_W64   /* Sonic Foundry's 64 bit RIFF/WAV */
	FORMAT_MAT4  int = C.SF_FORMAT_MAT4  /* Matlab (tm) V4.2 / GNU Octave 2.0 */
	FORMAT_MAT5  int = C.SF_FORMAT_MAT5  /* Matlab (tm) V5.0 / GNU Octave 2.1 */
	FORMAT_PVF   int = C.SF_FORMAT_PVF   /* Portable Voice Format */
	FORMAT_XI    int = C.SF_FORMAT_XI    /* Fasttracker 2 Extended Instrument */
	FORMAT_HTK   int = C.SF_FORMAT_HTK   /* HMM Tool Kit format */
	FORMAT_SDS   int = C.SF_FORMAT_SDS   /* Midi Sample Dump Standard */
	FORMAT_AVR   int = C.SF_FORMAT_AVR   /* Audio Visual Research */
	FORMAT_WAVEX int = C.SF_FORMAT_WAVEX /* MS WAVE with WAVEFORMATEX */
	FORMAT_SD2   int = C.SF_FORMAT_SD2   /* Sound Designer 2 */
	FORMAT_FLAC  int = C.SF_FORMAT_FLAC  /* FLAC lossless file format */
	FORMAT_CAF   int = C.SF_FORMAT_CAF   /* Core Audio File format */
	FORMAT_WVE   int = C.SF_FORMAT_WVE   /* Psion WVE format */
	FORMAT_OGG   int = C.SF_FORMAT_OGG   /* Xiph OGG container */
	FORMAT_MPC2K int = C.SF_FORMAT_MPC2K /* Akai MPC 2000 sampler */
	FORMAT_RF64  int = C.SF_FORMAT_RF64  /* RF64 WAV file */

	/* Subtypes from here on. */

	FORMAT_PCM_S8 int = C.SF_FORMAT_PCM_S8 /* Signed 8 bit data */
	FORMAT_PCM_16 int = C.SF_FORMAT_PCM_16 /* Signed 16 bit data */
	FORMAT_PCM_24 int = C.SF_FORMAT_PCM_24 /* Signed 24 bit data */
	FORMAT_PCM_32 int = C.SF_FORMAT_PCM_32 /* Signed 32 bit data */

	FORMAT_PCM_U8 int = C.SF_FORMAT_PCM_U8 /* Unsigned 8 bit data (WAV and RAW only) */

	FORMAT_FLOAT  int = C.SF_FORMAT_FLOAT  /* 32 bit float data */
	FORMAT_DOUBLE int = C.SF_FORMAT_DOUBLE /* 64 bit float data */

	FORMAT_ULAW      int = C.SF_FORMAT_ULAW      /* U-Law encoded. */
	FORMAT_ALAW      int = C.SF_FORMAT_ALAW      /* A-Law encoded. */
	FORMAT_IMA_ADPCM int = C.SF_FORMAT_IMA_ADPCM /* IMA ADPCM. */
	FORMAT_MS_ADPCM  int = C.SF_FORMAT_MS_ADPCM  /* Microsoft ADPCM. */

	FORMAT_GSM610    int = C.SF_FORMAT_GSM610    /* GSM 6.10 encoding. */
	FORMAT_VOX_ADPCM int = C.SF_FORMAT_VOX_ADPCM /* Oki Dialogic ADPCM encoding. */

	FORMAT_G721_32 int = C.SF_FORMAT_G721_32 /* 32kbs G721 ADPCM encoding. */
	FORMAT_G723_24 int = C.SF_FORMAT_G723_24 /* 24kbs G723 ADPCM encoding. */
	FORMAT_G723_40 int = C.SF_FORMAT_G723_40 /* 40kbs G723 ADPCM encoding. */

	FORMAT_DWVW_12 int = C.SF_FORMAT_DWVW_12 /* 12 bit Delta Width Variable Word encoding. */
	FORMAT_DWVW_16 int = C.SF_FORMAT_DWVW_16 /* 16 bit Delta Width Variable Word encoding. */
	FORMAT_DWVW_24 int = C.SF_FORMAT_DWVW_24 /* 24 bit Delta Width Variable Word encoding. */
	FORMAT_DWVW_N  int = C.SF_FORMAT_DWVW_N  /* N bit Delta Width Variable Word encoding. */

	FORMAT_DPCM_8  int = C.SF_FORMAT_DPCM_8  /* 8 bit differential PCM (XI only) */
	FORMAT_DPCM_16 int = C.SF_FORMAT_DPCM_16 /* 16 bit differential PCM (XI only) */

	FORMAT_VORBIS int = C.SF_FORMAT_VORBIS /* Xiph Vorbis encoding. */

	/* Endian-ness options. */

	ENDIAN_FILE   int = C.SF_ENDIAN_FILE   /* Default file endian-ness. */
	ENDIAN_LITTLE int = C.SF_ENDIAN_LITTLE /* Force little endian-ness. */
	ENDIAN_BIG    int = C.SF_ENDIAN_BIG    /* Force big endian-ness. */
	ENDIAN_CPU    int = C.SF_ENDIAN_CPU    /* Force CPU endian-ness. */

	// FIXME Usefull ?
	FORMAT_SUBMASK  int = C.SF_FORMAT_SUBMASK
	FORMAT_TYPEMASK int = C.SF_FORMAT_TYPEMASK
	FORMAT_ENDMASK  int = C.SF_FORMAT_ENDMASK
)

func SndFileOpen(path string, mode int, info *Info) (*SndFile, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	sndFile := &SndFile{path: path}
	if info != nil {
		sndFile.info = *info
	}
	sndFile.handle = C.sf_open(cPath, C.int(mode), (*C.SF_INFO)(&sndFile.info))
	if sndFile.handle == nil {
		errorMessage := fmt.Sprintf("Can't open SndFile : %v", C.GoString(C.sf_strerror(sndFile.handle)))
		return nil, errors.New(errorMessage)
	}

	return sndFile, nil
}
