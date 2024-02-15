package types

import (
	"encoding/json"
	"errors"
)

type AppSettings map[any]any

func (m AppSettings) MarshalJSON() ([]byte, error) {
	formattedMap := map[string]any{}
	for k, v := range m {
		keyString, ok := k.(string)
		if !ok {
			return nil, errors.New("non-string key found in map")
		}
		formattedMap[keyString] = v
	}
	return json.Marshal(formattedMap)
}

func (m AppSettings) UnmarshalJSON(data []byte) error {
	var rawMap map[string]any
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return err
	}
	for k, v := range rawMap {
		m[k] = v
	}
	return nil
}
