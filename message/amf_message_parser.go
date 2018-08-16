//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"bytes"
	"github.com/pkg/errors"
	"io"
)

type amfMessageParserFunc func(r io.Reader, d AMFDecoder, name string, v *AMFConvertible) error

func parseAMFMessage(r io.Reader, d AMFDecoder, name string, v *AMFConvertible) error {
	switch name {
	case "connect":
		var object map[string]interface{}
		if err := d.Decode(&object); err != nil {
			return errors.Wrap(err, "Failed to decode 'command' args[0]")
		}

		var cmd NetConnectionConnect
		if err := cmd.FromArgs(object); err != nil {
			return errors.Wrap(err, "Failed to reconstruct 'command'")
		}

		*v = &cmd

	case "createStream":
		var object interface{}
		if err := d.Decode(&object); err != nil {
			return errors.Wrap(err, "Failed to decode 'createStream' args[0]")
		}

		var cmd NetConnectionCreateStream
		if err := cmd.FromArgs(object); err != nil {
			return errors.Wrap(err, "Failed to reconstruct 'createStream'")
		}

		*v = &cmd

	case "deleteStream":
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

	case "publish":
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

	case "play":
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
			return errors.Wrap(err, "Failed to decode 'play' args[2]")
		}

		var cmd NetStreamPlay
		if err := cmd.FromArgs(commandObject, streamName, start); err != nil {
			return errors.Wrap(err, "Failed to reconstruct 'play'")
		}

		*v = &cmd

	case "releaseStream":
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

	case "FCPublish":
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

	case "FCUnpublish":
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

	case "@setDataFrame":
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, r); err != nil {
			return errors.Wrap(err, "Failed to decode '@setDataFrame' args[0]")
		}

		var cmd NetStreamSetDataFrame
		if err := cmd.FromArgs(buf.Bytes()); err != nil {
			return errors.Wrap(err, "Failed to reconstruct '@setDataFrame'")
		}

		*v = &cmd

	case "getStreamLength":
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

	case "ping": // NLE
		var commandObject interface{} // maybe nil
		if err := d.Decode(&commandObject); err != nil {
			return errors.Wrap(err, "Failed to decode 'ping' args[0]")
		}

		var cmd NetStreamPing
		if err := cmd.FromArgs(commandObject); err != nil {
			return errors.Wrap(err, "Failed to reconstruct 'ping'")
		}

		*v = &cmd

	case "closeStream":
		var commandObject interface{} // maybe nil
		if err := d.Decode(&commandObject); err != nil {
			return errors.Wrap(err, "Failed to decode 'closeStream' args[0]")
		}

		var cmd NetStreamCloseStream
		if err := cmd.FromArgs(commandObject); err != nil {
			return errors.Wrap(err, "Failed to reconstruct 'closeStream'")
		}

		*v = &cmd

	default:
		objs := make([]interface{}, 0)
		for {
			var tmp interface{}
			if err := d.Decode(&tmp); err != nil {
				break
			}
			objs = append(objs, tmp)
		}

		return &UnknownAMFParseError{
			Name: name,
			Objs: objs,
		}
	}

	return nil
}
