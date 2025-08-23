package argoparser

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type fieldsIndex struct {
	fieldsByLongName   map[string]*indexEntry
	fieldsByShortName  map[string]*indexEntry
	fieldsByIndex      map[int]*indexEntry
	positionalsDefault *indexEntry
	requiredFields     []*indexEntry
}

type fieldMeta struct {
	shortName    string
	longName     string
	isPositional bool
	isRequired   bool
}

func getFieldMeta(field reflect.StructField) (fieldMeta, error) {
	meta := fieldMeta{}

	argTag, ok := field.Tag.Lookup("arg")
	if !ok {
		argTag = ""
	}

	argTagParts := strings.Split(argTag, ",")

	for _, part := range argTagParts {
		tag := strings.TrimSpace(part)
		if tag == "positional" {
			meta.isPositional = true
		} else if tag == "required" {
			meta.isRequired = true
		} else if strings.HasPrefix(tag, "--") {
			meta.longName = tag
		} else if strings.HasPrefix(tag, "-") {
			meta.shortName = tag
		} else {
			return fieldMeta{}, fmt.Errorf("invalid arg tag: %s", tag)
		}
	}

	if meta.isPositional && (meta.shortName != "" || meta.longName != "") {
		return fieldMeta{}, fmt.Errorf("positional field cannot have short or long name")
	}

	return meta, nil
}

type indexEntry struct {
	v reflect.Value
	t reflect.Type
	m fieldMeta

	presented bool
}

func preinit(entry *indexEntry) {
	if entry.v.Kind() == reflect.Slice {
		entry.v.Set(reflect.MakeSlice(entry.v.Type(), 0, 0))
	}
}

func buildIndex(v any) (fieldsIndex, error) {
	index := fieldsIndex{
		fieldsByLongName:  make(map[string]*indexEntry),
		fieldsByShortName: make(map[string]*indexEntry),
		fieldsByIndex:     make(map[int]*indexEntry),
	}

	if err := validateInput(v); err != nil {
		return index, err
	}

	rt := reflect.TypeOf(v).Elem()
	rv := reflect.ValueOf(v).Elem()
	numFields := rt.NumField()
	positionalIndex := 0

	for i := 0; i < numFields; i++ {
		field := rt.Field(i)
		fv := rv.FieldByIndex(field.Index)
		fm, err := getFieldMeta(field)
		if err != nil {
			return index, err
		}

		entry := &indexEntry{
			v: fv,
			t: field.Type,
			m: fm,
		}

		preinit(entry)

		if fm.longName != "" {
			if _, ok := index.fieldsByLongName[fm.longName]; ok {
				return index, fmt.Errorf("multiple fields for one key: %s", fm.longName)
			}
			index.fieldsByLongName[fm.longName] = entry
		}
		if fm.shortName != "" {
			if _, ok := index.fieldsByShortName[fm.shortName]; ok {
				return index, fmt.Errorf("multiple fields for one key: %s", fm.shortName)
			}
			index.fieldsByShortName[fm.shortName] = entry
		}
		if fm.isPositional {
			if fv.Kind() == reflect.Slice {
				if index.positionalsDefault != nil {
					return index, fmt.Errorf("multiple positional default fields are not supported")
				}
				index.positionalsDefault = entry
			} else {
				index.fieldsByIndex[positionalIndex] = entry
				positionalIndex++
			}
		}
		if fm.isRequired {
			index.requiredFields = append(index.requiredFields, entry)
		}
	}

	return index, nil
}

func validateInput(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.New("input must be a pointer")
	}
	if rv.IsNil() {
		return errors.New("input must not be nil")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return errors.New("input must be a pointer to a struct")
	}

	return nil
}
