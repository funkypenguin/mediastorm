package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"google.golang.org/protobuf/proto"
	metapb "novastream/internal/nzb/metadata/proto"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: dumpmeta <file>")
		os.Exit(1)
	}
	data, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	var meta metapb.FileMetadata
	if err := proto.Unmarshal(data, &meta); err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&meta); err != nil {
		panic(err)
	}
}
