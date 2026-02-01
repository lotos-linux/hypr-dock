package ini

import (
	"log"
	"reflect"
	"slices"
	"strconv"

	"github.com/hashicorp/go-hclog"
)

type Section struct {
	raw    map[string]string
	logger hclog.Logger
}

func NewSection(raw map[string]string, logger hclog.Logger) *Section {
	return &Section{
		raw:    raw,
		logger: logger,
	}
}

func (b *Section) String(key string, def string, validateList []string) string {
	val, exist := b.raw[key]
	if !exist {
		b.logger.Warn("Config key not found", "key", key)
		return def
	}

	if validateList != nil {
		if !slices.Contains(validateList, val) {
			b.logger.Warn("Config value not valid", key, val)
			return def
		}
	}

	return val
}

func (b *Section) Strings(key string, def []string, sep ...string) []string {
	val, exist := b.raw[key]
	if !exist {
		b.logger.Warn("Config key not found", "key", key)
		return def
	}

	separator := ","
	if len(sep) > 0 && sep[0] != "" {
		separator = sep[0]
	}

	list := Split(val, separator)
	return list
}

func (b *Section) Bool(key string, def bool) bool {
	val, exist := b.raw[key]
	if !exist {
		b.logger.Warn("Config key not found", "key", key)
		return def
	}

	return val == "true"
}

func (b *Section) Int(key string, def int, validator func(int) bool) int {
	val, exist := b.raw[key]
	if !exist {
		b.logger.Warn("Config key not found", "key", key)
		return def
	}

	res, err := strconv.Atoi(val)
	if err != nil {
		b.logger.Warn("Config value not a number", key, val)
		return def
	}

	if validator != nil {
		if !validator(res) {
			b.logger.Warn("Config value not valid", key, val)
			return def
		}
	}

	return res
}

func (b *Section) Float32(key string, def float32, validator func(float32) bool) float32 {
	val, exist := b.raw[key]
	if !exist {
		b.logger.Warn("Config key not found", "key", key)
		return def
	}

	res, err := strconv.ParseFloat(val, 32)
	if err != nil {
		b.logger.Warn("Config value not a number", key, val)
		return def
	}

	res32 := float32(res)

	if validator != nil {
		if !validator(res32) {
			b.logger.Warn("Config value not valid", key, val)
			return def
		}
	}

	return res32
}

func (b *Section) Float64(key string, def float64, validator func(float64) bool) float64 {
	val, exist := b.raw[key]
	if !exist {
		b.logger.Warn("Config key not found", "key", key)
		return def
	}

	res, err := strconv.ParseFloat(val, 64)
	if err != nil {
		b.logger.Warn("Config value not a number", key, val)
		return def
	}

	if validator != nil {
		if !validator(res) {
			b.logger.Warn("Config value not valid", key, val)
			return def
		}
	}

	return res
}

func (b *Section) _string(key string, def string, valid string) string {
	validList := Split(valid, ",")
	return b.String(key, def, validList)
}

func (b *Section) _strings(key string, def string, sep ...string) []string {
	defList := []string{}
	if def != "" {
		defList = Split(def, ",")
	}

	return b.Strings(key, defList, sep...)
}

func (b *Section) _int(key string, def string, max string, min string) int {
	defI, err := strconv.Atoi(def)
	if err != nil {
		defI = 0
	}

	minI, err := strconv.Atoi(min)
	if err != nil {
		minI = 0
	}

	maxI, err := strconv.Atoi(max)
	ismax := err == nil

	return b.Int(key, defI, func(i int) bool {
		if ismax && maxI < i {
			log.Println(ismax, maxI)
			return false
		}

		return i >= minI
	})
}

func (b *Section) _float64(key string, def string, max string, min string) float64 {
	def64, err := strconv.ParseFloat(def, 64)
	if err != nil {
		def64 = 0
	}

	min64, err := strconv.ParseFloat(min, 64)
	if err != nil {
		min64 = 0
	}

	max64, err := strconv.ParseFloat(max, 64)
	ismax := err == nil

	return b.Float64(key, def64, func(f float64) bool {
		if ismax && max64 < f {
			return false
		}

		return f >= min64
	})
}

func (b *Section) _float32(key string, def string, max string, min string) float32 {
	def64, err := strconv.ParseFloat(def, 64)
	def32 := float32(def64)
	if err != nil {
		def64 = 0
	}

	min64, err := strconv.ParseFloat(min, 64)
	min32 := float32(min64)
	if err != nil {
		min64 = 0
	}

	max64, err := strconv.ParseFloat(max, 64)
	max32 := float32(max64)
	ismax := err == nil

	return b.Float32(key, def32, func(f float32) bool {
		if ismax && max32 < f {
			return false
		}

		return f >= min32
	})
}

func (b *Section) Unmarshal(s interface{}) error {
	return structForRange(s, func(field reflect.StructField, fieldValue interface{}) interface{} {
		key, ok := field.Tag.Lookup("key")
		if !ok {
			key = field.Name
		}

		def := field.Tag.Get("def")

		switch field.Type.Kind() {

		// STRING
		case reflect.String:
			valid, ok := field.Tag.Lookup("valid")
			if !ok {
				return b.String(key, def, nil)
			}
			return b._string(key, def, valid)

		// SLICE
		case reflect.Slice:
			sep := field.Tag.Get("sep")
			return b._strings(key, def, sep)

		// BOOL
		case reflect.Bool:
			return b.Bool(key, def == "true")

		// INT
		case reflect.Int:
			max := field.Tag.Get("max")
			min := field.Tag.Get("min")

			return b._int(key, def, max, min)

		// FLOAT32
		case reflect.Float32:
			max := field.Tag.Get("max")
			min := field.Tag.Get("min")

			return b._float32(key, def, max, min)

		// FLOAT64
		case reflect.Float64:
			max := field.Tag.Get("max")
			min := field.Tag.Get("min")

			return b._float64(key, def, max, min)
		}

		return nil
	}, b.logger)
}
