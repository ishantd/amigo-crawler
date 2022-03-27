package main

import (
	"fmt"
	"log"
	"time"

	"github.com/blevesearch/bleve/v2"
	badger "github.com/dgraph-io/badger/v3"
	uuid "github.com/google/uuid"
)

type WikiDoc struct {
	Title string
	Body  string
}


func timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    log.Printf("%s took %s", name, elapsed)
}

func main() {
	// open a new index
	defer timeTrack(time.Now(), "main")
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("test7.bleve", mapping)

	if err != nil {
		fmt.Println(err)
		return
	}

	batch := index.NewBatch()

	i := 0
	db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
	db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
		item := it.Item()
		k := item.Key()
		err := item.Value(func(v []byte) error {
			wiki_doc := WikiDoc{Title: string(k), Body: string(v)}
			batch.Index(uuid.New().String(), wiki_doc)
			i++
			if i > 100 {
				index.Batch(batch)
				i = 0
			}
			return err
		})
		if err != nil {
			return err
		}
		}
		return err
	})


	// search for some text
	query := bleve.NewMatchQuery("computer")
	search := bleve.NewSearchRequest(query)
	searchResults, err := index.Search(search)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(searchResults)
}