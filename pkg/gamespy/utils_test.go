//go:build unit

package gamespy

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeCRC16(t *testing.T) {
	// GIVEN
	input := "some-unique-nick"
	expected := uint16(32318)

	// WHEN
	actual := ComputeCRC16(input)

	// THEN
	assert.Equal(t, expected, actual)
}

func TestComputeMD5(t *testing.T) {
	// GIVEN
	input := "some-proof"
	expected := "131def0e93e67e3e62b39d74d6316511"

	// WHEN
	actual := ComputeMD5(input)

	// THEN
	assert.Equal(t, expected, actual)
}

func TestGenerateProof(t *testing.T) {
	// GIVEN
	nick := "some-nick"
	hash := "131def0e93e67e3e62b39d74d6316511"
	c1 := "4Jp6A4kK02"
	c2 := "YJk5UFExKBwn0PEpOpinWHsRCDcfejyJ"
	expected := "4b1ec6377ec7f3c99716df13680638e2"

	// WHEN
	actual := GenerateProof(nick, hash, c1, c2)

	// THEN
	assert.Equal(t, expected, actual)
}

func TestRandString(t *testing.T) {
	type test struct {
		name   string
		length int
	}

	tests := []test{
		{
			name:   "generates 10 character string",
			length: 10,
		},
		{
			name:   "generates 50 character string",
			length: 50,
		},
		{
			name:   "generates zero-length string",
			length: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WHEN
			s := RandString(tt.length)

			// THEN
			assert.Len(t, s, tt.length)
			for _, r := range s {
				// Every character must be a digit or an upper/lower case letter
				assert.True(t, unicode.IsDigit(r) || unicode.IsUpper(r) || unicode.IsLower(r))
			}
		})
	}
}

func TestEncodePassword(t *testing.T) {
	type test struct {
		name    string
		pass    string
		passenc string
	}

	tests := []test{
		{
			name:    "encodes password",
			pass:    "didykilikaj2",
			passenc: "cpuQi0CygMUSbhTY",
		},
		{
			name:    "encodes password with padding",
			pass:    "p@ssw0rd",
			passenc: "ZrKHgVzrnsg_",
		},
		{
			name:    "encodes zero-length password",
			pass:    "",
			passenc: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WHEN
			passenc := EncodePassword(tt.pass)

			// THEN
			assert.Equal(t, tt.passenc, passenc)
		})
	}
}

func TestDecodePassword(t *testing.T) {
	type test struct {
		name    string
		passenc string
		pass    string
		wantErr bool
	}

	tests := []test{
		{
			name:    "decodes password",
			passenc: "cpuQi0CygMUSbhTY",
			pass:    "didykilikaj2",
		},
		{
			name:    "decodes password with padding",
			passenc: "ZrKHgVzrnsg_",
			pass:    "p@ssw0rd",
		},
		{
			name:    "decodes zero-length password",
			passenc: "",
			pass:    "",
		},
		{
			name:    "fails for non-base64 string",
			passenc: "123456789",
			wantErr: true,
		},
		{
			name:    "fails for standard base64 string",
			passenc: "ZrKHgVzrnsg=",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WHEN
			pass, err := DecodePassword(tt.passenc)

			// THEN
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.pass, pass)
			}
		})
	}
}
