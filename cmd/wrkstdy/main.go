// Command wrkstdy builds wrkstdy.com from the markdown in content/.
//
//	wrkstdy build [-out public] [-drafts] [-baseurl URL]
//	wrkstdy serve [-addr localhost:1313] [-drafts]
//	wrkstdy new blog/my-post
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"wrkstdy.com/site/internal/site"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		usage()
	}

	var err error
	switch cmd, args := os.Args[1], os.Args[2:]; cmd {
	case "build":
		err = runBuild(args)
	case "serve":
		err = runServe(args)
	case "new":
		err = runNew(args)
	default:
		usage()
	}
	if err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: wrkstdy build | serve | new <section/slug>")
	os.Exit(2)
}

func runBuild(args []string) error {
	flags := flag.NewFlagSet("build", flag.ExitOnError)
	out := flags.String("out", "public", "output folder")
	drafts := flags.Bool("drafts", false, "include draft pages")
	baseURL := flags.String("baseurl", "", "override baseURL from site.toml")
	flags.Parse(args)

	return build(site.Options{Root: ".", Out: *out, Drafts: *drafts, BaseURL: *baseURL})
}

func build(opts site.Options) error {
	start := time.Now()
	n, err := site.Build(opts)
	if err != nil {
		return err
	}
	log.Printf("built %d pages into %s/ in %s", n, opts.Out, time.Since(start).Round(time.Millisecond))
	return nil
}
