package utils

import (
	"bytes"
	"encoding/gob"
	"hash/fnv"
)

func Hash[T any](input T) (int, error) {
	buffer := bytes.NewBuffer([]byte{})
	encoder := gob.NewEncoder(buffer)
	err := encoder.Encode(input)
	if err != nil {
		return 0, err
	}
	hasher := fnv.New32a()
	hasher.Write(buffer.Bytes())
	return int(hasher.Sum32()), nil
}
