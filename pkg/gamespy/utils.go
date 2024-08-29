package gamespy

import (
	"crypto/md5"
	"encoding/base64"
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

	// GameSpy-specific base64.Encoding (uses [ instead of +, ] for / and _ instead of = for padding)
	gamespyEncoding = base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789[]").WithPadding('_')
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

func EncodePassword(pass string) string {
	encoded := []byte(pass)
	gspassenc(encoded)

	return gamespyEncoding.EncodeToString(encoded)
}

func DecodePassword(passenc string) (string, error) {
	decoded, err := gamespyEncoding.DecodeString(passenc)
	if err != nil {
		return "", err
	}

	gspassenc(decoded)

	return string(decoded), nil
}

// gspassenc Encodes/decodes the password in place.
// Adapted from the original implementation by Luigi Auriemma.
//
// See https://aluigi.altervista.org/papers/gspassenc.zip
func gspassenc(pass []byte) {
	num := int32(0x79707367) // "gspy"
	for i := range pass {
		num = gslame(num)
		pass[i] ^= byte(num % 0xff)
	}
}

// gslame Shifts num around. Specifically uses int32 to allow required integer overflows.
func gslame(num int32) int32 {
	c := (num >> 16) & 0xffff
	a := num & 0xffff
	c *= 0x41a7
	a *= 0x41a7
	a += (c & 0x7fff) << 16
	if a < 0 {
		a &= 0x7fffffff
		a++
	}
	a += c >> 15
	if a < 0 {
		a &= 0x7fffffff
		a++
	}
	return a
}
