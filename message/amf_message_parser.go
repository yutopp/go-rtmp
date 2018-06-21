//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"errors"
	"log"
)

type amfMessageParserFunc func(d AMFDecoder, name string, v *AMFConvertible) error

func parseAMFMessage(d AMFDecoder, name string, v *AMFConvertible) error {
	switch name {
	case "onMetaData":
		var metadata map[string]interface{}
		if err := d.Decode(&metadata); err != nil {
			return err
		}

		var data NetStreamOnMetaData
		if err := data.FromArgs(metadata); err != nil {
			return err
		}

		*v = &data

	case "@setDataFrame":
		// TODO: implement
		log.Println("Ignored data message: @setDataFrame")

	case "connect":
		var object map[string]interface{}
		if err := d.Decode(&object); err != nil {
			return err
		}
		log.Printf("command: object = %+v", object)

		var cmd NetConnectionConnect
		if err := cmd.FromArgs(object); err != nil {
			return err
		}

		*v = &cmd

	case "releaseStream":
		log.Printf("ignored releaseStream")

	case "createStream":
		var object interface{}
		if err := d.Decode(&object); err != nil {
			return err
		}
		log.Printf("createStream: object = %+v", object)

		var cmd NetConnectionCreateStream
		if err := cmd.FromArgs(object); err != nil {
			return err
		}

		*v = &cmd

	case "publish":
		var commandObject interface{}
		if err := d.Decode(&commandObject); err != nil {
			return err
		}
		var publishingName string
		if err := d.Decode(&publishingName); err != nil {
			return err
		}
		var publishingType string
		if err := d.Decode(&publishingType); err != nil {
			return err
		}

		var cmd NetStreamPublish
		if err := cmd.FromArgs(commandObject, publishingName, publishingType); err != nil {
			return err
		}
		*v = &cmd

	case "FCPublish":
		log.Printf("Ignored FCPublish")

	case "_result":
		// TODO: implement
		log.Println("Ignored _result")

	default:
		return errors.New("Not supported amf packed message: " + name)
	}

	return nil
}
