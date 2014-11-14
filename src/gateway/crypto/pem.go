package crypto

import (
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
)

// PEMDataFromPath returns the data of the first block in the PEM encoded file at path.
func PEMDataFromPath(path string, blockType string) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return []byte{}, err
	}
	return PEMDataFromData(data, blockType)
}

// PEMDataFromData returns the data of the first block in the PEM encoded data.
func PEMDataFromData(data []byte, blockType string) ([]byte, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return []byte{}, errors.New("Data is not PEM encoded")
	}

	if block.Type != blockType {
		return []byte{}, fmt.Errorf("Expected '%s' as first block, got '%s'",
			blockType, block.Type)
	}

	return block.Bytes, nil
}
