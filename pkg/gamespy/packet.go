package gamespy

import (
	"bytes"
	"errors"
	"fmt"
	"iter"
	"reflect"
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

// All returns an iterator over the packet's key value pairs
func (p *Packet) All() iter.Seq[KeyValuePair] {
	return func(yield func(KeyValuePair) bool) {
		for _, element := range p.elements {
			if !yield(element) {
				return
			}
		}
	}
}

func (p *Packet) Bind(target any) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer to a struct or struct-slice")
	}

	if v.Elem().Kind() == reflect.Struct {
		return p.bindStruct(v)
	} else if v.Elem().Kind() == reflect.Slice && v.Elem().Type().Elem().Kind() == reflect.Struct {
		return p.bindSlice(v)
	}

	return fmt.Errorf("target must be a non-nil pointer to a struct or struct-slice")
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

func (p *Packet) bindStruct(v reflect.Value) error {
	v = v.Elem()
	t := v.Type()
	for i := range t.NumField() {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		key := t.Field(i).Tag.Get("gamespy")
		if key == "" {
			continue
		}

		// Get all values and use last one to mimic default JSON library behavior
		values := p.GetAll(key)
		if len(values) == 0 {
			continue
		}
		value := values[len(values)-1]

		if err := setValue(t, i, field, value); err != nil {
			return err
		}
	}

	return nil
}

func (p *Packet) bindSlice(v reflect.Value) error {
	v = v.Elem()
	t := v.Type()
	et := t.Elem()

	// Map tags to element type field indexes
	indexes := make(map[string]int, et.NumField())
	for i := range et.NumField() {
		key := et.Field(i).Tag.Get("gamespy")
		if key == "" {
			continue
		}

		indexes[key] = i
	}

	current := reflect.New(et).Elem()
	keys := make(map[string]struct{})
	for _, element := range p.elements {
		// Start building a new result when we reach a key we saw before
		_, seen := keys[element.Key]
		if seen {
			v.Set(reflect.Append(v, current))
			current = reflect.New(et).Elem()
			keys = make(map[string]struct{}, len(keys))
		}
		keys[element.Key] = struct{}{}

		i, ok := indexes[element.Key]
		if !ok {
			continue
		}

		field := current.Field(i)
		if !field.CanSet() {
			continue
		}

		if err := setValue(et, i, field, element.Value); err != nil {
			return err
		}
	}

	// Add current result if we found (some) keys, but never found a 2nd result
	// (we only "flush" current to results on the n+1st result)
	if len(keys) != 0 {
		v.Set(reflect.Append(v, current))
	}

	return nil
}

func setValue(t reflect.Type, i int, v reflect.Value, value string) error {
	switch v.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("%s.%s: %w", t.Name(), t.Field(i).Name, err)
		}
		v.SetInt(n)
	case reflect.String:
		v.SetString(value)
	case reflect.Pointer:
		v.Set(reflect.New(v.Type().Elem()))
		return setValue(t, i, v.Elem(), value)
	default:
		return fmt.Errorf("%s.%s: unsupported field type: %s", t.Name(), t.Field(i).Name, v.Type())
	}

	return nil
}
