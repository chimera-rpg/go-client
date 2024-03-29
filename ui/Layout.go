package ui

import (
	"fmt"
	"strings"
)

type LayoutEntry struct {
	Element  ElementI
	Tag      string        `yaml:"Tag"`
	Type     string        `yaml:"Type"`
	Class    string        `yaml:"Class"`
	Children []LayoutEntry `yaml:"Children"`
}

func (e LayoutEntry) Generate(styles map[string]string, cfgs map[string]interface{}) LayoutEntry {
	return e.generate(styles, cfgs)
}

func (e LayoutEntry) generate(styles map[string]string, cfgs map[string]interface{}) LayoutEntry {
	parts := strings.SplitN(e.Tag, ".", 2)
	typeString := "Container"
	classString := parts[0]
	if len(parts) > 1 {
		typeString = parts[0]
		classString = parts[1]
	}

	l := LayoutEntry{
		Class: classString,
		Type:  typeString,
	}

	el, _ := l.Construct(styles[l.Class], cfgs[l.Class])
	l.Element = el
	for _, child := range e.Children {
		childEl := child.generate(styles, cfgs)
		l.Element.GetAdoptChannel() <- childEl.Element
		l.Children = append(l.Children, childEl)
	}
	return l
}

func (e LayoutEntry) Find(class string) *LayoutEntry {
	if e.Class == class {
		return &e
	}
	for _, c := range e.Children {
		if r := c.Find(class); r != nil {
			return r
		}
	}
	return nil
}

func (e LayoutEntry) Construct(style string, cfg interface{}) (ElementI, error) {
	switch e.Type {
	case "Image":
		if cfg == nil {
			cfg = ImageElementConfig{}
		}
		c := cfg.(ImageElementConfig)
		if c.Style == "" {
			c.Style = style
		}
		return NewImageElement(c), nil
	case "Text":
		if cfg == nil {
			cfg = TextElementConfig{}
		}
		c := cfg.(TextElementConfig)
		if c.Style == "" {
			c.Style = style
		}
		return NewTextElement(c), nil

	case "Container":
		if cfg == nil {
			cfg = ContainerConfig{}
		}
		c := cfg.(ContainerConfig)
		if c.Style == "" {
			c.Style = style
		}
		return NewContainerElement(c)
	case "Input":
		if cfg == nil {
			cfg = InputElementConfig{}
		}
		c := cfg.(InputElementConfig)
		if c.Style == "" {
			c.Style = style
		}
		return NewInputElement(c), nil

	case "Button":
		if cfg == nil {
			cfg = ButtonElementConfig{}
		}
		c := cfg.(ButtonElementConfig)
		if c.Style == "" {
			c.Style = style
		}
		return NewButtonElement(c), nil
	}
	return nil, fmt.Errorf("no possible element")
}
