package gamespy

import (
	"bytes"
	"errors"
	"strings"
)

const (
	prefix    = "\\"
	suffix    = "\\final\\"
	separator = "\\"
)

type KeyValuePair struct {
	Key, Value string
}

func (p *KeyValuePair) String() string {
	return p.Key + separator + p.Value
}

type Packet struct {
	// Maps any KeyValuePair.Key in elements to the corresponding index.
	// Used to keep access times consistent when checking if key already exists
	// (rather than checking elements every time).
	keys     map[string]int
	elements []KeyValuePair // Store elements in list to maintain order
}

func NewPacket(elements ...KeyValuePair) *Packet {
	keys := make(map[string]int, len(elements))
	for i, element := range elements {
		keys[element.Key] = i
	}

	return &Packet{
		keys:     keys,
		elements: elements,
	}
}

// Deprecated: Use NewPacketFromBytes instead.
func NewPacketFromString(raw string) (*Packet, error) {
	return NewPacketFromBytes([]byte(raw))
}

func NewPacketFromBytes(b []byte) (*Packet, error) {
	if !bytes.HasPrefix(b, []byte(prefix)) || !bytes.HasSuffix(b, []byte(suffix)) {
		return nil, errors.New("gamespy packet string is malformed")
	}

	elements := bytes.Split(b[1:len(b)-7], []byte(separator))
	if len(elements)%2 != 0 {
		return nil, errors.New("gamespy packet string contains key without corresponding value")
	}

	packet := NewPacket()
	for i := 0; i < len(elements); i += 2 {
		packet.Set(string(elements[i]), string(elements[i+1]))
	}
	return packet, nil
}

// Set Adds a new KeyValuePair to the packet. If key exists, the existing KeyValuePair is updated instead.
func (p *Packet) Set(key string, value string) {
	i, ok := p.keys[key]
	if ok {
		p.elements[i].Value = value
	} else {
		p.elements = append(p.elements, KeyValuePair{Key: key, Value: value})
		p.keys[key] = len(p.elements) - 1
	}
}

func (p *Packet) Lookup(key string) (string, bool) {
	i, ok := p.keys[key]
	if !ok {
		return "", false
	}

	return p.elements[i].Value, true
}

func (p *Packet) Get(key string) string {
	value, _ := p.Lookup(key)
	return value
}

func (p *Packet) Map() map[string]string {
	elements := make(map[string]string, len(p.elements))
	for _, element := range p.elements {
		elements[element.Key] = element.Value
	}
	return elements
}

func (p *Packet) String() string {
	elements := make([]string, len(p.elements))
	for i, element := range p.elements {
		elements[i] = element.String()
	}
	return prefix + strings.Join(elements, separator) + suffix
}

func (p *Packet) Bytes() []byte {
	return []byte(p.String())
}
