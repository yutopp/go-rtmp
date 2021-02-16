//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"bytes"
	"io"

	"github.com/pkg/errors"
)

type BodyDecoderFunc func(r io.Reader, e AMFDecoder, v *AMFConvertible) error

var DataBodyDecoders = map[string]BodyDecoderFunc{
	"@setDataFrame": DecodeBodyAtSetDataFrame,
}

func DataBodyDecoderFor(name string) BodyDecoderFunc {
	dec, ok := DataBodyDecoders[name]
	if ok {
		return dec
	}

	return func(_ io.Reader, d AMFDecoder, _ *AMFConvertible) error {
		objs := make([]interface{}, 0)
		for {
			var tmp interface{}
			if err := d.Decode(&tmp); err != nil {
				break
			}
			objs = append(objs, tmp)
		}

		return &UnknownDataBodyDecodeError{
			Name: name,
			Objs: objs,
		}
	}
}

func DecodeBodyAtSetDataFrame(r io.Reader, _ AMFDecoder, v *AMFConvertible) error {
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r); err != nil {
		return errors.Wrap(err, "Failed to decode '@setDataFrame' args[0]")
	}

	var cmd NetStreamSetDataFrame
	if err := cmd.FromArgs(buf.Bytes()); err != nil {
		return errors.Wrap(err, "Failed to reconstruct '@setDataFrame'")
	}

	*v = &cmd

	return nil
}

var CmdBodyDecoders = map[string]BodyDecoderFunc{
	"connect":         DecodeBodyConnect,
	"createStream":    DecodeBodyCreateStream,
	"deleteStream":    DecodeBodyDeleteStream,
	"publish":         DecodeBodyPublish,
	"play":            DecodeBodyPlay,
	"releaseStream":   DecodeBodyReleaseStream,
	"FCPublish":       DecodeBodyFCPublish,
	"FCUnpublish":     DecodeBodyFCUnpublish,
	"getStreamLength": DecodeBodyGetStreamLength,
	"ping":            DecodeBodyPing,
	"closeStream":     DecodeBodyCloseStream,
}

func CmdBodyDecoderFor(name string, transactionID int64) BodyDecoderFunc {
	dec, ok := CmdBodyDecoders[name]
	if ok {
		return dec
	}

	// TODO: support result

	return func(_ io.Reader, d AMFDecoder, _ *AMFConvertible) error {
		objs := make([]interface{}, 0)
		for {
			var tmp interface{}
			if err := d.Decode(&tmp); err != nil {
				break
			}
			objs = append(objs, tmp)
		}

		return &UnknownCommandBodyDecodeError{
			Name:          name,
			TransactionID: transactionID,
			Objs:          objs,
		}
	}
}

func DecodeBodyConnect(_ io.Reader, d AMFDecoder, v *AMFConvertible) error {
	var object map[string]interface{}
	if err := d.Decode(&object); err != nil {
		return errors.Wrap(err, "Failed to decode 'connect' args[0]")
	}

	var cmd NetConnectionConnect
	if err := cmd.FromArgs(object); err != nil {
		return errors.Wrap(err, "Failed to reconstruct 'connect'")
	}

	*v = &cmd
	return nil
}

func DecodeBodyConnectResult(_ io.Reader, d AMFDecoder, v *AMFConvertible) error {
	var properties interface{}
	if err := d.Decode(&properties); err != nil {
		return errors.Wrap(err, "Failed to decode 'connect.result' args[0]")
	}

	var information interface{}
	if err := d.Decode(&information); err != nil {
		return errors.Wrap(err, "Failed to decode 'connect.result' args[1]")
	}

	var result NetConnectionConnectResult
	if err := result.FromArgs(properties, information); err != nil {
		return errors.Wrap(err, "Failed to reconstruct 'connect.result'")
	}

	*v = &result
	return nil
}

func DecodeBodyCreateStream(_ io.Reader, d AMFDecoder, v *AMFConvertible) error {
	var object interface{}
	if err := d.Decode(&object); err != nil {
		return errors.Wrap(err, "Failed to decode 'createStream' args[0]")
	}

	var cmd NetConnectionCreateStream
	if err := cmd.FromArgs(object); err != nil {
		return errors.Wrap(err, "Failed to reconstruct 'createStream'")
	}

	*v = &cmd
	return nil
}

func DecodeBodyCreateStreamResult(_ io.Reader, d AMFDecoder, v *AMFConvertible) error {
	var commandObject interface{} // maybe nil
	if err := d.Decode(&commandObject); err != nil {
		return errors.Wrap(err, "Failed to decode 'createStream.result' args[0]")
	}

	var streamID uint32
	if err := d.Decode(&streamID); err != nil {
		return errors.Wrap(err, "Failed to decode 'createStream.result' args[1]")
	}

	var data NetConnectionCreateStreamResult
	if err := data.FromArgs(commandObject, streamID); err != nil {
		return errors.Wrap(err, "Failed to reconstruct 'createStream.result'")
	}

	*v = &data
	return nil
}

func DecodeBodyDeleteStream(_ io.Reader, d AMFDecoder, v *AMFConvertible) error {
	var commandObject interface{} // maybe nil
	if err := d.Decode(&commandObject); err != nil {
		return errors.Wrap(err, "Failed to decode 'deleteStream' args[0]")
	}

	var streamID uint32
	if err := d.Decode(&streamID); err != nil {
		return errors.Wrap(err, "Failed to decode 'deleteStream' args[1]")
	}

	var data NetStreamDeleteStream
	if err := data.FromArgs(commandObject, streamID); err != nil {
		return errors.Wrap(err, "Failed to reconstruct 'deleteStream'")
	}

	*v = &data
	return nil
}

func DecodeBodyPublish(_ io.Reader, d AMFDecoder, v *AMFConvertible) error {
	var commandObject interface{}
	if err := d.Decode(&commandObject); err != nil {
		return errors.Wrap(err, "Failed to decode 'publish' args[0]")
	}
	var publishingName string
	if err := d.Decode(&publishingName); err != nil {
		return errors.Wrap(err, "Failed to decode 'publish' args[1]")
	}
	var publishingType string
	if err := d.Decode(&publishingType); err != nil {
		return errors.Wrap(err, "Failed to decode 'publish' args[2]")
	}

	var cmd NetStreamPublish
	if err := cmd.FromArgs(commandObject, publishingName, publishingType); err != nil {
		return errors.Wrap(err, "Failed to reconstruct 'publish'")
	}

	*v = &cmd
	return nil
}

func DecodeBodyPlay(_ io.Reader, d AMFDecoder, v *AMFConvertible) error {
	var commandObject interface{}
	if err := d.Decode(&commandObject); err != nil {
		return errors.Wrap(err, "Failed to decode 'play' args[0]")
	}
	var streamName string
	if err := d.Decode(&streamName); err != nil {
		return errors.Wrap(err, "Failed to decode 'play' args[1]")
	}
	var start int64
	if err := d.Decode(&start); err != nil {
		//
		// io.EOF occurs when the start position is not specified.
		//  'NetStream.play(streamName,null)'
		// set start to 0 to avoid it.
		//
		if err != io.EOF {
			return errors.Wrap(err, "Failed to decode 'play' args[2]")
		}
		start = 0
	}

	var cmd NetStreamPlay
	if err := cmd.FromArgs(commandObject, streamName, start); err != nil {
		return errors.Wrap(err, "Failed to reconstruct 'play'")
	}

	*v = &cmd
	return nil
}

func DecodeBodyReleaseStream(_ io.Reader, d AMFDecoder, v *AMFConvertible) error {
	var commandObject interface{} // maybe nil
	if err := d.Decode(&commandObject); err != nil {
		return errors.Wrap(err, "Failed to decode 'releaseStream' args[0]")
	}
	var streamName string
	if err := d.Decode(&streamName); err != nil {
		return errors.Wrap(err, "Failed to decode 'releaseStream' args[1]")
	}

	var cmd NetConnectionReleaseStream
	if err := cmd.FromArgs(commandObject, streamName); err != nil {
		return errors.Wrap(err, "Failed to reconstruct 'releaseStream'")
	}

	*v = &cmd
	return nil
}

func DecodeBodyFCPublish(_ io.Reader, d AMFDecoder, v *AMFConvertible) error {
	var commandObject interface{} // maybe nil
	if err := d.Decode(&commandObject); err != nil {
		return errors.Wrap(err, "Failed to decode 'FCPublish' args[0]")
	}
	var streamName string
	if err := d.Decode(&streamName); err != nil {
		return errors.Wrap(err, "Failed to decode 'FCPublish' args[1]")
	}

	var cmd NetStreamFCPublish
	if err := cmd.FromArgs(commandObject, streamName); err != nil {
		return errors.Wrap(err, "Failed to reconstruct 'FCPublish'")
	}

	*v = &cmd
	return nil
}

func DecodeBodyFCUnpublish(_ io.Reader, d AMFDecoder, v *AMFConvertible) error {
	var commandObject interface{} // maybe nil
	if err := d.Decode(&commandObject); err != nil {
		return errors.Wrap(err, "Failed to decode 'FCUnpublish' args[0]")
	}
	var streamName string
	if err := d.Decode(&streamName); err != nil {
		return errors.Wrap(err, "Failed to decode 'FCUnpublish' args[1]")
	}

	var cmd NetStreamFCUnpublish
	if err := cmd.FromArgs(commandObject, streamName); err != nil {
		return errors.Wrap(err, "Failed to reconstruct 'FCUnpublish'")
	}

	*v = &cmd
	return nil
}

func DecodeBodyGetStreamLength(_ io.Reader, d AMFDecoder, v *AMFConvertible) error {
	var commandObject interface{} // maybe nil
	if err := d.Decode(&commandObject); err != nil {
		return errors.Wrap(err, "Failed to decode 'getStreamLength' args[0]")
	}
	var streamName string
	if err := d.Decode(&streamName); err != nil {
		return errors.Wrap(err, "Failed to decode 'getStreamLength' args[1]")
	}

	var cmd NetStreamGetStreamLength
	if err := cmd.FromArgs(commandObject, streamName); err != nil {
		return errors.Wrap(err, "Failed to reconstruct 'getStreamLength'")
	}

	*v = &cmd
	return nil
}

func DecodeBodyPing(_ io.Reader, d AMFDecoder, v *AMFConvertible) error { // NLE
	var commandObject interface{} // maybe nil
	if err := d.Decode(&commandObject); err != nil {
		return errors.Wrap(err, "Failed to decode 'ping' args[0]")
	}

	var cmd NetStreamPing
	if err := cmd.FromArgs(commandObject); err != nil {
		return errors.Wrap(err, "Failed to reconstruct 'ping'")
	}

	*v = &cmd
	return nil
}

func DecodeBodyCloseStream(_ io.Reader, d AMFDecoder, v *AMFConvertible) error {
	var commandObject interface{} // maybe nil
	if err := d.Decode(&commandObject); err != nil {
		return errors.Wrap(err, "Failed to decode 'closeStream' args[0]")
	}

	var cmd NetStreamCloseStream
	if err := cmd.FromArgs(commandObject); err != nil {
		return errors.Wrap(err, "Failed to reconstruct 'closeStream'")
	}

	*v = &cmd

	return nil
}
