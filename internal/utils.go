package internal

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/npat-efault/crc16"
)

var (
	conf = &crc16.Conf{
		Poly:   0x8005,
		BitRev: true,
		IniVal: 0x0,
		FinVal: 0x0,
		BigEnd: true,
	}
	nextPlayerID = 500000000
	players      = map[string]int{}
)

func ComputeChecksum(p string) string {
	return strconv.Itoa(int(crc16.Checksum(conf, []byte(p))))
}

func GetPlayerID(nick, passwordHash string) int {
	key := fmt.Sprintf("%s:%s", nick, passwordHash)
	playerID, ok := players[key]
	if !ok {
		playerID = nextPlayerID
		nextPlayerID += 1
		players[key] = playerID
	}

	return playerID
}

func GenerateProof(nick, passwordHash, c1, c2 string) string {
	var b strings.Builder
	b.WriteString(strings.Repeat(" ", 48))
	b.WriteString(nick)
	b.WriteString(c1)
	b.WriteString(c2)
	b.WriteString(passwordHash)

	digest := md5.New()
	digest.Write([]byte(b.String()))
	return fmt.Sprintf("%x", digest.Sum(nil))
}

func RandString(n int) string {
	data := make([]byte, n)
	for i := 0; i < n; i++ {
		data[i] = byte(65 + rand.Intn(25))
	}
	return string(data)
}

func Pointer[T any](p T) *T {
	return &p
}
