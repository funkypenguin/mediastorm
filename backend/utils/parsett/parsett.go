package parsett

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

// ParsedTitle represents the result from PTT's parse_title function
type ParsedTitle struct {
	Title      string   `json:"title"`
	Year       int      `json:"year,omitempty"`
	Resolution string   `json:"resolution,omitempty"`
	Quality    string   `json:"quality,omitempty"`
	Codec      string   `json:"codec,omitempty"`
	Audio      []string `json:"audio,omitempty"`      // Can be array
	Channels   []string `json:"channels,omitempty"`   // Audio channels like 5.1
	Group      string   `json:"group,omitempty"`
	Container  string   `json:"container,omitempty"`
	Episodes   []int    `json:"episodes,omitempty"`
	Seasons    []int    `json:"seasons,omitempty"`
	Languages  []string `json:"languages,omitempty"`
	Extended   bool     `json:"extended,omitempty"`
	Hardcoded  bool     `json:"hardcoded,omitempty"`
	Proper     bool     `json:"proper,omitempty"`
	Repack     bool     `json:"repack,omitempty"`
	Site       string   `json:"site,omitempty"`
	BitDepth   string   `json:"bit_depth,omitempty"`  // e.g., "10bit"
	HDR        []string `json:"hdr,omitempty"`        // HDR formats like DV, HDR, HDR10+
}

// ParseTitle calls the Python PTT library to parse a media title
// Returns a ParsedTitle struct with the parsed information
func ParseTitle(title string) (*ParsedTitle, error) {
	// Get the path to the Python script
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get current file path")
	}

	// The script is at the root of the NovaStream directory
	scriptPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "parse_title.py")
	venvPython := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", ".venv", "bin", "python3")

	// Execute the Python script with the title as an argument
	cmd := exec.Command(venvPython, scriptPath, title)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("python script error: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute python script: %w", err)
	}

	// Parse the JSON output
	var result ParsedTitle
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w (output: %s)", err, string(output))
	}

	return &result, nil
}

// BatchResult represents a single result from batch parsing
type BatchResult struct {
	Title  string       `json:"title"`
	Parsed *ParsedTitle `json:"parsed"`
	Error  string       `json:"error"`
}

// ParseTitleBatch parses multiple titles in a single Python subprocess call
// This is much faster than calling ParseTitle repeatedly (100x speedup for 50 items)
// Returns a map of title -> parsed result
func ParseTitleBatch(titles []string) (map[string]*ParsedTitle, error) {
	if len(titles) == 0 {
		return make(map[string]*ParsedTitle), nil
	}

	// Get the path to the batch Python script
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get current file path")
	}

	// The batch script is at the root of the NovaStream directory
	scriptPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "parse_title_batch.py")
	venvPython := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", ".venv", "bin", "python3")

	// Build command with all titles as arguments
	args := append([]string{scriptPath}, titles...)
	cmd := exec.Command(venvPython, args...)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("python batch script error: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to execute python batch script: %w", err)
	}

	// Parse the JSON array output
	var results []BatchResult
	if err := json.Unmarshal(output, &results); err != nil {
		return nil, fmt.Errorf("failed to parse batch JSON output: %w (output: %s)", err, string(output))
	}

	// Build result map
	resultMap := make(map[string]*ParsedTitle, len(results))
	for _, r := range results {
		if r.Error != "" {
			// Store nil for failed parses (will be treated as parse errors in filter)
			resultMap[r.Title] = nil
		} else {
			resultMap[r.Title] = r.Parsed
		}
	}

	return resultMap, nil
}
