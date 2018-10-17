//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package rtmp

import (
	"encoding/binary"
	"fmt"
	"io"
)

type chunkBasicHeader struct {
	fmt           byte
	chunkStreamID int /* [0, 65599] */
}

func decodeChunkBasicHeader(r io.Reader, buf []byte, bh *chunkBasicHeader) error {
	if buf == nil || len(buf) < 3 {
		buf = make([]byte, 3)
	}

	if _, err := io.ReadAtLeast(r, buf[:1], 1); err != nil {
		return err
	}

	fmtTy := (buf[0] & 0xc0) >> 6 // 0b11000000 >> 6
	csID := int(buf[0] & 0x3f)    // 0b00111111

	switch csID {
	case 0:
		if _, err := io.ReadAtLeast(r, buf[1:2], 1); err != nil {
			return err
		}
		csID = int(buf[1]) + 64

	case 1:
		if _, err := io.ReadAtLeast(r, buf[1:], 2); err != nil {
			return err
		}
		csID = int(buf[2])*256 + int(buf[1]) + 64
	}

	bh.fmt = fmtTy
	bh.chunkStreamID = csID

	return nil
}

func encodeChunkBasicHeader(w io.Writer, mh *chunkBasicHeader) error {
	buf := make([]byte, 3)
	buf[0] = byte(mh.fmt&0x03) << 6 // 0b00000011 << 6

	switch {
	case mh.chunkStreamID >= 2 && mh.chunkStreamID <= 63:
		buf[0] |= byte(mh.chunkStreamID & 0x3f) // 0x00111111
		_, err := w.Write(buf[:1])              // TODO: should check length?
		return err

	case mh.chunkStreamID >= 64 && mh.chunkStreamID <= 319:
		buf[0] |= byte(0 & 0x3f) // 0x00111111
		buf[1] = byte(mh.chunkStreamID - 64)
		_, err := w.Write(buf[:2]) // TODO: should check length?
		return err

	case mh.chunkStreamID >= 320 && mh.chunkStreamID <= 65599:
		buf[0] |= byte(1 & 0x3f) // 0x00111111
		buf[1] = byte(int(mh.chunkStreamID-64) % 256)
		buf[2] = byte(int(mh.chunkStreamID-64) / 256)
		_, err := w.Write(buf) // TODO: should check length?
		return err

	default:
		return fmt.Errorf("Chunk stream id is out of range: %d must be in range [2, 65599]", mh.chunkStreamID)
	}
}

type chunkMessageHeader struct {
	timestamp       uint32 // fmt = 0
	timestampDelta  uint32 // fmt = 1 | 2
	messageLength   uint32 // fmt = 0 | 1
	messageTypeID   byte   // fmt = 0 | 1
	messageStreamID uint32 // fmt = 0
}

func decodeChunkMessageHeader(r io.Reader, fmt byte, buf []byte, mh *chunkMessageHeader) error {
	if buf == nil || len(buf) < 11 {
		buf = make([]byte, 11)
	}
	cache32bits := make([]byte, 4)

	switch fmt {
	case 0:
		if _, err := io.ReadAtLeast(r, buf[:11], 11); err != nil {
			return err
		}

		copy(cache32bits[1:], buf[0:3]) // 24bits BE
		mh.timestamp = binary.BigEndian.Uint32(cache32bits)
		copy(cache32bits[1:], buf[3:6]) // 24bits BE
		mh.messageLength = binary.BigEndian.Uint32(cache32bits)
		mh.messageTypeID = buf[6]                                  // 8bits
		mh.messageStreamID = binary.LittleEndian.Uint32(buf[7:11]) // 32bits

		if mh.timestamp == 0xffffff {
			_, err := io.ReadAtLeast(r, cache32bits, 4)
			if err != nil {
				return err
			}
			mh.timestamp = binary.BigEndian.Uint32(cache32bits)
		}

	case 1:
		if _, err := io.ReadAtLeast(r, buf[:7], 7); err != nil {
			return err
		}

		copy(cache32bits[1:], buf[0:3]) // 24bits BE
		mh.timestampDelta = binary.BigEndian.Uint32(cache32bits)
		copy(cache32bits[1:], buf[3:6]) // 24bits BE
		mh.messageLength = binary.BigEndian.Uint32(cache32bits)
		mh.messageTypeID = buf[6] // 8bits

		if mh.timestampDelta == 0xffffff {
			_, err := io.ReadAtLeast(r, cache32bits, 4)
			if err != nil {
				return err
			}
			mh.timestampDelta = binary.BigEndian.Uint32(cache32bits)
		}

	case 2:
		if _, err := io.ReadAtLeast(r, buf[:3], 3); err != nil {
			return err
		}

		copy(cache32bits[1:], buf[0:3]) // 24bits BE
		mh.timestampDelta = binary.BigEndian.Uint32(cache32bits)

		if mh.timestampDelta == 0xffffff {
			_, err := io.ReadAtLeast(r, cache32bits, 4)
			if err != nil {
				return err
			}
			mh.timestampDelta = binary.BigEndian.Uint32(cache32bits)
		}

	case 3:
		// DO NOTHING

	default:
		panic("Unexpected fmt")
	}

	return nil
}

func encodeChunkMessageHeader(w io.Writer, fmt byte, mh *chunkMessageHeader) error {
	buf := make([]byte, 11+4)
	cache32bits := make([]byte, 4)
	ext := false

	switch fmt {
	case 0:
		buflen := 11
		ts := mh.timestamp
		if ts >= 0xffffff {
			ts = 0xffffff
			ext = true
			buflen += 4
		}

		binary.BigEndian.PutUint32(cache32bits, ts)
		copy(buf[0:3], cache32bits[1:]) // 24 bits BE
		binary.BigEndian.PutUint32(cache32bits, mh.messageLength)
		copy(buf[3:6], cache32bits[1:]) // 24 bits BE
		buf[6] = mh.messageTypeID       // 8bits
		binary.LittleEndian.PutUint32(buf[7:11], mh.messageStreamID)

		if ext {
			binary.BigEndian.PutUint32(buf[11:], mh.timestamp)
		}

		_, err := w.Write(buf[:buflen]) // TODO: should check length?
		return err

	case 1:
		buflen := 7
		td := mh.timestampDelta
		if td >= 0xffffff {
			td = 0xffffff
			ext = true
			buflen += 4
		}

		binary.BigEndian.PutUint32(cache32bits, td)
		copy(buf[0:3], cache32bits[1:]) // 24bits BE
		binary.BigEndian.PutUint32(cache32bits, mh.messageLength)
		copy(buf[3:6], cache32bits[1:]) // 24bits BE
		buf[6] = mh.messageTypeID       // 8bits

		if ext {
			binary.BigEndian.PutUint32(buf[7:], mh.timestampDelta)
		}

		_, err := w.Write(buf[:buflen]) // TODO: should check length?
		return err

	case 2:
		buflen := 3
		td := mh.timestampDelta
		if td >= 0xffffff {
			td = 0xffffff
			ext = true
			buflen += 4
		}

		binary.BigEndian.PutUint32(cache32bits, td)
		copy(buf[0:3], cache32bits[1:]) // 24bits BE

		if ext {
			binary.BigEndian.PutUint32(buf[3:], mh.timestampDelta)
		}

		_, err := w.Write(buf[:buflen]) // TODO: should check length?
		return err

	case 3:
		// DO NOTHING
		return nil

	default:
		panic("Unexpected fmt")
	}
}
