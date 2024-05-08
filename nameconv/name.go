package nameconv

import (
	"strings"
	"unicode"
)

// Name 对象名, 需要输入驼峰规则的名
type Name string

// ExportedCamel 生成公开(大写开头)的名称
func (name Name) ExportedCamel() string {
	n := len(name)
	if n == 0 {
		return string(name)
	}

	w := &strings.Builder{}

	if name[0] >= 'a' && name[0] <= 'z' {
		w.WriteByte(name[0] ^ 0x20)
	} else {
		w.WriteByte(name[0])
	}
	return name.camel(w, 1)
}

// UnexportedCamel 生成私有(小写开头)的名称
func (name Name) UnexportedCamel() string {
	n := len(name)
	if n == 0 {
		return string(name)
	}

	w := &strings.Builder{}

	if name[0] >= 'A' && name[0] <= 'Z' {
		w.WriteByte(name[0] ^ 0x20)
	} else {
		w.WriteByte(name[0])
	}
	return name.camel(w, 1)
}

func (name Name) camel(w *strings.Builder, startIdx int64) string {
	n := int64(len(name))

	upper := false
	for i := startIdx; i < n; i++ {
		if name[i] == '_' {
			upper = true
			continue
		}
		if upper {
			if name[i] >= 'a' && name[i] <= 'z' {
				upper = false
				w.WriteByte(name[i] ^ 0x20)
			} else {
				w.WriteByte(name[i])
			}
		} else {
			w.WriteByte(name[i])
		}
	}
	if upper { // Write the last '_'
		w.WriteByte('_')
	}
	return w.String()
}

// Snake 生成蛇式的名称
func (name Name) Snake() string {
	var output []rune
	var segment []rune
	prevStrIsUpper := false
	prevStrIsNumber := false
	for i, r := range name {

		// not treat number as separate segment
		if unicode.IsUpper(r) &&
			(!prevStrIsUpper || i < len(name)-1 && unicode.IsLower(rune(name[i+1])) ||
				prevStrIsNumber && unicode.IsUpper(r)) {
			output = addSegment(output, segment)
			segment = nil
		}
		if unicode.IsUpper(r) {
			prevStrIsUpper = true
		} else if unicode.IsLower(r) {
			prevStrIsUpper = false
		}
		if unicode.IsNumber(r) {
			prevStrIsNumber = true
		} else if !unicode.IsNumber(r) {
			prevStrIsNumber = false
		}
		segment = append(segment, unicode.ToLower(r))
	}
	output = addSegment(output, segment)
	return string(output)
}

func addSegment(inrune, segment []rune) []rune {
	if len(segment) == 0 {
		return inrune
	}
	if len(inrune) != 0 {
		inrune = append(inrune, '_')
	}
	inrune = append(inrune, segment...)
	return inrune
}

func (name Name) String() string {
	return string(name)
}

// ExportedCamel 生成公开(大写开头)的名称
func ExportedCamel(n string) string {
	return Name(n).ExportedCamel()
}

// UnexportedCamel 生成私有(小写开头)的名称
func UnexportedCamel(n string) string {
	return Name(n).UnexportedCamel()
}

// Snake 生成蛇式的名称
func Snake(n string) string {
	return Name(n).Snake()
}
