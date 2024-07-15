//
// Copyright (c) 2023- yutopp (yutopp@gmail.com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at  https://www.boost.org/LICENSE_1_0.txt)
//

package message

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodeCmdMessageOnStatus(t *testing.T) {
	tests := []struct {
		name    string
		decoder BodyDecoderFunc
		body    AMFConvertible
	}{
		{
			name:    "NetConnectionConnect",
			decoder: DecodeBodyConnect,
			body: &NetConnectionConnect{
				Command: NetConnectionConnectCommand{
					App:            "app",
					Type:           "type",
					FlashVer:       "flashVer",
					TCURL:          "tcUrl",
					Fpad:           true,
					Capabilities:   1,
					AudioCodecs:    2,
					VideoCodecs:    3,
					VideoFunction:  4,
					ObjectEncoding: EncodingTypeAMF3,
				},
			},
		},
		{
			name:    "NetConnectionConnectResult",
			decoder: DecodeBodyConnectResult,
			body: &NetConnectionConnectResult{
				Properties: NetConnectionConnectResultProperties{
					FMSVer:       "FMS/3,0,1,123",
					Capabilities: 31,
					Mode:         1,
				},
				Information: NetConnectionConnectResultInformation{
					Level:       "status",
					Code:        NetConnectionConnectCodeSuccess,
					Description: "Connection succeeded",
					Data:        map[string]interface{}{},
				},
			},
		},
		{
			name:    "NetConnectionConnectResult with data",
			decoder: DecodeBodyConnectResult,
			body: &NetConnectionConnectResult{
				Properties: NetConnectionConnectResultProperties{
					FMSVer:       "FMS/3,0,1,123",
					Capabilities: 31,
					Mode:         1,
				},
				Information: NetConnectionConnectResultInformation{
					Level:       "status",
					Code:        NetConnectionConnectCodeSuccess,
					Description: "Connection succeeded",
					Data: map[string]interface{}{
						"test": "test",
					},
				},
			},
		},
		{
			name:    "NetConnectionCreateStream",
			decoder: DecodeBodyCreateStream,
			body:    &NetConnectionCreateStream{},
		},
		{
			name:    "NetConnectionCreateStreamResult",
			decoder: DecodeBodyCreateStreamResult,
			body: &NetConnectionCreateStreamResult{
				StreamID: 1,
			},
		},
		{
			name:    "NetConnectionReleaseStream",
			decoder: DecodeBodyReleaseStream,
			body: &NetConnectionReleaseStream{
				StreamName: "stream",
			},
		},
		{
			name:    "NetStreamOnStatus",
			decoder: DecodeBodyOnStatus,
			body: &NetStreamOnStatus{
				InfoObject: NetStreamOnStatusInfoObject{
					Level:           NetStreamOnStatusLevelStatus,
					Code:            NetStreamOnStatusCodePlayStart,
					Description:     "abc",
					ExtraProperties: map[string]interface{}{},
				},
			},
		},
		{
			name:    "NetStreamOnStatus with extra properties",
			decoder: DecodeBodyOnStatus,
			body: &NetStreamOnStatus{
				InfoObject: NetStreamOnStatusInfoObject{
					Level:           NetStreamOnStatusLevelStatus,
					Code:            NetStreamOnStatusCodePlayStart,
					Description:     "abc",
					ExtraProperties: map[string]interface{}{"foo": "bar"},
				},
			},
		},
	}

	for _, test := range tests {
		test := test // capture

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			amfTy := EncodingTypeAMF0
			buf := new(bytes.Buffer)

			// object to bytes(AMF0)
			amfEnc := NewAMFEncoder(buf, amfTy)
			err := EncodeBodyAnyValues(amfEnc, test.body)
			require.Nil(t, err)

			// bytes(AMF0) to object
			amfDec := NewAMFDecoder(buf, amfTy)
			var v AMFConvertible
			err = test.decoder(buf, amfDec, &v)
			require.Nil(t, err)

			require.Equal(t, test.body, v)
		})
	}
}
