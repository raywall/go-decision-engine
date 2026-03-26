package tags

import (
	"reflect"
	"strings"
)

// TagMeta represents parsed metadata from a struct tag.
type TagMeta struct {
	Name     string
	Required bool
	Path     string
}

// ParseToMap extracts only schema-defined fields from any input.
func ParseToMap(input any, schema map[string]any) (map[string]any, error) {
	out := make(map[string]any)

	err := extract(reflect.ValueOf(input), out, schema)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// extract recursively walks through structs and maps.
func extract(val reflect.Value, out map[string]any, schema map[string]any) error {
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	switch val.Kind() {

	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			tag := field.Tag.Get("decision")

			fieldVal := val.Field(i)

			if tag != "" {
				meta := parseTag(tag)

				if _, exists := schema[meta.Name]; exists {
					out[meta.Name] = fieldVal.Interface()
				}
			}

			// recursive exploration
			if fieldVal.Kind() == reflect.Struct {
				if err := extract(fieldVal, out, schema); err != nil {
					return err
				}
			}
		}

	case reflect.Map:
		iter := val.MapRange()
		for iter.Next() {
			k := iter.Key().String()
			v := iter.Value()

			if _, ok := schema[k]; ok {
				out[k] = v.Interface()
			}
		}
	}

	return nil
}

// parseTag parses decision tag.
func parseTag(tag string) TagMeta {
	parts := strings.Split(tag, ",")

	meta := TagMeta{
		Name: parts[0],
	}

	for _, p := range parts[1:] {
		if p == "required" {
			meta.Required = true
		}

		if strings.HasPrefix(p, "path=") {
			meta.Path = strings.TrimPrefix(p, "path=")
		}
	}

	return meta
}
