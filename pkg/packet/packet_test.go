package packet

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromString(t *testing.T) {
	type test struct {
		name            string
		raw             string
		expectedPacket  *GamespyPacket
		wantErrContains string
	}

	tests := []test{
		{
			name: "parses challenge prompt packet",
			raw:  "\\lc\\1\\challenge\\TcP1s0FtTB\\id\\1\\final\\",
			expectedPacket: &GamespyPacket{
				elements: []KeyValuePair{
					{
						Key:   "lc",
						Value: "1",
					},
					{
						Key:   "challenge",
						Value: "TcP1s0FtTB",
					},
					{
						Key:   "id",
						Value: "1",
					},
				},
			},
		},
		{
			name: "parses login request packet",
			raw:  "\\login\\\\challenge\\YJk5UFExKBwn0PEpOpinWHsRCDcfejyJ\\uniquenick\\some-nick\\response\\638ac6fccc7f5a79f25b82132c87572b\\port\\2475\\productid\\10493\\gamename\\battlefield2\\namespaceid\\12\\sdkrevision\\3\\id\\1\\final\\",
			expectedPacket: &GamespyPacket{
				elements: []KeyValuePair{
					{
						Key:   "login",
						Value: "",
					},
					{
						Key:   "challenge",
						Value: "YJk5UFExKBwn0PEpOpinWHsRCDcfejyJ",
					},
					{
						Key:   "uniquenick",
						Value: "some-nick",
					},
					{
						Key:   "response",
						Value: "638ac6fccc7f5a79f25b82132c87572b",
					},
					{
						Key:   "port",
						Value: "2475",
					},
					{
						Key:   "productid",
						Value: "10493",
					},
					{
						Key:   "gamename",
						Value: "battlefield2",
					},
					{
						Key:   "namespaceid",
						Value: "12",
					},
					{
						Key:   "sdkrevision",
						Value: "3",
					},
					{
						Key:   "id",
						Value: "1",
					},
				},
			},
		},
		{
			name: "parses login response packet",
			raw:  "\\lc\\2\\sesskey\\19745\\proof\\8c628092b8ac503e184e68c96d27e758\\userid\\123\\profileid\\456\\uniquenick\\some-nick\\lt\\SIYCIWSEARGXPMEUJRBKKE__\\id\\1\\final\\",
			expectedPacket: &GamespyPacket{
				elements: []KeyValuePair{
					{
						Key:   "lc",
						Value: "2",
					},
					{
						Key:   "sesskey",
						Value: "19745",
					},
					{
						Key:   "proof",
						Value: "8c628092b8ac503e184e68c96d27e758",
					},
					{
						Key:   "userid",
						Value: "123",
					},
					{
						Key:   "profileid",
						Value: "456",
					},
					{
						Key:   "uniquenick",
						Value: "some-nick",
					},
					{
						Key:   "lt",
						Value: "SIYCIWSEARGXPMEUJRBKKE__",
					},
					{
						Key:   "id",
						Value: "1",
					},
				},
			},
		},
		{
			name:            "error for packet not starting with \\",
			raw:             "key\\value\\final\\",
			wantErrContains: "gamespy packet string is malformed",
		},
		{
			name:            "error for packet not ending with \\final\\",
			raw:             "\\key\\value",
			wantErrContains: "gamespy packet string is malformed",
		},
		{
			name:            "error for packet containing uneven number of elements",
			raw:             "\\key\\value\\key-without-value\\final\\",
			wantErrContains: "gamespy packet string contains key without corresponding value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WHEN
			packet, err := FromString(tt.raw)

			// THEN
			if tt.wantErrContains != "" {
				require.ErrorContains(t, err, tt.wantErrContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedPacket, packet)
			}
		})
	}
}

func TestGamespyPacket_Map(t *testing.T) {
	type test struct {
		name        string
		packet      *GamespyPacket
		expectedMap map[string]string
	}

	tests := []test{
		{
			name: "converts packet to map",
			packet: &GamespyPacket{
				elements: []KeyValuePair{
					{
						Key:   "key",
						Value: "value",
					},
				},
			},
			expectedMap: map[string]string{
				"key": "value",
			},
		},
		{
			name: "converts empty element slice packet to map",
			packet: &GamespyPacket{
				elements: []KeyValuePair{},
			},
			expectedMap: map[string]string{},
		},
		{
			name:        "converts element nil slice packet to map",
			packet:      &GamespyPacket{},
			expectedMap: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WHEN
			actual := tt.packet.Map()

			// THEN
			assert.Equal(t, tt.expectedMap, actual)
		})
	}
}

func TestGamespyPacket_Bytes(t *testing.T) {
	type test struct {
		name          string
		packet        *GamespyPacket
		expectedBytes []byte
	}

	tests := []test{
		{
			name: "converts packet to string",
			packet: &GamespyPacket{
				elements: []KeyValuePair{
					{
						Key:   "key",
						Value: "value",
					},
				},
			},
			expectedBytes: []byte("\\key\\value\\final\\"),
		},
		{
			name: "converts empty element slice packet to string",
			packet: &GamespyPacket{
				elements: []KeyValuePair{},
			},
			expectedBytes: []byte("\\\\final\\"),
		},
		{
			name:          "converts element nil slice packet to string",
			packet:        &GamespyPacket{},
			expectedBytes: []byte("\\\\final\\"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WHEN
			actual := tt.packet.Bytes()

			// THEN
			assert.Equal(t, tt.expectedBytes, actual)
		})
	}
}
