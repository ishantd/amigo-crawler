package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/gocolly/colly"
)

func IsUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}


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

func crawlWikipedia(article string, db *badger.DB) (string) {

	split_url := strings.Split(article, ":")
    article_name := split_url[len(split_url)-1]
	article_urlified := strings.ReplaceAll(split_url[len(split_url)-1], " ", "_")
	crawl_url := "https://en.wikipedia.org/wiki/" + article_urlified

    if !IsUrl(crawl_url) {
        fmt.Println(crawl_url)
        return ""
    }
	
    c := colly.NewCollector(
        colly.AllowedDomains("en.wikipedia.org"),
    )

    c.OnHTML("body", func(e *colly.HTMLElement) {
        data := e.ChildText("p")
        db.Update(func(txn *badger.Txn) error {
            err := txn.Set([]byte(article_name), []byte(data))
            return err
        })
    })


    c.Visit(crawl_url)
	return "true"
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

    lines, err := readLines("../wiki1k.txt")
    if err != nil {
        log.Fatalf("readLines: %s", err)
    }

    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for article := range in {
                res := crawlWikipedia(article, db)
                fmt.Println(res)
            }
        }()
    }

    for _, line := range lines {
        if line != "" {
            in <- line
        }
    }
    close(in)
    wg.Wait()
}