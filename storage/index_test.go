package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/log"
	"github.com/mosuka/blast/mapping"
	"github.com/mosuka/blast/util"
)

func TestClose(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dir := filepath.Join(tmpDir, "index")

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	index, err := NewIndex(dir, indexMapping, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if index == nil {
		t.Fatal("failed to create index")
	}

	if err := index.Close(); err != nil {
		t.Fatalf("%v", err)
	}
}

func TestIndex(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dir := filepath.Join(tmpDir, "index")

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	index, err := NewIndex(dir, indexMapping, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if index == nil {
		t.Fatal("failed to create index")
	}

	id := "1"
	fields := map[string]interface{}{
		"title":     "Search engine (computing)",
		"text":      "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"_type":     "example",
	}

	if err := index.Index(id, fields); err != nil {
		t.Fatal("failed to index document")
	}

	if err := index.Close(); err != nil {
		t.Fatalf("%v", err)
	}
}

func TestGet(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dir := filepath.Join(tmpDir, "index")

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	index, err := NewIndex(dir, indexMapping, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if index == nil {
		t.Fatal("failed to create index")
	}

	id := "1"
	fields := map[string]interface{}{
		"title":     "Search engine (computing)",
		"text":      "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"_type":     "example",
	}

	if err := index.Index(id, fields); err != nil {
		t.Fatal("failed to index document")
	}

	f, err := index.Get(id)
	if err != nil {
		t.Fatal("failed to get document")
	}
	if fields["title"].(string) != f["title"].(string) {
		t.Fatalf("expected content to see %v, saw %v", fields["title"].(string), f["title"].(string))
	}

	if err := index.Close(); err != nil {
		t.Fatalf("%v", err)
	}
}

func TestDelete(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dir := filepath.Join(tmpDir, "index")

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	index, err := NewIndex(dir, indexMapping, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if index == nil {
		t.Fatal("failed to create index")
	}

	id := "1"
	fields := map[string]interface{}{
		"title":     "Search engine (computing)",
		"text":      "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"_type":     "example",
	}

	if err := index.Index(id, fields); err != nil {
		t.Fatal("failed to index document")
	}

	fields, err = index.Get(id)
	if err != nil {
		t.Fatal("failed to get document")
	}

	if err := index.Delete(id); err != nil {
		t.Fatal("failed to delete document")
	}

	fields, err = index.Get(id)
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			// ok
		default:
			t.Fatal("failed to get document")
		}
	}

	if err := index.Close(); err != nil {
		t.Fatalf("%v", err)
	}
}

func TestBulkIndex(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dir := filepath.Join(tmpDir, "index")

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	index, err := NewIndex(dir, indexMapping, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if index == nil {
		t.Fatal("failed to create index")
	}

	docs := make([]map[string]interface{}, 0)
	for i := 1; i <= 100; i++ {
		id := strconv.Itoa(i)
		fields := map[string]interface{}{
			"title":     fmt.Sprintf("Search engine (computing) %d", i),
			"text":      fmt.Sprintf("A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web. %d", i),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"_type":     "example",
		}
		doc := map[string]interface{}{
			"id":     id,
			"fields": fields,
		}

		docs = append(docs, doc)
	}

	count, err := index.BulkIndex(docs)
	if err != nil {
		t.Fatal("failed to index documents")
	}
	if count <= 0 {
		t.Fatal("failed to index documents")
	}

	if err := index.Close(); err != nil {
		t.Fatalf("%v", err)
	}
}

func TestBulkDelete(t *testing.T) {
	curDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}

	tmpDir := util.TmpDir()
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	dir := filepath.Join(tmpDir, "index")

	indexMapping, err := mapping.NewIndexMappingFromFile(filepath.Join(curDir, "../examples/example_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}

	logger := log.NewLogger("WARN", "", 500, 3, 30, false)

	index, err := NewIndex(dir, indexMapping, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if index == nil {
		t.Fatal("failed to create index")
	}

	docs := make([]map[string]interface{}, 0)
	for i := 1; i <= 100; i++ {
		id := strconv.Itoa(i)
		fields := map[string]interface{}{
			"title":     fmt.Sprintf("Search engine (computing) %d", i),
			"text":      fmt.Sprintf("A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web. %d", i),
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"_type":     "example",
		}
		doc := map[string]interface{}{
			"id":     id,
			"fields": fields,
		}

		docs = append(docs, doc)
	}

	count, err := index.BulkIndex(docs)
	if err != nil {
		t.Fatal("failed to index documents")
	}
	if count <= 0 {
		t.Fatal("failed to index documents")
	}

	ids := make([]string, 0)
	for i := 1; i <= 100; i++ {
		id := strconv.Itoa(i)

		ids = append(ids, id)
	}

	count, err = index.BulkDelete(ids)
	if err != nil {
		t.Fatal("failed to delete documents")
	}
	if count <= 0 {
		t.Fatal("failed to delete documents")
	}

	if err := index.Close(); err != nil {
		t.Fatalf("%v", err)
	}
}
