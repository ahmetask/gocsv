package gocsv

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type ConvertField func(v string, k reflect.Kind, t reflect.Type) (interface{}, error)

type converter interface {
	convert(data []string) (interface{}, error)
}

type structConverter struct {
	format     interface{}
	slice      string
	cf         ConvertField
	fieldKinds []reflect.Kind
	fieldTypes []reflect.Type
	fieldNames []string
}

func newConverter(format interface{}, cf ConvertField) (converter, error) {
	item := reflect.ValueOf(format)
	if item.Kind() == reflect.Ptr {
		item = item.Elem()
	}
	if item.Kind() != reflect.Struct {
		return nil, errors.New("item format should be struct")
	}
	var fieldKinds []reflect.Kind
	var fieldTypes []reflect.Type
	var fieldNames []string

	for i := 0; i < item.NumField(); i++ {
		kind := item.Field(i).Kind()
		t := item.Field(i).Type()
		fieldKinds = append(fieldKinds, kind)
		fieldTypes = append(fieldTypes, t)
		fieldNames = append(fieldNames, item.Type().Field(i).Name)
	}

	if fieldNames == nil || fieldTypes == nil || fieldKinds == nil {
		return nil, errors.New(fmt.Sprintf("invalid format:%v", format))
	}
	return &structConverter{
		format:     format,
		cf:         cf,
		fieldKinds: fieldKinds,
		fieldTypes: fieldTypes,
		fieldNames: fieldNames,
	}, nil
}

func (c *structConverter) convert(values []string) (interface{}, error) {

	var data = reflect.New(reflect.TypeOf(c.format))
	if data.Kind() == reflect.Ptr {
		for i, v := range values {
			typeOf, err := c.typeof(v, c.fieldKinds[i], c.fieldTypes[i])
			if err == nil && typeOf != nil {
				field := data.Elem().FieldByName(c.fieldNames[i])
				value := reflect.ValueOf(typeOf)

				if value.Kind() == reflect.Ptr && field.Kind() != reflect.Ptr {
					value = value.Elem()
					field.Set(value)
				} else if field.Kind() == reflect.Ptr && value.Kind() != reflect.Ptr {
					pointer := reflect.New(value.Type())
					pointer.Elem().Set(value)
					field.Set(pointer)
				} else {
					field.Set(value)
				}

			} else {
				return nil, err
			}
		}
	} else {
		return nil, errors.New(fmt.Sprintf("invalid field configuration data:%v", strings.Join(values, " ")))
	}

	return data.Interface(), nil
}

func (c *structConverter) typeof(v string, k reflect.Kind, t reflect.Type) (interface{}, error) {
	if v == "" {
		return nil, nil
	}
	switch k {
	case reflect.Uint64:
		r, err := strconv.ParseUint(v, 10, 64)
		return r, err
	case reflect.Uint:
		r, err := strconv.Atoi(v)
		return r, err
	case reflect.Int64:
		r, err := strconv.ParseInt(v, 10, 64)
		return r, err
	case reflect.Int:
		r, err := strconv.Atoi(v)
		return r, err
	case reflect.Float64:
		r, err := strconv.ParseFloat(v, 64)
		return r, err
	case reflect.Bool:
		r, err := strconv.ParseBool(v)
		return r, err
	case reflect.String:
		return v, nil
	case reflect.Slice:
		var a []string
		if reflect.TypeOf(a) == t {
			return strings.Split(v, ","), nil
		}
		return c.cf(v, k, t)
	case reflect.Struct:
		d := reflect.New(t)
		a := d.Interface()
		err := json.Unmarshal([]byte(v), a)
		if err != nil {
			return nil, err
		}
		return a, nil
	default:
		if k == reflect.Ptr {
			return c.typeof(v, t.Elem().Kind(), t.Elem())
		}
		if c.cf == nil {
			return nil, errors.New("custom field converter is nil")
		}
		return c.cf(v, k, t)
	}
}
