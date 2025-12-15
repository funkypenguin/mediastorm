package main

import (
	"fmt"
	"log"
	"novastream/utils/parsett"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <title>")
		fmt.Println("Example: go run main.go 'The.Matrix.1999.1080p.BluRay.x264-SPARKS'")
		os.Exit(1)
	}

	title := os.Args[1]

	result, err := parsett.ParseTitle(title)
	if err != nil {
		log.Fatalf("Error parsing title: %v", err)
	}

	// Print the parsed information
	fmt.Printf("Title: %s\n", result.Title)
	if result.Year != 0 {
		fmt.Printf("Year: %d\n", result.Year)
	}
	if result.Resolution != "" {
		fmt.Printf("Resolution: %s\n", result.Resolution)
	}
	if result.Quality != "" {
		fmt.Printf("Quality: %s\n", result.Quality)
	}
	if result.Codec != "" {
		fmt.Printf("Codec: %s\n", result.Codec)
	}
	if len(result.Audio) > 0 {
		fmt.Printf("Audio: %v\n", result.Audio)
	}
	if len(result.Channels) > 0 {
		fmt.Printf("Channels: %v\n", result.Channels)
	}
	if result.Group != "" {
		fmt.Printf("Group: %s\n", result.Group)
	}
	if len(result.Seasons) > 0 {
		fmt.Printf("Seasons: %v\n", result.Seasons)
	}
	if len(result.Episodes) > 0 {
		fmt.Printf("Episodes: %v\n", result.Episodes)
	}
	if result.BitDepth != "" {
		fmt.Printf("Bit Depth: %s\n", result.BitDepth)
	}
}
