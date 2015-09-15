package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"time"

	"github.com/miku/oaimi"
)

func main() {

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	cacheDir := flag.String("cache", path.Join(usr.HomeDir, ".oaimi"), "oaimi cache dir")
	set := flag.String("set", "", "OAI set")
	prefix := flag.String("prefix", "oai_dc", "OAI metadataPrefix")
	from := flag.String("from", "2000-01-01", "OAI from")
	until := flag.String("until", time.Now().Format("2006-01-02"), "OAI until")
	retry := flag.Int("retry", 10, "retry count for exponential backoff")
	dirname := flag.Bool("dirname", false, "show shard directory for request")
	verbose := flag.Bool("verbose", false, "more output")
	showVersion := flag.Bool("v", false, "prints current program version")

	flag.Parse()

	if *showVersion {
		fmt.Println(oaimi.Version)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		log.Fatal("URL to OAI endpoint required")
	}

	From, err := time.Parse("2006-01-02", *from)
	if err != nil {
		log.Fatal(err)
	}

	Until, err := time.Parse("2006-01-02", *until)
	if err != nil {
		log.Fatal(err)
	}

	if Until.Before(From) {
		log.Fatal(oaimi.ErrInvalidDateRange)
	}

	endpoint := flag.Arg(0)

	if _, err := os.Stat(*cacheDir); os.IsNotExist(err) {
		err := os.MkdirAll(*cacheDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	req := oaimi.BatchedRequest{
		Cache: oaimi.Cache{Directory: *cacheDir},
		Request: oaimi.Request{
			Verbose:  *verbose,
			Verb:     "ListRecords",
			Set:      *set,
			Prefix:   *prefix,
			From:     From,
			Until:    Until,
			Endpoint: endpoint,
			MaxRetry: *retry,
		},
	}

	if *dirname {
		req := oaimi.CachedRequest{
			Cache: oaimi.Cache{Directory: *cacheDir},
			Request: oaimi.Request{
				Set:      *set,
				Prefix:   *prefix,
				Endpoint: endpoint,
			},
		}
		fmt.Println(filepath.Dir(req.Path()))
		os.Exit(0)
	}

	w := bufio.NewWriter(os.Stdout)
	err = req.Do(w)
	if err != nil {
		log.Fatal(err)
	}
	w.Flush()
}
