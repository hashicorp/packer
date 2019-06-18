package request

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// ToQueryMap will convert a request to string map
func ToQueryMap(req Common) (map[string]string, error) {
	v := reflect.ValueOf(req)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	return encode(&v, "")
}

func encodeOne(v *reflect.Value) (string, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.String:
		return v.String(), nil
	case reflect.Ptr:
		ptrValue := v.Elem()
		return encodeOne(&ptrValue)
	default:
		message := fmt.Sprintf(
			"Invalid variable type, type must be one of int-, uint-,"+
				" float-, bool, string and ptr, got %s",
			v.Kind().String(),
		)
		return "", errors.New(message)
	}
}

func encode(v *reflect.Value, prefix string) (map[string]string, error) {
	result := make(map[string]string)

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		name := v.Type().Field(i).Name
		if prefix != "" && prefix != "CommonBase" {
			name = fmt.Sprintf("%s.%s", prefix, name)
		}

		// skip unexported field
		if !f.CanSet() {
			continue
		}

		// find the real value of pointer
		// such as **struct to struct
		for f.Kind() == reflect.Ptr {
			if f.IsNil() {
				break
			}
			f = f.Elem()
		}

		// check if non-pointer
		if f.Kind() == reflect.Ptr && f.IsNil() {
			continue
		}

		switch f.Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < f.Len(); i++ {
				item := f.Index(i)
				if item.Kind() == reflect.Ptr && item.IsNil() {
					continue
				}

				keyPrefix := fmt.Sprintf("%s.%v", name, i)

				if item.Kind() == reflect.Struct {
					kv, err := encode(&item, keyPrefix)
					if err != nil {
						return result, err
					}

					for k, v := range kv {
						if v != "" {
							result[k] = v
						}
					}
				} else {
					s, err := encodeOne(&item)
					if err != nil {
						return result, err
					}

					if s != "" {
						result[keyPrefix] = s
					}
				}
			}
		case reflect.Struct:
			m, err := encode(&f, name)
			if err != nil {
				return result, err
			}

			// flatten composited struct into result map
			for k, v := range m {
				result[k] = v
			}
		default:
			s, err := encodeOne(&f)
			if err != nil {
				return result, err
			}

			// set field value into result
			if s != "" {
				result[name] = s
			}
		}
	}

	return result, nil
}
