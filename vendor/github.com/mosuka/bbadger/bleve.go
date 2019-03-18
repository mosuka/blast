package bbadger

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

// BleveIndex a helper function that open (creates if not exists a new) bleve index
func BleveIndex(path string, mapping mapping.IndexMapping) (bleve.Index, error) {
	index, err := bleve.NewUsing(path, mapping, bleve.Config.DefaultIndexType, Name, nil)

	if err != nil && err == bleve.ErrorIndexPathExists {
		index, err = bleve.Open(path)
	}

	return index, err
}
