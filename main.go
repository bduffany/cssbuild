package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/bduffany/cssbuild/cssbuild"
)

var (
	inputPath  = flag.String("in", "", "Input file path")
	outputPath = flag.String("out", "", "Output file path")
)

type Opts struct {
	InputPath  string
	OutputPath string
}

func main() {
	flag.Parse()
	opts, err := getOpts()
	if err != nil {
		io.WriteString(os.Stderr, fmt.Sprintf("%s\n", err))
		flag.Usage()
		os.Exit(1)
	}

	in, err := os.Open(opts.InputPath)
	if err != nil {
		fatal(err)
	}
	out, err := os.Create(opts.OutputPath)
	if err != nil {
		fatal(err)
	}

	if err := cssbuild.Transform(in, out, &cssbuild.TransformOpts{}); err != nil {
		fatal(err)
	}
}

func getOpts() (*Opts, error) {
	if *inputPath == "" {
		return nil, fmt.Errorf("missing input path argument")
	}
	if *outputPath == "" {
		return nil, fmt.Errorf("missing --out path")
	}
	return &Opts{
		InputPath:  *inputPath,
		OutputPath: *outputPath,
	}, nil
}

func fatal(err error) {
	io.WriteString(os.Stderr, err.Error()+"\n")
	os.Exit(1)
}
