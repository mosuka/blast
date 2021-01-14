package mapping

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/blevesearch/bleve/v2/mapping"
)

func NewIndexMapping() *mapping.IndexMappingImpl {
	return mapping.NewIndexMapping()
}

func NewIndexMappingFromBytes(indexMappingBytes []byte) (*mapping.IndexMappingImpl, error) {
	indexMapping := mapping.NewIndexMapping()

	if err := indexMapping.UnmarshalJSON(indexMappingBytes); err != nil {
		return nil, err
	}

	if err := indexMapping.Validate(); err != nil {
		return nil, err
	}

	return indexMapping, nil
}

func NewIndexMappingFromMap(indexMappingMap map[string]interface{}) (*mapping.IndexMappingImpl, error) {
	indexMappingBytes, err := json.Marshal(indexMappingMap)
	if err != nil {
		return nil, err
	}

	indexMapping, err := NewIndexMappingFromBytes(indexMappingBytes)
	if err != nil {
		return nil, err
	}

	return indexMapping, nil
}

func NewIndexMappingFromFile(indexMappingPath string) (*mapping.IndexMappingImpl, error) {
	_, err := os.Stat(indexMappingPath)
	if err != nil {
		if os.IsNotExist(err) {
			// does not exist
			return nil, err
		}
		// other error
		return nil, err
	}

	// read index mapping file
	indexMappingFile, err := os.Open(indexMappingPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = indexMappingFile.Close()
	}()

	indexMappingBytes, err := ioutil.ReadAll(indexMappingFile)
	if err != nil {
		return nil, err
	}

	indexMapping, err := NewIndexMappingFromBytes(indexMappingBytes)
	if err != nil {
		return nil, err
	}

	return indexMapping, nil
}
