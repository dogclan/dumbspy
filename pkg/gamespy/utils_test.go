//go:build unit

package gamespy

import (
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
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
