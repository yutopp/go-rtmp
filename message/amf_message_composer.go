//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import ()

type amfMessageComposerFunc func(e AMFEncoder, v AMFConvertible) error

func composeAMFMessage(e AMFEncoder, v AMFConvertible) error {
	if v == nil {
		return nil // Do nothing
	}

	args, err := v.ToArgs()
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
