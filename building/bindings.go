package building

import (
	"fmt"
	"sklair/htmlUtilities"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

type propKind uint8

const (
	stringProp propKind = iota
	booleanProp
)

type componentProps struct {
	values      map[string]string
	kinds       map[string]propKind
	signature   string
	description string
}

func bind(nodes []*html.Node, attributes []html.Attribute) (*componentProps, error) {
	kinds := make(map[string]propKind)
	for _, node := range nodes {
		if err := inspectBindings(node, kinds); err != nil {
			return nil, err
		}
	}

	values := make(map[string]string, len(attributes))
	for _, attribute := range attributes {
		name := strings.ToLower(attribute.Key)
		if _, exists := values[name]; exists {
			return nil, fmt.Errorf("prop %q is supplied more than once", name)
		}
		if _, exists := kinds[name]; !exists {
			return nil, fmt.Errorf("component does not accept prop %q", name)
		}
		values[name] = attribute.Val
	}

	props := &componentProps{values: values, kinds: kinds}
	if err := props.normalise(); err != nil {
		return nil, err
	}
	for _, node := range nodes {
		if err := bindNode(node, props); err != nil {
			return nil, err
		}
	}
	return props, nil
}

func inspectBindings(node *html.Node, kinds map[string]propKind) error {
	if node.Type == html.ElementNode {
		for _, attribute := range node.Attr {
			kind, bound, err := bindingKind(attribute.Key)
			if err != nil {
				return err
			}
			if !bound {
				continue
			}

			name, err := propName(attribute.Val)
			if err != nil {
				return fmt.Errorf("invalid %s binding : %s", attribute.Key, err.Error())
			}
			if previous, exists := kinds[name]; exists && previous != kind {
				return fmt.Errorf("prop %q is used as both a string and a boolean", name)
			}
			kinds[name] = kind
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if err := inspectBindings(child, kinds); err != nil {
			return err
		}
	}
	return nil
}

func bindingKind(name string) (propKind, bool, error) {
	switch {
	case name == "sklair-if", strings.HasPrefix(name, "sklair-class-"):
		if name == "sklair-class-" {
			return 0, false, fmt.Errorf("sklair-class binding is missing its class name")
		}
		return booleanProp, true, nil
	case name == "sklair-text", strings.HasPrefix(name, "sklair-attr-"):
		if name == "sklair-attr-" {
			return 0, false, fmt.Errorf("sklair-attr binding is missing its attribute name")
		}
		return stringProp, true, nil
	case strings.HasPrefix(name, "sklair-"):
		return 0, false, fmt.Errorf("unknown compile-time binding %q", name)
	default:
		return 0, false, nil
	}
}

func propName(value string) (string, error) {
	name := strings.ToLower(strings.TrimSpace(value))
	if name == "" {
		return "", fmt.Errorf("prop name cannot be empty")
	}
	if strings.ContainsAny(name, " \t\r\n") {
		return "", fmt.Errorf("prop name %q cannot contain whitespace", value)
	}
	return name, nil
}

func (p *componentProps) normalise() error {
	names := make([]string, 0, len(p.kinds))
	for name := range p.kinds {
		names = append(names, name)
	}
	sort.Strings(names)

	var signature strings.Builder
	descriptions := make([]string, 0, len(p.values))
	for _, name := range names {
		value, exists := p.values[name]
		if p.kinds[name] == booleanProp {
			boolean, err := booleanValue(value, exists)
			if err != nil {
				return fmt.Errorf("invalid value for boolean prop %q : %s", name, err.Error())
			}
			value = fmt.Sprintf("%t", boolean)
			exists = true
		}

		signature.WriteString(fmt.Sprintf("%d:%s:%t:%d:%s;", len(name), name, exists, len(value), value))
		if _, supplied := p.values[name]; supplied {
			descriptions = append(descriptions, fmt.Sprintf("%s=%q", name, value))
		}
	}

	p.signature = signature.String()
	p.description = "{}"
	if len(descriptions) > 0 {
		p.description = "{" + strings.Join(descriptions, ", ") + "}"
	}
	return nil
}

func booleanValue(value string, exists bool) (bool, error) {
	if !exists {
		return false, nil
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("expected an empty value, true, or false; got %q", value)
	}
}

func bindNode(node *html.Node, props *componentProps) error {
	if node.Type == html.ElementNode {
		ifName := ""
		textName := ""
		attributeBindings := make(map[string]string)
		classBindings := make(map[string]string)
		attributes := make([]html.Attribute, 0, len(node.Attr))

		for _, attribute := range node.Attr {
			name, err := propName(attribute.Val)
			switch {
			case attribute.Key == "sklair-if":
				if err != nil {
					return err
				}
				ifName = name
			case attribute.Key == "sklair-text":
				if err != nil {
					return err
				}
				textName = name
			case strings.HasPrefix(attribute.Key, "sklair-attr-"):
				if err != nil {
					return err
				}
				attributeBindings[strings.TrimPrefix(attribute.Key, "sklair-attr-")] = name
			case strings.HasPrefix(attribute.Key, "sklair-class-"):
				if err != nil {
					return err
				}
				classBindings[strings.TrimPrefix(attribute.Key, "sklair-class-")] = name
			default:
				attributes = append(attributes, attribute)
			}
		}
		node.Attr = attributes

		if ifName != "" {
			value, exists := props.values[ifName]
			keep, err := booleanValue(value, exists)
			if err != nil {
				return fmt.Errorf("invalid value for boolean prop %q : %s", ifName, err.Error())
			}
			if !keep {
				node.Parent.RemoveChild(node)
				return nil
			}
		}

		attributeNames := make([]string, 0, len(attributeBindings))
		for name := range attributeBindings {
			attributeNames = append(attributeNames, name)
		}
		sort.Strings(attributeNames)
		for _, name := range attributeNames {
			value, exists := props.values[attributeBindings[name]]
			if !exists {
				return fmt.Errorf("required prop %q was not supplied", attributeBindings[name])
			}
			htmlUtilities.SetAttribute(node, name, value)
		}

		classNames := make([]string, 0, len(classBindings))
		for name := range classBindings {
			classNames = append(classNames, name)
		}
		sort.Strings(classNames)
		for _, name := range classNames {
			value, exists := props.values[classBindings[name]]
			enabled, err := booleanValue(value, exists)
			if err != nil {
				return fmt.Errorf("invalid value for boolean prop %q : %s", classBindings[name], err.Error())
			}
			htmlUtilities.ToggleClass(node, name, enabled)
		}

		if textName != "" {
			value, exists := props.values[textName]
			if !exists {
				return fmt.Errorf("required prop %q was not supplied", textName)
			}
			for child := node.FirstChild; child != nil; {
				next := child.NextSibling
				node.RemoveChild(child)
				child = next
			}
			node.AppendChild(&html.Node{Type: html.TextNode, Data: value})
			return nil
		}
	}

	for child := node.FirstChild; child != nil; {
		next := child.NextSibling
		if err := bindNode(child, props); err != nil {
			return err
		}
		child = next
	}
	return nil
}
