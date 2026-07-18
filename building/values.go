package building

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

type sklairValueKind uint8

const (
	nilValue sklairValueKind = iota
	booleanValue
	stringValue
	numberValue
	arrayValue
	objectValue
)

type sklairValue struct {
	kind    sklairValueKind
	boolean bool
	string  string
	number  float64
	array   []sklairValue
	object  map[string]sklairValue
}

func ownValue(value any) (sklairValue, error) {
	switch value := value.(type) {
	case nil:
		return sklairValue{kind: nilValue}, nil
	case bool:
		return sklairValue{kind: booleanValue, boolean: value}, nil
	case string:
		return sklairValue{kind: stringValue, string: value}, nil
	case int:
		return finiteNumber(float64(value))
	case int8:
		return finiteNumber(float64(value))
	case int16:
		return finiteNumber(float64(value))
	case int32:
		return finiteNumber(float64(value))
	case int64:
		return finiteNumber(float64(value))
	case uint:
		return finiteNumber(float64(value))
	case uint8:
		return finiteNumber(float64(value))
	case uint16:
		return finiteNumber(float64(value))
	case uint32:
		return finiteNumber(float64(value))
	case uint64:
		return finiteNumber(float64(value))
	case float32:
		return finiteNumber(float64(value))
	case float64:
		return finiteNumber(value)
	case []any:
		array := make([]sklairValue, len(value))
		for index, child := range value {
			owned, err := ownValue(child)
			if err != nil {
				return sklairValue{}, fmt.Errorf("array item %d : %s", index+1, err.Error())
			}
			array[index] = owned
		}
		return sklairValue{kind: arrayValue, array: array}, nil
	case map[string]any:
		object := make(map[string]sklairValue, len(value))
		for key, child := range value {
			owned, err := ownValue(child)
			if err != nil {
				return sklairValue{}, fmt.Errorf("object field %q : %s", key, err.Error())
			}
			object[key] = owned
		}
		return sklairValue{kind: objectValue, object: object}, nil
	default:
		return sklairValue{}, fmt.Errorf("unsupported Sklair value %T", value)
	}
}

func finiteNumber(value float64) (sklairValue, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return sklairValue{}, fmt.Errorf("number must be finite")
	}
	return sklairValue{kind: numberValue, number: value}, nil
}

func ownValues(values map[string]any) (map[string]sklairValue, error) {
	owned := make(map[string]sklairValue, len(values))
	for suppliedName, value := range values {
		name, err := propName(suppliedName)
		if err != nil {
			return nil, err
		}
		if _, exists := owned[name]; exists {
			return nil, fmt.Errorf("prop %q is supplied more than once", name)
		}
		ownedValue, err := ownValue(value)
		if err != nil {
			return nil, fmt.Errorf("invalid prop %q : %s", name, err.Error())
		}
		owned[name] = ownedValue
	}
	return owned, nil
}

func (v sklairValue) scalarString() (string, bool, error) {
	switch v.kind {
	case nilValue:
		return "", false, nil
	case booleanValue:
		return strconv.FormatBool(v.boolean), true, nil
	case stringValue:
		return v.string, true, nil
	case numberValue:
		return strconv.FormatFloat(v.number, 'g', -1, 64), true, nil
	case arrayValue:
		return "", false, fmt.Errorf("arrays cannot be used by scalar bindings")
	case objectValue:
		return "", false, fmt.Errorf("objects cannot be used by scalar bindings")
	default:
		return "", false, fmt.Errorf("unknown Sklair value kind")
	}
}

func (v sklairValue) booleanBinding() (bool, error) {
	switch v.kind {
	case nilValue:
		return false, nil
	case booleanValue:
		return v.boolean, nil
	case stringValue:
		switch strings.ToLower(strings.TrimSpace(v.string)) {
		case "", "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return false, fmt.Errorf("expected an empty value, true, or false; got %q", v.string)
		}
	case numberValue:
		return false, fmt.Errorf("numbers cannot be used by boolean bindings")
	case arrayValue:
		return false, fmt.Errorf("arrays cannot be used by boolean bindings")
	case objectValue:
		return false, fmt.Errorf("objects cannot be used by boolean bindings")
	default:
		return false, fmt.Errorf("unknown Sklair value kind")
	}
}

func (v sklairValue) luaValue() any {
	switch v.kind {
	case nilValue:
		return nil
	case booleanValue:
		return v.boolean
	case stringValue:
		return v.string
	case numberValue:
		return v.number
	case arrayValue:
		array := make([]any, len(v.array))
		for index, child := range v.array {
			array[index] = child.luaValue()
		}
		return array
	case objectValue:
		object := make(map[string]any, len(v.object))
		for key, child := range v.object {
			object[key] = child.luaValue()
		}
		return object
	default:
		return nil
	}
}

func (v sklairValue) appendSignature(signature *strings.Builder) {
	signature.WriteByte(byte('0' + v.kind))
	switch v.kind {
	case booleanValue:
		signature.WriteString(strconv.FormatBool(v.boolean))
	case stringValue:
		signature.WriteString(strconv.Itoa(len(v.string)))
		signature.WriteByte(':')
		signature.WriteString(v.string)
	case numberValue:
		signature.WriteString(strconv.FormatFloat(v.number, 'g', -1, 64))
	case arrayValue:
		for _, child := range v.array {
			child.appendSignature(signature)
			signature.WriteByte(';')
		}
	case objectValue:
		keys := make([]string, 0, len(v.object))
		for key := range v.object {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			signature.WriteString(strconv.Itoa(len(key)))
			signature.WriteByte(':')
			signature.WriteString(key)
			v.object[key].appendSignature(signature)
			signature.WriteByte(';')
		}
	}
}
