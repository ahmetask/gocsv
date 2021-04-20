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
	format interface{}
	slice  string
	cf     ConvertField
}

func newConverter(format interface{}, cf ConvertField) converter {
	return &structConverter{
		format: format,
		cf:     cf,
	}
}

func (c *structConverter) convert(values []string) (interface{}, error) {
	item := reflect.ValueOf(c.format)
	if item.Kind() == reflect.Ptr {
		item = item.Elem()
	}
	if item.Kind() != reflect.Struct {
		return nil, errors.New("item format should be struct")
	}
	var fieldKinds []reflect.Kind
	var fieldTypes []reflect.Type
	for i := 0; i < item.NumField(); i++ {
		kind := item.Field(i).Kind()
		t := item.Field(i).Type()
		fieldKinds = append(fieldKinds, kind)
		fieldTypes = append(fieldTypes, t)
	}

	var fieldNames []string

	val := reflect.Indirect(item)
	for i := 0; i < item.NumField(); i++ {
		fieldNames = append(fieldNames, val.Type().Field(i).Name)
	}

	var data = reflect.New(reflect.TypeOf(c.format))

	if len(fieldKinds) == len(values) && fieldKinds != nil && fieldNames != nil && fieldTypes != nil && data.Kind() == reflect.Ptr {
		for i, v := range values {
			typeOf, err := c.typeof(v, fieldKinds[i], fieldTypes[i])
			if err == nil && typeOf != nil {
				field := data.Elem().FieldByName(fieldNames[i])
				value := reflect.ValueOf(typeOf)

				if value.Kind() == reflect.Ptr {
					value = value.Elem()
				}
				field.Set(value)

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
		if c.cf == nil {
			return nil, errors.New("custom field converter is nil")
		}
		return c.cf(v, k, t)
	}
}
