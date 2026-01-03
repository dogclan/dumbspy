package internal

import (
	"math/rand"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/dogclan/dumbspy/pkg/gamespy"
)

const (
	basePlayerID = 600000000
)

var (
	players      = map[int]string{}
	playersMutex = sync.RWMutex{}
)

func GetPlayerID(nick, productID, gameName, namespaceID, sdkRevision string) int {
	// Join all unique/constant attributes in a login request to get a unique identifier
	identifier := strings.Join([]string{nick, productID, gameName, namespaceID, sdkRevision}, ":")
	playerID := basePlayerID + int(gamespy.ComputeCRC16(identifier))

	// Lock and unlock players map to prevent concurrent access
	playersMutex.Lock()
	defer playersMutex.Unlock()

	existingIdentifier, ok := players[playerID]
	if !ok {
		players[playerID] = identifier
	} else if existingIdentifier != identifier {
		log.Warn().
			Str("existingIdentifier", existingIdentifier).
			Str("identifier", identifier).
			Msg("Player identifier mismatch, assigning random player id")

		// A *different* random player id will be assigned for each collision with the same identifier. A player whose
		// identifier is colliding will thus receive a different player id each time they log in. However, collisions
		// should be rare in reality and the id is unique for the running duration of the dumbspy process.
		for ok {
			playerID = basePlayerID - rand.Intn(10000)
			_, ok = players[playerID]
		}
	}

	return playerID
}

func ToPointer[T any](p T) *T {
	return &p
}
