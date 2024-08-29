package gamespy

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"strings"

	"github.com/npat-efault/crc16"
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
)

func ComputeCRC16(p string) uint16 {
	return crc16.Checksum(conf, []byte(p))
}

func ComputeMD5(p string) string {
	digest := md5.New()
	digest.Write([]byte(p))
	return fmt.Sprintf("%x", digest.Sum(nil))
}

func GenerateProof(nick, hash, c1, c2 string) string {
	var b strings.Builder
	b.WriteString(hash)
	b.WriteString(strings.Repeat(" ", 48))
	b.WriteString(nick)
	b.WriteString(c1)
	b.WriteString(c2)
	b.WriteString(hash)

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
