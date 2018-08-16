//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

type EncodingType uint8

const (
	EncodingTypeAMF0 EncodingType = 0
	EncodingTypeAMF3 EncodingType = 3
)

type AMFConvertible interface {
	FromArgs(args ...interface{}) error
	ToArgs(ty EncodingType) ([]interface{}, error)
}

type AMFDecoder interface {
	Decode(interface{}) error
}

type AMFEncoder interface {
	Encode(interface{}) error
}
