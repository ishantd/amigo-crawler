package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	// "github.com/gocolly/colly"
)

func readLines(path string) ([]string, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    return lines, scanner.Err()
}

func timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    log.Printf("%s took %s", name, elapsed)
}

func crawlWikipedia(article string) () {

	split_url := strings.Split(article, ":")
	article_urlified := strings.ReplaceAll(split_url[len(split_url)-1], " ", "_")
	crawl_url := "https://en.wikipedia.org/wiki/" + article_urlified

    if strings.Contains(article, "%") {
        return
    }
	
    client := &http.Client{}

    res, err := http.NewRequest("GET", crawl_url, nil)
    if err != nil {
        log.Printf("Error fetching: %v", err)
    }
    
    resp, err := client.Do(res)

    if err != nil {
        return 
    }
    defer resp.Body.Close()

	return
}

func main() {

    db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    const workers = 250
    
    wg := new(sync.WaitGroup)
    in := make(chan string, 2*workers)
    defer timeTrack(time.Now(), "main")

    fmt.Println("Starting to crawl Wikipedia")

    lines, err := readLines("../wiki1m.txt")
    if err != nil {
        log.Fatalf("readLines: %s", err)
    }

    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for article := range in {
                crawlWikipedia(article)
            }
        }()
    }

    for _, line := range lines {
        if line != "" {
            in <- line
            dberr := db.Update(func(txn *badger.Txn) error {
                err := txn.Set([]byte(line), []byte(time))
                return err
            })
        }
    }
    close(in)
    wg.Wait()
}