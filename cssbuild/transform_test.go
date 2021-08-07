package cssbuild

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestTransform(t *testing.T) {
	input := readFileAsString(t, "testdata/input.css")
	expected := readFileAsString(t, "testdata/expected_output.module.css")
	var actual bytes.Buffer

	err := Transform(strings.NewReader(input), &actual, &TransformOpts{Suffix: []byte("__SUFFIX__")})

	checkErr(t, err)
	checkDiff(t, expected, formatCSS(t, actual.String()))
}

func readFileAsString(t *testing.T, path string) string {
	f, err := os.Open(path)
	checkErr(t, err)
	defer f.Close()
	b, err := io.ReadAll(f)
	checkErr(t, err)
	return string(b)
}

func checkDiff(t *testing.T, left, right string) {
	leftPath := writeTmp(t, left)
	rightPath := writeTmp(t, right)
	b, err := exec.Command("bash", "-c", fmt.Sprintf(`
		set -euo pipefail
		diff -Pdpru %q %q | colordiff | tail -n +3
	`, leftPath, rightPath)).CombinedOutput()
	if err, ok := err.(*exec.ExitError); ok {
		if err.ExitCode() == 0 {
			// No diff!
			return
		}
		t.Fatalf("unexpected diff:\n%s", string(b))
	}
	checkErr(t, err)
}

func formatCSS(t *testing.T, css string) string {
	cmd := exec.Command("prettier", "--parser", "css")
	cmd.Stdin = strings.NewReader(css)
	b, err := cmd.CombinedOutput()
	checkErr(t, err)
	return string(b)
}

func writeTmp(t *testing.T, content string) string {
	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(f.Name()); err != nil {
			t.Fatalf("failed to clean up temp file: %s", err)
		}
	})
	io.WriteString(f, content)
	return f.Name()
}

func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}