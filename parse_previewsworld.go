package main

import (
	"encoding/csv"
	"strings"
	"log"
	"os"
	"io"
	"bufio"
  "io/ioutil"
	// "fmt"

	"github.com/gocolly/colly"
)

var (
    Trace   *log.Logger
    Info    *log.Logger
    Warning *log.Logger
    Error   *log.Logger
)

func Init(
    traceHandle io.Writer,
    infoHandle io.Writer,
    warningHandle io.Writer,
    errorHandle io.Writer) {

    Trace = log.New(traceHandle,
        "TRACE: ",
        log.Ldate|log.Ltime|log.Lshortfile)

    Info = log.New(infoHandle,
        "INFO: ",
        log.Ldate|log.Ltime|log.Lshortfile)

    Warning = log.New(warningHandle,
        "WARNING: ",
        log.Ldate|log.Ltime|log.Lshortfile)

    Error = log.New(errorHandle,
        "ERROR: ",
        log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	// Some logging stuff
	Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	// Init(os.Stdout, os.Stdout, os.Stdout, os.Stderr)

	// Open read csv
	truAll, _ := os.Open("truall.csv")
	reader := csv.NewReader(bufio.NewReader(truAll))
	reader.LazyQuotes = true
	var diamondIds []string

	for {
		line, error := reader.Read()
		if error == io.EOF {
			Info.Printf("Finished retrieving all URLs in %q for results\n", truAll)
			break
		} else if error != nil {
			Error.Fatal(error)
		}
		diamondIds = append(diamondIds, line[0])
	}

	// Open write CSV
	fName := "previews_world.csv"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fName, err)
		return
	}

	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	// Write CSV header
	writer.Write([]string{"DiamondID", "StockList#", "PreviewsWorldImageURL", "DiamondImageURL"})

	// Instantiate default collector
	c := colly.NewCollector(
		// Allow requests only to previews world and diamond
		colly.AllowedDomains("previewsworld.com", "retailerservices.diamondcomics.com"),

		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./previews_world_cache"),
	)

	for _, v := range diamondIds {
		Info.Println("https://previewsworld.com/Catalog/"+v[0:9])
		continue
		// Extract product details
		c.OnHTML(".mainContentImage", func(e *colly.HTMLElement) {
			writer.Write([]string{
				v[0:9],
				strings.Split(strings.Split(e.ChildAttr("img", "src"), "/")[3], "?")[0],
				e.Request.AbsoluteURL(e.ChildAttr("a", "href")),
				"https://retailerservices.diamondcomics.com/Image/Resource/2/" + strings.Split(strings.Split(e.ChildAttr("img", "src"), "/")[3], "?")[0]+".jpg",
			})
		})

		c.OnRequest(func(r *colly.Request) {
			Trace.Println("Visiting", r.URL)
		})
		
		c.Visit("https://previewsworld.com/Catalog/"+v[0:9])
	}

	Info.Printf("Scraping finished, check file %q for results\n", fName)

	// Display collector's statistics
	Info.Println(c)
}