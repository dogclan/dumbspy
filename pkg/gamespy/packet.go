package gamespy

import (
	"bytes"
	"errors"
	"strconv"
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
	elements []KeyValuePair // Store elements in list to maintain order
}

func NewPacket(elements ...KeyValuePair) *Packet {
	return &Packet{
		elements: elements,
	}
}

// Deprecated: Use NewPacketFromBytes instead.
//
//goland:noinspection GoUnusedExportedFunction
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
		packet.Add(string(elements[i]), string(elements[i+1]))
	}
	return packet, nil
}

// Set Adds a new KeyValuePair to the packet. If key exists, the existing KeyValuePair is updated instead.
// Equivalent of calling Packet.Remove and Packet.Add
func (p *Packet) Set(key string, value string) {
	p.Remove(key)
	p.Add(key, value)
}

func (p *Packet) SetInt(key string, value int) {
	p.Set(key, strconv.Itoa(value))
}

// Add Adds a new KeyValuePair to the packet, regardless of whether key already exists in packet.
func (p *Packet) Add(key string, value string) {
	p.elements = append(p.elements, KeyValuePair{Key: key, Value: value})
}

func (p *Packet) AddInt(key string, value int) {
	p.Add(key, strconv.Itoa(value))
}

// Lookup Checks if key exists in packet and returns the first value with a matching key.
func (p *Packet) Lookup(key string) (string, bool) {
	for _, element := range p.elements {
		if element.Key == key {
			return element.Value, true
		}
	}

	return "", false
}

// Get Retrieves the first value with a matching key.
func (p *Packet) Get(key string) string {
	value, _ := p.Lookup(key)
	return value
}

func (p *Packet) GetInt(key string) (int, error) {
	value := p.Get(key)
	if value == "" {
		return 0, nil
	}

	return strconv.Atoi(value)
}

// GetAll Retrieves all values with a matching key. Returns nil if key does not exist in packet.
func (p *Packet) GetAll(key string) []string {
	values := make([]string, 0, len(p.elements))
	for _, element := range p.elements {
		if element.Key == key {
			values = append(values, element.Value)
		}
	}

	if len(values) > 0 {
		return values
	}

	return nil
}

// Remove Removes all KeyValuePair-s which match key.
func (p *Packet) Remove(key string) {
	keepers := make([]KeyValuePair, 0, len(p.elements))
	for _, element := range p.elements {
		if element.Key != key {
			keepers = append(keepers, element)
		}
	}
	p.elements = keepers
}

// Do Calls function f for every KeyValuePair in the Packet.
// The behavior of Do is undefined if f changes *p.
func (p *Packet) Do(f func(element KeyValuePair)) {
	for _, element := range p.elements {
		f(element)
	}
}

// Map
//
// Deprecated: Packet may contain duplicate keys, which will not be reflected in the map.
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
