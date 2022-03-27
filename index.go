package main

import (
	"fmt"

	"github.com/blevesearch/bleve/v2"
	badger "github.com/dgraph-io/badger/v3"
)

type WikiDoc struct {
	Title string
	Body  string
}


func main() {
	// open a new index
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("test2.bleve", mapping)
	if err != nil {
		fmt.Println(err)
		return
	}


	index_data := []WikiDoc{}

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
			index_data = append(index_data, wiki_doc)
			return err
		})
		if err != nil {
			return err
		}
		}
		return err
	})

	index.Index("id", index_data)

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