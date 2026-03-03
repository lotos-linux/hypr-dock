package ini

import (
	"reflect"

	"github.com/hashicorp/go-hclog"
)

type Manager struct {
	raw    map[string]map[string]string
	path   string
	logger hclog.Logger
}

func New(file string, logger hclog.Logger, globalKey ...string) *Manager {
	initialSection := "General"
	if len(globalKey) != 0 && globalKey[0] != "" {
		initialSection = globalKey[0]
	}

	raw, err := GetMap(file, initialSection)
	if err != nil {
		raw = make(map[string]map[string]string, 0)
		logger.Error("INI PARSING",
			"error", err,
		)
	}

	return &Manager{
		raw:    raw,
		path:   file,
		logger: logger,
	}
}

func (cm *Manager) GetSection(name string) *Section {
	sraw, exist := cm.raw[name]
	if !exist {
		cm.logger.Warn("Config section not found",
			"section", name,
			"file", cm.path,
		)

		sraw = make(map[string]string, 0)
	}

	return NewSection(sraw, cm.logger)
}

func (cm *Manager) ParseSection(v interface{}, name string) (*Section, error) {
	section := cm.GetSection(name)
	err := section.Unmarshal(v)
	if err != nil {
		return nil, err
	}
	return section, nil
}

func (cm *Manager) Unmarshal(v interface{}) error {
	return structForRange(v, func(field reflect.StructField, currentValue interface{}) interface{} {
		if field.Type.Kind() != reflect.Struct {
			return nil
		}

		sectionName, ok := field.Tag.Lookup("section")
		if !ok {
			return nil
		}

		newStructPtr := reflect.New(field.Type)
		newStructInterface := newStructPtr.Interface()

		_, err := cm.ParseSection(newStructInterface, sectionName)
		if err != nil {
			cm.logger.Error("Failed to parse section",
				"section", sectionName,
				"name", field.Name,
				"error", err,
			)

			return nil
		}

		return newStructPtr.Elem().Interface()
	}, cm.logger)
}
