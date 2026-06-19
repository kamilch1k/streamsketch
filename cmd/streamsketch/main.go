package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kamilch1k/streamsketch/internal/csvio"
	"github.com/kamilch1k/streamsketch/internal/sketch"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("streamsketch", flag.ContinueOnError)
	flags.SetOutput(stderr)

	var config sketch.Config
	input := flags.String("input", "", "CSV file with timestamp,stream,user_id,item,count columns")
	output := flags.String("out", "", "optional JSON report path")
	format := flags.String("format", "text", "output format: text or json")
	flags.IntVar(&config.Width, "width", 2048, "count-min sketch width")
	flags.IntVar(&config.Depth, "depth", 5, "count-min sketch depth")
	flags.IntVar(&config.TopK, "top-k", 5, "number of heavy hitters to print")
	precision := flags.Int("precision", 12, "HyperLogLog precision from 4 to 18")

	if err := flags.Parse(args); err != nil {
		return 64
	}
	config.Precision = uint8(*precision)
	if strings.TrimSpace(*input) == "" {
		_, _ = fmt.Fprintln(stderr, "missing required -input path")
		return 64
	}

	file, err := os.Open(*input)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "open input: %v\n", err)
		return 65
	}
	defer file.Close()

	events, err := csvio.ReadEvents(file)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "read input: %v\n", err)
		return 65
	}
	summaries, err := sketch.Analyze(events, config)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "analyze events: %v\n", err)
		return 65
	}

	if strings.TrimSpace(*output) != "" {
		if err := writeJSON(*output, summaries); err != nil {
			_, _ = fmt.Fprintf(stderr, "write report: %v\n", err)
			return 65
		}
	}

	switch strings.ToLower(strings.TrimSpace(*format)) {
	case "json":
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(summaries); err != nil {
			_, _ = fmt.Fprintf(stderr, "write json: %v\n", err)
			return 65
		}
	case "text":
		printText(stdout, summaries)
	default:
		_, _ = fmt.Fprintf(stderr, "unknown -format %q\n", *format)
		return 64
	}

	return 0
}

func writeJSON(path string, summaries []sketch.Summary) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(summaries)
}

func printText(writer io.Writer, summaries []sketch.Summary) {
	for _, summary := range summaries {
		_, _ = fmt.Fprintf(
			writer,
			"stream=%s events=%d estimated_uniques=%.0f\n",
			summary.Stream,
			summary.Events,
			summary.EstimatedUniques,
		)
		for _, item := range summary.TopItems {
			_, _ = fmt.Fprintf(writer, "  %-24s estimate=%d\n", item.Item, item.Estimate)
		}
	}
}
