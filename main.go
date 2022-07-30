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

	jsModuleName      = flag.String("js_module_name", "", "JS module name. Required.")
	jsOutputPath      = flag.String("js_out", "", "JS mapping output path. By default, it will be placed next to the output file, with the same basename as the input path.")
	tsDeclarationPath = flag.String("ts_declaration_out", "", "TS declaration output path (*.d.ts). By default, it will be the same as the JS output path, with the \".js\" suffix replaced by \".d.ts\"")
	tsPath            = flag.String("ts_out", "", "TS mapping output path.")
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
	var js, tsd, ts io.WriteCloser

	if *tsPath == "" {
		jsPath := *jsOutputPath
		if jsPath == "" {
			jsOutDir := path.Dir(*outputPath)
			inputCSSBase := path.Base(*inputPath)
			jsPath = path.Join(jsOutDir, inputCSSBase+".js")
		}
		js, err = os.Create(jsPath)
		if err != nil {
			fatal(err)
		}
		defer js.Close()

		tsDeclarationPath := *tsDeclarationPath
		if tsDeclarationPath == "" {
			tsDeclarationPath = strings.TrimSuffix(jsPath, ".js") + ".d.ts"
		}
		tsd, err = os.Create(tsDeclarationPath)
		if err != nil {
			fatal(err)
		}
		defer tsd.Close()
	} else {
		ts, err = os.Create(*tsPath)
		if err != nil {
			fatal(err)
		}
		defer ts.Close()
	}

	opts := &cssbuild.TransformOpts{
		JSWriter:            js,
		TSDeclarationWriter: tsd,
		TSWriter:            ts,
		JSModuleName:        *jsModuleName,
		CamelCaseJSKeys:     *camelCaseJSKeys,
	}
	if err := cssbuild.Transform(in, out, opts); err != nil {
		fatal(err)
	}
}

func validateFlags() error {
	if *inputPath == "" {
		return fmt.Errorf("missing input CSS module path (`-in` flag)")
	}
	if *outputPath == "" {
		return fmt.Errorf("missing output CSS path (`-out` flag)")
	}
	if *jsModuleName == "" && *tsPath == "" {
		return fmt.Errorf("missing JS module name (`-js_module_name` flag)")
	}
	if *tsDeclarationPath != "" && *tsPath != "" {
		return fmt.Errorf("cannot specify both `-ts_declaration_out` flag and `-ts_out` flag")
	}
	if *jsOutputPath != "" && *tsPath != "" {
		return fmt.Errorf("cannot specify both `-js_out` flag and `-ts_out` flag")
	}
	return nil
}

func fatal(err error) {
	io.WriteString(os.Stderr, "fatal: "+err.Error()+"\n")
	os.Exit(1)
}
