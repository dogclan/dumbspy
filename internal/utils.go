package internal

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"

	"github.com/npat-efault/crc16"
	"github.com/rs/zerolog/log"
)

const (
	asciiFirstNumber          = 48
	asciiFirstUpperCaseLetter = 65
	asciiFirstLowerCaseLetter = 97
)

var (
	conf = &crc16.Conf{
		Poly:   0x8005,
		BitRev: true,
		IniVal: 0x0,
		FinVal: 0x0,
		BigEnd: true,
	}
	basePlayerID = 500000000
	players      = map[int]string{}
	playersMutex = sync.RWMutex{}
)

func ComputeCRC16Str(p string) string {
	return strconv.Itoa(ComputeCRC16Int(p))
}

func ComputeCRC16Int(p string) int {
	return int(crc16.Checksum(conf, []byte(p)))
}

func ComputeMD5(p string) string {
	digest := md5.New()
	digest.Write([]byte(p))
	return fmt.Sprintf("%x", digest.Sum(nil))
}

func GetPlayerID(nick, productID, gameName, namespaceID, sdkRevision string) int {
	// Compute hash of all unique/constant attributes in a login request
	hash := ComputeMD5(strings.Join([]string{nick, productID, gameName, namespaceID, sdkRevision}, ":"))
	playerID := basePlayerID + ComputeCRC16Int(hash)

	// Lock and unlock players map to prevent concurrent access
	playersMutex.Lock()
	defer playersMutex.Unlock()

	existingHash, ok := players[playerID]
	if !ok {
		players[playerID] = hash
	} else if existingHash != hash {
		log.Warn().
			Str("nick", nick).
			Str("productID", productID).
			Msg("Player hash mismatch, assigning random player id")

		// A *different* random player id will be assigned for each collision with the same hash. A player who's
		// hash is colliding will thus receive a different player id each time they log in. However, collisions should
		// be rare in reality and the id is unique for the running duration of the dumbspy process.
		for ok {
			playerID = basePlayerID - rand.Intn(10000)
			_, ok = players[playerID]
		}
	}

	return playerID
}

func GenerateProof(nick, response, c1, c2 string) string {
	var b strings.Builder
	b.WriteString(strings.Repeat(" ", 48))
	b.WriteString(nick)
	b.WriteString(c1)
	b.WriteString(c2)
	b.WriteString(response)

	return ComputeMD5(b.String())
}

func RandString(n int) string {
	data := make([]byte, n)
	for i := 0; i < n; i++ {
		t := rand.Intn(3)
		var c int
		switch t {
		case 1:
			c = asciiFirstLowerCaseLetter + rand.Intn(25) // lower case letter
		case 2:
			c = asciiFirstNumber + rand.Intn(9) // number
		default:
			c = asciiFirstUpperCaseLetter + rand.Intn(25) // upper case letter
		}
		data[i] = byte(c)
	}
	return string(data)
}

func Pointer[T any](p T) *T {
	return &p
}
