package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/bcap/flamegrapher/assets"
)

func main() {
	var fileName string
	var regexStr string

	flag.StringVar(&fileName, "file", "", "which file to load data from. If not passed, data is read from stdin")
	flag.StringVar(&regexStr, "sep", `\s+`, "which regular expression to use as field separator")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	panicOnErr(validateConflictingAssets())

	separator, err := regexp.Compile(regexStr)
	panicOnErr(err)

	reader := os.Stdin
	source := "stdin"
	if fileName != "" {
		file, err := os.Open(fileName)
		panicOnErr(err)
		reader = file
		source = "file " + fileName
	}

	log.Printf("Reading data from %s and splitting data with regex %s", source, regexStr)

	tree, err := BuildTree(ctx, reader, separator)
	panicOnErr(err)

	tree.Value = "TOTAL"

	encodedFlameGraph, err := json.Marshal(tree.ToFlameGraph())
	panicOnErr(err)

	fetcher := func(ctx context.Context) (json.RawMessage, error) {
		return encodedFlameGraph, nil
	}

	server := NewServer(8080, fetcher)
	if err := server.Run(ctx); err != context.Canceled {
		panic(err)
	}
}

func validateConflictingAssets() error {
	if _, err := assets.FS.Open(DataPath); err == nil {
		return fmt.Errorf("%s present in assets directory. %s is a special handler path for dynamic data", DataPath, DataPath)
	}
	return nil
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
