package packet

import (
	"fmt"
	"strings"
)

type KeyValuePair struct {
	Key, Value string
}

func (p *KeyValuePair) String() string {
	return fmt.Sprintf("%s\\%s", p.Key, p.Value)
}

type GamespyPacket struct {
	elements []KeyValuePair // Store elements in list to maintain order
}

func NewGamespyPacket(elements ...KeyValuePair) *GamespyPacket {
	return &GamespyPacket{
		elements: elements,
	}
}

func FromString(raw string) (*GamespyPacket, error) {
	if !strings.HasPrefix(raw, "\\") || !strings.HasSuffix(raw, "\\final\\") {
		return nil, fmt.Errorf("gamespy packet string is malformed")
	}
	packet := NewGamespyPacket()
	elements := strings.Split(raw[1:len(raw)-7], "\\")
	for i := 0; i < len(elements); i += 2 {
		packet.Write(elements[i], elements[i+1])
	}
	return packet, nil
}

func (p *GamespyPacket) Write(key string, value string) {
	p.elements = append(p.elements, KeyValuePair{key, value})
}

func (p *GamespyPacket) Map() map[string]string {
	elements := map[string]string{}
	for _, element := range p.elements {
		elements[element.Key] = element.Value
	}
	return elements
}

func (p *GamespyPacket) String() string {
	elements := make([]string, len(p.elements))
	for i, element := range p.elements {
		elements[i] = element.String()
	}
	return fmt.Sprintf("\\%s\\final\\", strings.Join(elements, "\\"))
}

func (p *GamespyPacket) Bytes() []byte {
	return []byte(p.String())
}
