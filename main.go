package main

import (
	"bytes"
	"compress/gzip"
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
	var port int

	flag.StringVar(&fileName, "file", "", "which file to load data from. If not passed, data is read from stdin")
	flag.StringVar(&regexStr, "sep", `\s+`, "which regular expression to use as field separator")
	flag.IntVar(&port, "port", 0, "which port to run the webserver at. If not defined, the port will be defined by the operating system")
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

	flameGraph := tree.ToFlameGraph()
	flameGraph.Name = "TOTAL"

	jsonEncodedFlameGraph := bytes.Buffer{}
	jsonEncoder := json.NewEncoder(&jsonEncodedFlameGraph)
	// jsonEncoder.SetIndent("", "  ")
	panicOnErr(jsonEncoder.Encode(flameGraph))

	compressedFlameGraph := bytes.Buffer{}
	compressor := gzip.NewWriter(&compressedFlameGraph)
	_, err = compressor.Write(jsonEncodedFlameGraph.Bytes())
	panicOnErr(err)
	panicOnErr(compressor.Close())

	log.Printf("Generated a tree with %d nodes representing a total of %d samples", tree.Size(), tree.Samples)

	server := NewServer(port, compressedFlameGraph.Bytes())

	// allow GC'ing
	tree = nil
	flameGraph = nil
	jsonEncodedFlameGraph = bytes.Buffer{}
	compressedFlameGraph = bytes.Buffer{}

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
