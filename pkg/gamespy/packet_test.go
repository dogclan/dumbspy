package gamespy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromBytes(t *testing.T) {
	type test struct {
		name            string
		bytes           []byte
		expectedPacket  *Packet
		wantErrContains string
	}

	tests := []test{
		{
			name:  "parses challenge prompt packet",
			bytes: []byte("\\lc\\1\\challenge\\TcP1s0FtTB\\id\\1\\final\\"),
			expectedPacket: &Packet{
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
			name:  "parses login request packet",
			bytes: []byte("\\login\\\\challenge\\YJk5UFExKBwn0PEpOpinWHsRCDcfejyJ\\uniquenick\\some-nick\\response\\638ac6fccc7f5a79f25b82132c87572b\\port\\2475\\productid\\10493\\gamename\\battlefield2\\namespaceid\\12\\sdkrevision\\3\\id\\1\\final\\"),
			expectedPacket: &Packet{
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
			name:  "parses login response packet",
			bytes: []byte("\\lc\\2\\sesskey\\19745\\proof\\8c628092b8ac503e184e68c96d27e758\\userid\\123\\profileid\\456\\uniquenick\\some-nick\\lt\\SIYCIWSEARGXPMEUJRBKKE__\\id\\1\\final\\"),
			expectedPacket: &Packet{
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
			name:  "parse nicks check response packet with single result",
			bytes: []byte("\\nr\\0\\nick\\a-nick\\uniquenick\\a-uniquenick\\ndone\\\\final\\"),
			expectedPacket: &Packet{
				elements: []KeyValuePair{
					{
						Key:   "nr",
						Value: "0",
					},
					{
						Key:   "nick",
						Value: "a-nick",
					},
					{
						Key:   "uniquenick",
						Value: "a-uniquenick",
					},
					{
						Key:   "ndone",
						Value: "",
					},
				},
			},
		},
		{
			name:  "parse nicks check response packet with multiple results",
			bytes: []byte("\\nr\\0\\nick\\a-nick\\uniquenick\\a-uniquenick\\nick\\b-nick\\uniquenick\\b-uniquenick\\ndone\\\\final\\"),
			expectedPacket: &Packet{
				elements: []KeyValuePair{
					{
						Key:   "nr",
						Value: "0",
					},
					{
						Key:   "nick",
						Value: "a-nick",
					},
					{
						Key:   "uniquenick",
						Value: "a-uniquenick",
					},
					{
						Key:   "nick",
						Value: "b-nick",
					},
					{
						Key:   "uniquenick",
						Value: "b-uniquenick",
					},
					{
						Key:   "ndone",
						Value: "",
					},
				},
			},
		},
		{
			name:            "error for packet not starting with \\",
			bytes:           []byte("key\\value\\final\\"),
			wantErrContains: "gamespy packet string is malformed",
		},
		{
			name:            "error for packet not ending with \\final\\",
			bytes:           []byte("\\key\\value"),
			wantErrContains: "gamespy packet string is malformed",
		},
		{
			name:            "error for packet containing uneven number of elements",
			bytes:           []byte("\\key\\value\\key-without-value\\final\\"),
			wantErrContains: "gamespy packet string contains key without corresponding value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WHEN
			packet, err := NewPacketFromBytes(tt.bytes)

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

func TestGamespyPacket_Set(t *testing.T) {
	t.Run("adds initial key", func(t *testing.T) {
		// GIVEN
		packet := new(Packet)

		// WHEN
		packet.Set("initial-key", "initial-value")

		// THEN
		assert.Equal(t, []KeyValuePair{
			{
				Key:   "initial-key",
				Value: "initial-value",
			},
		}, packet.elements)
	})

	t.Run("adds new key", func(t *testing.T) {
		// GIVEN
		packet := &Packet{
			elements: []KeyValuePair{
				{
					Key:   "original-key",
					Value: "original-value",
				},
			},
		}

		// WHEN
		packet.Set("new-key", "new-value")

		// THEN
		assert.Equal(t, []KeyValuePair{
			{
				Key:   "original-key",
				Value: "original-value",
			},
			{
				Key:   "new-key",
				Value: "new-value",
			},
		}, packet.elements)
	})

	t.Run("updates existing key", func(t *testing.T) {
		// GIVEN
		packet := &Packet{
			elements: []KeyValuePair{
				{
					Key:   "original-key",
					Value: "original-value",
				},
			},
		}

		// WHEN
		packet.Set("original-key", "new-value")

		// THEN
		assert.Equal(t, []KeyValuePair{
			{
				Key:   "original-key",
				Value: "new-value",
			},
		}, packet.elements)
	})
}

func TestGamespyPacket_Add(t *testing.T) {
	t.Run("adds duplicate key", func(t *testing.T) {
		// GIVEN
		packet := &Packet{
			elements: []KeyValuePair{
				{
					Key:   "key",
					Value: "a-value",
				},
			},
		}

		// WHEN
		packet.Add("key", "b-value")

		// THEN
		assert.Equal(t, []KeyValuePair{
			{
				Key:   "key",
				Value: "a-value",
			},
			{
				Key:   "key",
				Value: "b-value",
			},
		}, packet.elements)
	})
}

func TestGamespyPacket_Lookup(t *testing.T) {
	const key = "key"
	const value = "value"
	packet := &Packet{
		elements: []KeyValuePair{
			{
				Key:   key,
				Value: value,
			},
		},
	}

	t.Run("returns value, true for existing key", func(t *testing.T) {
		// WHEN
		actual, ok := packet.Lookup(key)

		// THEN
		assert.Equal(t, value, actual)
		assert.True(t, ok)
	})

	t.Run("returns empty string, false for missing key", func(t *testing.T) {
		// WHEN
		actual, ok := packet.Lookup("missing")

		// THEN
		assert.Empty(t, actual)
		assert.False(t, ok)
	})
}

func TestGamespyPacket_Get(t *testing.T) {
	const key = "key"
	const value = "value"
	packet := &Packet{
		elements: []KeyValuePair{
			{
				Key:   key,
				Value: value,
			},
		},
	}

	t.Run("returns value for existing key", func(t *testing.T) {
		// WHEN
		actual := packet.Get(key)

		// THEN
		assert.Equal(t, value, actual)
	})

	t.Run("returns empty string for missing key", func(t *testing.T) {
		// WHEN
		actual := packet.Get("missing")

		// THEN
		assert.Empty(t, actual)
	})
}

func TestGamespyPacket_GetAll(t *testing.T) {
	const key = "key"
	packet := &Packet{
		elements: []KeyValuePair{
			{
				Key:   key,
				Value: "a-value",
			},
			{
				Key:   key,
				Value: "b-value",
			},
			{
				Key:   "other-key",
				Value: "other-value",
			},
		},
	}

	t.Run("returns multiple values for duplicate key", func(t *testing.T) {
		// WHEN
		actual := packet.GetAll(key)

		// THEN
		assert.Equal(t, []string{"a-value", "b-value"}, actual)
	})

	t.Run("returns single value for unique key", func(t *testing.T) {
		// WHEN
		actual := packet.GetAll("other-key")

		// THEN
		assert.Equal(t, []string{"other-value"}, actual)
	})

	t.Run("returns nil for missing key", func(t *testing.T) {
		// WHEN
		actual := packet.GetAll("missing")

		// THEN
		assert.Nil(t, actual)
	})
}

func TestGamespyPacket_Map(t *testing.T) {
	type test struct {
		name        string
		packet      *Packet
		expectedMap map[string]string
	}

	tests := []test{
		{
			name: "converts packet to map",
			packet: &Packet{
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
			packet: &Packet{
				elements: []KeyValuePair{},
			},
			expectedMap: map[string]string{},
		},
		{
			name:        "converts element nil slice packet to map",
			packet:      &Packet{},
			expectedMap: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WHEN
			//goland:noinspection GoDeprecation
			actual := tt.packet.Map()

			// THEN
			assert.Equal(t, tt.expectedMap, actual)
		})
	}
}

func TestGamespyPacket_Bytes(t *testing.T) {
	type test struct {
		name          string
		packet        *Packet
		expectedBytes []byte
	}

	tests := []test{
		{
			name: "converts packet to string",
			packet: &Packet{
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
			packet: &Packet{
				elements: []KeyValuePair{},
			},
			expectedBytes: []byte("\\\\final\\"),
		},
		{
			name:          "converts element nil slice packet to string",
			packet:        &Packet{},
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
