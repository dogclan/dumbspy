//go:build unit

package internal

import (
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestComputeCRC16Str(t *testing.T) {
	// GIVEN
	input := "some-unique-nick"
	expected := "32318"

	// WHEN
	actual := ComputeCRC16Str(input)

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

func TestGetPlayerID(t *testing.T) {
	type test struct {
		name             string
		nick             string
		productID        string
		gameName         string
		namespaceID      string
		sdkRevision      string
		players          map[int]string
		expectedPlayerID int
		wantRandom       bool
	}

	tests := []test{
		{
			name:             "assigns player id to new player",
			nick:             "some-nick",
			productID:        "some-productID",
			gameName:         "some-gameName",
			namespaceID:      "some-namespaceID",
			sdkRevision:      "some-sdkRevision",
			players:          map[int]string{},
			expectedPlayerID: 500057082,
		},
		{
			name:        "re-assigns player id to returning player",
			nick:        "some-nick",
			productID:   "some-productID",
			gameName:    "some-gameName",
			namespaceID: "some-namespaceID",
			sdkRevision: "some-sdkRevision",
			players: map[int]string{
				500057082: ComputeMD5(strings.Join([]string{
					"some-nick",
					"some-productID",
					"some-gameName",
					"some-namespaceID",
					"some-sdkRevision",
				}, ":")),
			},
			expectedPlayerID: 500057082,
		},
		{
			name:        "assigns random player id on hash collision",
			nick:        "some-nick",
			productID:   "some-productID",
			gameName:    "some-gameName",
			namespaceID: "some-namespaceID",
			sdkRevision: "some-sdkRevision",
			players: map[int]string{
				500057082: "some-other-hash",
			},
			expectedPlayerID: 500057082,
			wantRandom:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN
			players = tt.players

			// WHEN
			firstIterationID := GetPlayerID(tt.nick, tt.productID, tt.gameName, tt.namespaceID, tt.sdkRevision)

			// THEN
			if tt.wantRandom {
				assert.NotEqual(t, tt.expectedPlayerID, firstIterationID)
			} else {
				assert.Equal(t, tt.expectedPlayerID, firstIterationID)
			}

			// WHEN we run GetPlayerID again using the same inputs
			secondIterationID := GetPlayerID(tt.nick, tt.productID, tt.gameName, tt.namespaceID, tt.sdkRevision)

			// THEN we should either receive another random, unique player id or the same id again
			if tt.wantRandom {
				assert.NotEqual(t, tt.expectedPlayerID, secondIterationID)
				assert.NotEqual(t, firstIterationID, secondIterationID)
			} else {
				assert.Equal(t, tt.expectedPlayerID, firstIterationID)
			}
		})
	}
}

func TestGenerateProof(t *testing.T) {
	// GIVEN
	nick := "some-nick"
	response := "131def0e93e67e3e62b39d74d6316511"
	c1 := "4Jp6A4kK02"
	c2 := "YJk5UFExKBwn0PEpOpinWHsRCDcfejyJ"
	expected := "f7914097b628f62ac64cd9400472ce98"

	// WHEN
	actual := GenerateProof(nick, response, c1, c2)

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
