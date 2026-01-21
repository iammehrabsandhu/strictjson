package strictjson

import (
	"encoding/json"
	"reflect"
)

// Unmarshal and stores the result in the value pointed to by v.
// Unlike encoding/json.Unmarshal, this function enforces case-sensitive matching
// between JSON keys and struct field names. JSON keys must exactly match the
// struct field name or its json tag value.
//
// Case-sensitive validation is applied recursively to -
//   - Nested structs
//   - Slices/arrays containing structs
//   - Maps with struct values
func Unmarshal(data []byte, v any) error {
	d := NewDecoder()
	return d.Unmarshal(data, v)
}

func (d *Decoder) Unmarshal(data []byte, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return newNonPointerError()
	}

	return d.unmarshalValue(data, rv.Elem())
}

func (d *Decoder) unmarshalValue(data []byte, v reflect.Value) error {
	if string(data) == "null" {
		return nil
	}
	t := v.Type()
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if v.CanAddr() && implementsUnmarshaler(v.Addr().Type()) {
		return json.Unmarshal(data, v.Addr().Interface())
	}

	v = allocatePointers(v)

	switch v.Kind() {
	case reflect.Struct:
		return d.unmarshalStruct(data, v)
	case reflect.Slice:
		return d.unmarshalSlice(data, v)
	case reflect.Map:
		return d.unmarshalMap(data, v)
	default:
		return json.Unmarshal(data, v.Addr().Interface())
	}
}

func implementsUnmarshaler(t reflect.Type) bool {
	unmarshalerType := reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
	return t.Implements(unmarshalerType)
}

func (d *Decoder) unmarshalStruct(data []byte, v reflect.Value) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	sf, err := getStructFields(v.Type())
	if err != nil {
		return err
	}

	if d.DisallowUnknownFields {
		for jsonKey := range raw {
			if _, exists := sf.fields[jsonKey]; !exists {
				suggestion := ""
				if d.SuggestClosest {
					suggestion = findSuggestion(jsonKey, sf.allNames)
				}
				return newUnknownFieldError(jsonKey, suggestion)
			}
		}
	}

	for jsonKey, rawValue := range raw {
		fi, exists := sf.fields[jsonKey]
		if !exists {
			continue
		}

		fieldValue := getFieldByIndex(v, fi.fieldIndex)
		if !fieldValue.IsValid() || !fieldValue.CanSet() {
			continue
		}

		if err := d.unmarshalValue(rawValue, fieldValue); err != nil {
			return err
		}
	}

	return nil
}

func (d *Decoder) unmarshalSlice(data []byte, v reflect.Value) error {
	var rawSlice []json.RawMessage
	if err := json.Unmarshal(data, &rawSlice); err != nil {
		return err
	}

	elemType := v.Type().Elem()
	needsValidation := containsStruct(elemType)

	if !needsValidation {
		return json.Unmarshal(data, v.Addr().Interface())
	}

	newSlice := reflect.MakeSlice(v.Type(), len(rawSlice), len(rawSlice))

	for i, rawElem := range rawSlice {
		elem := newSlice.Index(i)
		if err := d.unmarshalValue(rawElem, elem); err != nil {
			return err
		}
	}

	v.Set(newSlice)
	return nil
}

func (d *Decoder) unmarshalMap(data []byte, v reflect.Value) error {
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return err
	}

	valueType := v.Type().Elem()
	needsValidation := containsStruct(valueType)

	if !needsValidation {
		return json.Unmarshal(data, v.Addr().Interface())
	}

	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	keyType := v.Type().Key()

	for key, rawValue := range rawMap {
		keyVal := reflect.ValueOf(key)
		if keyType.Kind() != reflect.String {
			keyVal = keyVal.Convert(keyType)
		}
		elemVal := reflect.New(valueType).Elem()
		if err := d.unmarshalValue(rawValue, elemVal); err != nil {
			return err
		}

		v.SetMapIndex(keyVal, elemVal)
	}

	return nil
}

func containsStruct(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		ptrType := reflect.PointerTo(t)
		if implementsUnmarshaler(ptrType) {
			return false
		}
		return true
	case reflect.Slice, reflect.Array:
		return containsStruct(t.Elem())
	case reflect.Map:
		return containsStruct(t.Elem())
	default:
		return false
	}
}

// getFieldByIndex retrieves a field value by its index path.
// This handles embedded structs by following the index path.
func getFieldByIndex(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				if !v.CanSet() {
					return reflect.Value{}
				}
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return reflect.Value{}
		}
		v = v.Field(i)
	}
	return v
}

func allocatePointers(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}
