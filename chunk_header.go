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

func decodeChunkBasicHeader(r io.Reader, bh *chunkBasicHeader) error {
	buf := make([]byte, 3)
	_, err := io.ReadAtLeast(r, buf[:1], 1)
	if err != nil {
		return err
	}

	fmt := (buf[0] & 0xC0) >> 6 // 0b11000000 >> 6
	csID := int(buf[0] & 0x3f)  // 0b00111111

	// TODO: implement
	switch csID {
	case 0:
		panic("not implemented")
	case 1:
		panic("not implemented")
	}

	bh.fmt = fmt
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
		panic("not implemented")
	case mh.chunkStreamID >= 320 && mh.chunkStreamID <= 65599:
		panic("not implemented")
	default:
		panic(fmt.Sprintf("unexpected chunk stream id: %d", mh.chunkStreamID))
	}
}

type chunkMessageHeader struct {
	timestamp       uint32 // fmt = 0
	timestampDelta  uint32 // fmt = 1 | 2
	messageLength   uint32 // fmt = 0 | 1
	messageTypeID   byte   // fmt = 0 | 1
	messageStreamID uint32 // fmt = 0
}

func decodeChunkMessageHeader(r io.Reader, fmt byte, mh *chunkMessageHeader) error {
	cache32bits := make([]byte, 4)

	switch fmt {
	case 0:
		buf := make([]byte, 11)
		_, err := io.ReadAtLeast(r, buf, len(buf))
		if err != nil {
			return err
		}

		copy(cache32bits[1:], buf[0:3]) // 24bits BE
		mh.timestamp = binary.BigEndian.Uint32(cache32bits)
		copy(cache32bits[1:], buf[3:6]) // 24bits BE
		mh.messageLength = binary.BigEndian.Uint32(cache32bits)
		mh.messageTypeID = buf[6]                                  // 8bits
		mh.messageStreamID = binary.LittleEndian.Uint32(buf[7:11]) // 32bits

		// TODO: extended timestamp
		if mh.timestamp == 0xffffff {
			panic("not implemented extended timestamp")
		}

	case 1:
		buf := make([]byte, 7)
		_, err := io.ReadAtLeast(r, buf, len(buf))
		if err != nil {
			return err
		}

		copy(cache32bits[1:], buf[0:3]) // 24bits BE
		mh.timestampDelta = binary.BigEndian.Uint32(cache32bits)
		copy(cache32bits[1:], buf[3:6]) // 24bits BE
		mh.messageLength = binary.BigEndian.Uint32(cache32bits)
		mh.messageTypeID = buf[6] // 8bits

		// TODO: extended timestamp delta
		if mh.timestampDelta == 0xffffff {
			panic("not implemented extended timestamp delta")
		}

	case 2:
		buf := make([]byte, 3)
		_, err := io.ReadAtLeast(r, buf, len(buf))
		if err != nil {
			return err
		}

		copy(cache32bits[1:], buf[0:3]) // 24bits BE
		mh.timestampDelta = binary.BigEndian.Uint32(cache32bits)

		// TODO: extended timestamp delta
		if mh.timestampDelta == 0xffffff {
			panic("not implemented extended timestamp delta")
		}

	case 3:
		// DO NOTHING

	default:
		panic("unexpected fmt")
	}

	return nil
}

func encodeChunkMessageHeader(w io.Writer, fmt byte, mh *chunkMessageHeader) error {
	cache32bits := make([]byte, 4)

	switch fmt {
	case 0:
		// TODO: support extended timestamp
		buf := make([]byte, 11)
		binary.BigEndian.PutUint32(cache32bits, mh.timestamp)
		copy(buf[0:3], cache32bits[1:]) // 24 bits BE
		binary.BigEndian.PutUint32(cache32bits, mh.messageLength)
		copy(buf[3:6], cache32bits[1:]) // 24 bits BE
		buf[6] = mh.messageTypeID       // 8bits
		binary.LittleEndian.PutUint32(buf[7:11], mh.messageStreamID)

		_, err := w.Write(buf) // TODO: should check length?
		return err

	case 1:
		// TODO: extended timestamp delta
		buf := make([]byte, 7)
		binary.BigEndian.PutUint32(cache32bits, mh.timestampDelta)
		copy(buf[0:3], cache32bits[1:]) // 24bits BE
		binary.BigEndian.PutUint32(cache32bits, mh.messageLength)
		copy(buf[3:6], cache32bits[1:]) // 24bits BE
		buf[6] = mh.messageTypeID       // 8bits

		_, err := w.Write(buf) // TODO: should check length?
		return err

	case 2:
		// TODO: extended timestamp delta
		buf := make([]byte, 3)
		binary.BigEndian.PutUint32(cache32bits, mh.timestampDelta)
		copy(buf[0:3], cache32bits[1:]) // 24bits BE

		_, err := w.Write(buf) // TODO: should check length?
		return err

	case 3:
		// DO NOTHING
		return nil

	default:
		panic("unexpected fmt")
	}
}
