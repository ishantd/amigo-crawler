package main

import (
	"fmt"
	"log"

	badger "github.com/dgraph-io/badger/v3"
)

func main() {
  // Open the Badger database located in the /tmp/badger directory.
  // It will be created if it doesn't exist.
  var count int = 0
  db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
  db.View(func(txn *badger.Txn) error {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchSize = 10
	it := txn.NewIterator(opts)
	defer it.Close()
	for it.Rewind(); it.Valid(); it.Next() {
	  item := it.Item()
	  err := item.Value(func(v []byte) error {
		count += 1
		return err
	  })
	  if err != nil {
		return err
	  }
	}
	return nil
  })
  
  if err != nil {
	  log.Fatal(err)
  }
  defer db.Close()
  
  fmt.Println("count: ", count)

}