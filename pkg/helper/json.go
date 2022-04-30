package helper

import (
	"encoding/json"
	"os"
)

func LoadJsonFile[TOut any](path string) (*TOut, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var obj *TOut
	if err = json.Unmarshal(fileBytes, &obj); err != nil {
		return nil, err
	}

	return obj, nil
}
