package main

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
)

// parseQuery parses query values into a struct. It assumes that the
// struct field names are the same as the names of the query values,
// unless the fields have a "query" tag attached to them, in which
// case the value of that tag is used instead.
func parseQuery(query url.Values, into interface{}) error {
	v := reflect.Indirect(reflect.ValueOf(into))
	if (v.Kind() != reflect.Struct) || !v.CanAddr() {
		return errors.New("invalid into value")
	}

	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		f := v.Type().Field(i)

		name := f.Tag.Get("query")
		if name == "" {
			name = f.Name
		}

		qv := query.Get(name)
		if qv == "" {
			continue
		}

		switch fv.Kind() {
		case reflect.Bool:
			fv.SetBool(qv == "true")

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			p, err := strconv.ParseInt(qv, 10, 0)
			if err != nil {
				return fmt.Errorf("parse %q: %w", name, err)
			}
			fv.SetInt(p)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			p, err := strconv.ParseUint(qv, 10, 0)
			if err != nil {
				return fmt.Errorf("parse %q: %w", name, err)
			}
			fv.SetUint(p)

		case reflect.String:
			fv.SetString(qv)

		default:
			return fmt.Errorf("unsupported kind: %q", fv.Kind())
		}
	}

	return nil
}
