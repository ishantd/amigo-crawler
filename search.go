package main

import (
	"fmt"

	"github.com/blevesearch/bleve/v2"
)




func main() {

	// take user input

	// var user_input string
	// fmt.Print("Enter a search term: ")
	// fmt.Scanln(&user_input)
	user_input := "anti"
	fmt.Println("Searching for:", user_input)
	// open a new index

	index, err := bleve.Open("test7.bleve")
	fmt.Println(index)
	if err != nil {
		fmt.Println(err)
		return
	}

	query := bleve.NewMatchQuery(user_input)
	search := bleve.NewSearchRequest(query)
	search.Highlight = bleve.NewHighlightWithStyle("html")
	search.Fields = []string{"Title"}
	search.Size = 10
	searchResults, err := index.Search(search)
	if err != nil {
		fmt.Println(err)
		return
	}
	search.Highlight = bleve.NewHighlight()
	fmt.Println(searchResults)
	fmt.Println(search.Highlight)
}