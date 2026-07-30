package building

import (
	"fmt"
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

func structuredSklairValue(value any) sklairValue {
	switch value := value.(type) {
	case nil:
		return sklairValue{kind: nilValue}
	case bool:
		return sklairValue{kind: booleanValue, boolean: value}
	case string:
		return sklairValue{kind: stringValue, string: value}
	case float64:
		return sklairValue{kind: numberValue, number: value}
	case []any:
		array := make([]sklairValue, len(value))
		for index, child := range value {
			array[index] = structuredSklairValue(child)
		}
		return sklairValue{kind: arrayValue, array: array}
	case map[string]any:
		object := make(map[string]sklairValue, len(value))
		for key, child := range value {
			object[key] = structuredSklairValue(child)
		}
		return sklairValue{kind: objectValue, object: object}
	default:
		panic(fmt.Sprintf("unsupported structured Sklair value %T", value))
	}
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
