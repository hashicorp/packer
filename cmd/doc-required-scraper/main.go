package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

const (
	DocsUrl  = "https://www.packer.io/docs/"
	CacheDir = "cache/"
)

func main() {
	c := colly.NewCollector()

	// Find and visit all doc pages
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		url := e.Attr("href")
		if !strings.HasPrefix(url, "/docs/builders") {
			return
		}
		e.Request.Visit(url)
	})

	c.OnHTML("#required- + ul a[name]", func(e *colly.HTMLElement) {

		builder := e.Request.URL.Path[strings.Index(e.Request.URL.Path, "/builders/")+len("/builders/"):]
		builder = strings.TrimSuffix(builder, ".html")

		text := e.DOM.Parent().Text()
		text = strings.ReplaceAll(text, "\n", "")
		text = strings.TrimSpace(text)

		fmt.Printf("required: %25s builder: %20s text: %s\n", e.Attr("name"), builder, text)
	})

	c.CacheDir = CacheDir

	c.Visit(DocsUrl)
}
