//
// Copyright (c) 2018- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

type NetStreamPublish struct {
	CommandObject  interface{}
	PublishingName string
	PublishingType string
}

type NetStreamOnStatus struct {
	CommandObject interface{}
	InfoObject    interface{}
}

type NetStreamOnMetaData struct {
	MetaData map[string]interface{} // TODO: to more detailed data
}
