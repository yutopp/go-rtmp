//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"github.com/pkg/errors"
	"github.com/yutopp/go-amf0"
	"io"
)

type BodyEncoder struct {
	writer     io.Writer
	amfEnc     AMFEncoder
	Value      AMFConvertible
	MsgEncoder func(w io.Writer, e AMFEncoder, v AMFConvertible) error
}

func (be *BodyEncoder) Encode() error {
	return be.MsgEncoder(be.writer, be.amfEnc, be.Value)
}

type BodyEncoderFunc func(r io.Reader, e AMFDecoder, v *AMFConvertible) error

func EncodeBodyAnyValues(w io.Writer, e AMFEncoder, v AMFConvertible) error {
	if v == nil {
		return nil // Do nothing
	}

	var amfTy EncodingType
	switch e.(type) {
	case *amf0.Encoder:
		amfTy = EncodingTypeAMF0
	default:
		return errors.Errorf("Unsupported AMF Encoder: Type = %T", e)
	}

	args, err := v.ToArgs(amfTy)
	if err != nil {
		return err
	}

	for _, arg := range args {
		if err := e.Encode(arg); err != nil {
			return err
		}
	}

	return nil
}
