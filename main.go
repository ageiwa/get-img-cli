package main

import (
	"flag"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

func downloadFromLink(title string, index int, link string) error {
	resp, err := http.Get(link)
	if err != nil {
		return err
	}

	builder := strings.Builder{}
	builder.WriteString(strconv.Itoa(index))
	builder.WriteString(filepath.Ext(link))

	file, err := os.Create(filepath.Join(title, builder.String()))
	if err != nil {
		return err
	}

	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return err
	}

	return nil
}

func main() {
	fLink := flag.String("from", "", "use for set link")
	flag.Parse()

	title := ""
	links := make([]string, 0)

	resp, err := http.Get(*fLink)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "title" {
			c := n.FirstChild
			title = c.Data
		}

		if n.Type == html.ElementNode && n.Data == "img" {
			for _, a := range n.Attr {
				if a.Key == "src" {
					links = append(links, a.Val)
				}
			}
		}
	}

	if err := os.MkdirAll(title, 0666); err != nil {
		log.Fatal(err)
	}

	wg := sync.WaitGroup{}

	for index, link := range links {
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := downloadFromLink(title, index, link)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	wg.Wait()
}
