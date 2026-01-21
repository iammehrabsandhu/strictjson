package strictjson

import (
	"reflect"
	"strings"
	"sync"
)

type fieldInfo struct {
	jsonName   string
	fieldIndex []int
}

type structFields struct {
	fields   map[string]*fieldInfo
	allNames []string
	conflict string
}

// fieldCache caches struct field mappings by type to avoid repeated reflection.
var fieldCache sync.Map

func getStructFields(t reflect.Type) (*structFields, error) {
	if cached, ok := fieldCache.Load(t); ok {
		sf := cached.(*structFields)
		if sf.conflict != "" {
			return nil, newFieldConflictError(sf.conflict)
		}
		return sf, nil
	}

	sf := buildStructFields(t)
	fieldCache.Store(t, sf)

	if sf.conflict != "" {
		return nil, newFieldConflictError(sf.conflict)
	}
	return sf, nil
}

// buildStructFields extracts field information using BFS to handle shadowing correctly.
func buildStructFields(t reflect.Type) *structFields {
	sf := &structFields{
		fields:   make(map[string]*fieldInfo),
		allNames: make([]string, 0),
	}

	type fieldScan struct {
		typ   reflect.Type
		index []int
	}
	currentLevel := []fieldScan{{typ: t, index: nil}}
	nextLevel := []fieldScan{}

	visitedTypes := map[reflect.Type]bool{}

	for len(currentLevel) > 0 {
		fieldsFoundThisLevel := make(map[string]bool)

		for _, scan := range currentLevel {
			typ := scan.typ

			for typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			if typ.Kind() != reflect.Struct {
				continue
			}

			if visitedTypes[typ] {
				continue
			}
			visitedTypes[typ] = true

			for i := 0; i < typ.NumField(); i++ {
				f := typ.Field(i)
				if f.Anonymous {
					nextIndex := make([]int, len(scan.index)+1)
					copy(nextIndex, scan.index)
					nextIndex[len(scan.index)] = i

					nextLevel = append(nextLevel, fieldScan{
						typ:   f.Type,
						index: nextIndex,
					})
					continue
				}

				if !f.IsExported() {
					continue
				}

				tag := f.Tag.Get("json")
				if tag == "-" {
					continue
				}
				name, _ := parseTag(tag)
				if name == "" {
					name = f.Name
				}

				if fieldsFoundThisLevel[name] {
					delete(sf.fields, name)
					sf.conflict = name
					continue
				}

				if _, exists := sf.fields[name]; exists {
					continue
				}

				indexPath := make([]int, len(scan.index)+1)
				copy(indexPath, scan.index)
				indexPath[len(scan.index)] = i

				sf.fields[name] = &fieldInfo{
					jsonName:   name,
					fieldIndex: indexPath,
				}
				fieldsFoundThisLevel[name] = true
			}
		}

		for name := range fieldsFoundThisLevel {
			if _, ok := sf.fields[name]; ok {
				sf.allNames = append(sf.allNames, name)
			}
		}

		currentLevel = nextLevel
		nextLevel = []fieldScan{}
	}

	return sf
}

func parseTag(tag string) (name, opts string) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tag[idx+1:]
	}
	return tag, ""
}

func findSuggestion(unknown string, knownNames []string) string {
	unknownLower := strings.ToLower(unknown)

	for _, name := range knownNames {
		if strings.ToLower(name) == unknownLower {
			return name
		}
	}

	for _, name := range knownNames {
		if levenshteinDistance(unknown, name) <= 2 {
			return name
		}
	}

	return ""
}

// levenshteinDistance distance between two strings.
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	prev := make([]int, len(s2)+1)
	curr := make([]int, len(s2)+1)

	for j := 0; j <= len(s2); j++ {
		prev[j] = j
	}

	for i := 1; i <= len(s1); i++ {
		curr[0] = i
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			curr[j] = minOfThree(
				prev[j]+1,
				curr[j-1]+1,
				prev[j-1]+cost,
			)
		}
		prev, curr = curr, prev
	}

	return prev[len(s2)]
}

func minOfThree(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
