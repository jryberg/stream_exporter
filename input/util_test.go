package input

import (
	"reflect"
	"strings"
	"testing"
)

func TestReadLinesWithLimit(t *testing.T) {
	// Lines over 4096 bytes exceed bufio.Reader's default buffer and are read in several parts
	long := strings.Repeat("a", 10000)
	cases := []struct {
		name        string
		limit       int
		input       string
		wantLines   []string
		wantSkipped []int
	}{
		{name: "short lines", limit: 10, input: "one\ntwo\n", wantLines: []string{"one", "two"}},
		{name: "empty line kept", limit: 10, input: "one\n\ntwo\n", wantLines: []string{"one", "", "two"}},
		{name: "crlf stripped", limit: 10, input: "one\r\ntwo\r\n", wantLines: []string{"one", "two"}},
		{name: "no trailing newline", limit: 10, input: "one\ntwo", wantLines: []string{"one", "two"}},
		{name: "exactly at limit", limit: 5, input: "12345\n", wantLines: []string{"12345"}},
		{name: "over limit skipped", limit: 5, input: "b\n123456\nb\n", wantLines: []string{"b", "b"}, wantSkipped: []int{6}},
		{name: "multi-part line within limit", limit: 20000, input: long + "\nb\n", wantLines: []string{long, "b"}},
		{name: "multi-part line over limit skipped", limit: 5000, input: "b\n" + long + "\nb\n", wantLines: []string{"b", "b"}, wantSkipped: []int{10000}},
		{name: "over limit without trailing newline", limit: 5000, input: "b\n" + long, wantLines: []string{"b"}, wantSkipped: []int{10000}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ch := make(chan string, len(c.wantLines)+1)
			var skipped []int
			err := readLinesWithLimit(strings.NewReader(c.input), ch, c.limit, func(size int) {
				skipped = append(skipped, size)
			})
			if err != nil {
				t.Fatal(err)
			}
			close(ch)

			var lines []string
			for l := range ch {
				lines = append(lines, l)
			}
			if !reflect.DeepEqual(c.wantLines, lines) {
				t.Fatalf("Mismatch! Expected lines\n\t%q\nGot\n\t%q", c.wantLines, lines)
			}
			if !reflect.DeepEqual(c.wantSkipped, skipped) {
				t.Fatalf("Mismatch! Expected skipped sizes %v, got %v", c.wantSkipped, skipped)
			}
		})
	}
}
