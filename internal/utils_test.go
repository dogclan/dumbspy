package internal

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			expectedPlayerID: 600001095,
		},
		{
			name:        "re-assigns player id to returning player",
			nick:        "some-nick",
			productID:   "some-productID",
			gameName:    "some-gameName",
			namespaceID: "some-namespaceID",
			sdkRevision: "some-sdkRevision",
			players: map[int]string{
				600001095: strings.Join([]string{
					"some-nick",
					"some-productID",
					"some-gameName",
					"some-namespaceID",
					"some-sdkRevision",
				}, ":"),
			},
			expectedPlayerID: 600001095,
		},
		{
			name:        "assigns random player id on identifier collision",
			nick:        "some-nick",
			productID:   "some-productID",
			gameName:    "some-gameName",
			namespaceID: "some-namespaceID",
			sdkRevision: "some-sdkRevision",
			players: map[int]string{
				600001095: "some-other-identifier",
			},
			expectedPlayerID: 600001095,
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
