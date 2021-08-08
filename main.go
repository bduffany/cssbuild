package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/bduffany/cssbuild/cssbuild"
)

var (
	inputPath  = flag.String("in", "", "Input file path")
	outputPath = flag.String("out", "", "Output file path")

	jsOutputPath      = flag.String("js_out", "", "JS mapping output path. By default, it will be placed next to the output file, with the same basename as the input path.")
	tsDeclarationPath = flag.String("ts_out", "", "TS declaration output path (*.d.ts). By default, it will be the same as the JS output path, with the \".js\" suffix replaced by \".d.ts\"")
	camelCaseJSKeys   = flag.Bool("camel_case_js_keys", false, "Whether to convert kebab-case class names in the stylesheet to camelCase in the generated JS.")
)

func main() {
	flag.Parse()
	if err := validateFlags(); err != nil {
		io.WriteString(os.Stderr, fmt.Sprintf("%s\n", err))
		flag.Usage()
		os.Exit(1)
	}
	in, err := os.Open(*inputPath)
	if err != nil {
		fatal(err)
	}
	out, err := os.Create(*outputPath)
	if err != nil {
		fatal(err)
	}
	jsPath := *jsOutputPath
	if jsPath == "" {
		jsOutDir := path.Dir(*outputPath)
		inputCSSBase := path.Base(*inputPath)
		jsPath = path.Join(jsOutDir, inputCSSBase+".js")
	}
	js, err := os.Create(jsPath)
	if err != nil {
		fatal(err)
	}
	tsPath := *tsDeclarationPath
	if tsPath == "" {
		tsPath = strings.TrimSuffix(jsPath, ".js") + ".d.ts"
	}
	ts, err := os.Create(tsPath)
	if err != nil {
		fatal(err)
	}
	opts := &cssbuild.TransformOpts{
		JSWriter:            js,
		TSDeclarationWriter: ts,
		CamelCaseJSKeys:     *camelCaseJSKeys,
	}
	if err := cssbuild.Transform(in, out, opts); err != nil {
		fatal(err)
	}
}

func validateFlags() error {
	if *inputPath == "" {
		return fmt.Errorf("missing input path argument")
	}
	if *outputPath == "" {
		return fmt.Errorf("missing --out path")
	}
	return nil
}

func fatal(err error) {
	io.WriteString(os.Stderr, err.Error()+"\n")
	os.Exit(1)
}
