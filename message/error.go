//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"fmt"
)

type UnknownDataBodyDecodeError struct {
	Name string
	Objs []interface{}
}

func (e *UnknownDataBodyDecodeError) Error() string {
	return fmt.Sprintf("UnknownDataBodyDecodeError: Name = %s, Objs = %+v", e.Name, e.Objs)
}

type UnknownCommandBodyDecodeError struct {
	Name          string
	TransactionID int64
	Objs          []interface{}
}

func (e *UnknownCommandBodyDecodeError) Error() string {
	return fmt.Sprintf("UnknownCommandMessageDecodeError: Name = %s, TransactionID = %d, Objs = %+v",
		e.Name,
		e.TransactionID,
		e.Objs,
	)
}
