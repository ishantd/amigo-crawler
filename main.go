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

func crawlWikipedia(article string, db *badger.DB) (int) {

	split_url := strings.Split(article, ":")
    article_name := split_url[len(split_url)-1]
	article_urlified := strings.ReplaceAll(split_url[len(split_url)-1], " ", "_")
	crawl_url := "https://en.wikipedia.org/wiki/" + article_urlified

    if !IsUrl(crawl_url) {
        fmt.Println(crawl_url)
        return -1
    }
	
    c := colly.NewCollector(
        colly.AllowedDomains("en.wikipedia.org"),
    )
    var db_err error
    c.OnHTML("body", func(e *colly.HTMLElement) {
        data := e.ChildText("p")
        db_err = db.Update(func(txn *badger.Txn) error {
            err := txn.Set([]byte(article_name), []byte(data))
            fmt.Println("Saved: ", crawl_url, " ", article_name)
            return err
        })
    })
    if db_err != nil {
        fmt.Println(db_err)
        return -1
    }


    c.Visit(crawl_url)
	return 1
}

func main() {

    db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
    if err != nil {
        fmt.Println(err)
    }

    var total_crawl int
    fmt.Println("Enter the number of articles to crawl: ")
    fmt.Scanln(&total_crawl)

    defer db.Close()
    const workers = 250
    
    wg := new(sync.WaitGroup)
    in := make(chan string, 2*workers)
    defer timeTrack(time.Now(), "main")

    fmt.Println("Starting to crawl Wikipedia")

    lines_all, err := readLines("../wiki.txt")
    lines := lines_all[:total_crawl]

    
    if err != nil {
        fmt.Println("readLines: %s", err)
    }
    var count int = 0
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for article := range in {
                res := crawlWikipedia(article, db)
                count += res
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
    fmt.Println("Total pages crawled and saved into badger DB: ", count)
}