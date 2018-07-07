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

type UnknownAMFParseError struct {
	Name string
	Objs []interface{}
}

func (e *UnknownAMFParseError) Error() string {
	return fmt.Sprintf("UnknownAMFParseError: Name = %s, Objs = %+v", e.Name, e.Objs)
}
