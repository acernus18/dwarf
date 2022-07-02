package serializeutil

import "encoding/json"

func Serialize[T any](object T) []byte {
	result, err := json.Marshal(object)
	if err != nil {
		return []byte("")
	}
	return result
}

func Deserialize[T any](data []byte) (T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return result, err
	}
	return result, nil
}
