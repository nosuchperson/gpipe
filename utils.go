package gpipe

import (
	"gopkg.in/yaml.v3"
)

func ConfigMapUnmarshal[T any](m interface{}, ptr *T) (*T, error) {
	if data, err := yaml.Marshal(m); err != nil {
		return nil, err
	} else if err := yaml.Unmarshal(data, ptr); err != nil {
		return nil, err
	} else {
		return ptr, nil
	}
}
