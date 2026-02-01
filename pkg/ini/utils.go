package ini

import (
	"errors"
	"reflect"
	"strings"

	"github.com/hashicorp/go-hclog"
)

func structForRange(s interface{}, callback func(field reflect.StructField, currentValue interface{}) interface{}, logger hclog.Logger) error {
	v := reflect.ValueOf(s)

	if v.Kind() != reflect.Ptr {
		return errors.New("expected a pointer to struct")
	}

	v = v.Elem()

	if v.Kind() != reflect.Struct {
		return errors.New("not a struct")
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		currentValue := field.Interface()

		newValue := callback(fieldType, currentValue)

		if newValue == nil {
			continue
		}

		newVal := reflect.ValueOf(newValue)

		if field.Type() != newVal.Type() {
			if newVal.CanConvert(field.Type()) {
				newVal = newVal.Convert(field.Type())
			} else {
				logger.Warn("Type mismatch for field",
					"name", fieldType.Name,
					"expected", field.Type(),
					"got", newVal.Type(),
				)

				continue
			}
		}

		if !field.CanSet() {
			logger.Warn("Field type cannot be set (possibly private)", "name", fieldType.Name)
			continue
		}

		field.Set(newVal)
	}

	return nil
}

func Split(s string, sep string) []string {
	slice := strings.Split(s, sep)
	for i, str := range slice {
		slice[i] = strings.TrimSpace(str)
	}

	return slice
}
