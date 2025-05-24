package utils

import (
	"errors"
	"fmt"
	"os"
)

func ReadBinary(fileName string) ([]byte, error) {
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		return bytes, errors.New(fmt.Sprintf("error while opening file: %v\n", err))
	}

	return bytes, nil
}
