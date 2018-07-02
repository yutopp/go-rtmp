//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

type AMFType uint8

const (
	AMFType0 AMFType = iota
	AMDType3
)

type AMFConvertible interface {
	FromArgs(args ...interface{}) error
	ToArgs(ty AMFType) ([]interface{}, error)
}

type AMFDecoder interface {
	Decode(interface{}) error
}

type AMFEncoder interface {
	Encode(interface{}) error
}
