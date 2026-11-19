package input

import (
	"testing"
)

func TestNewNamedPipeInput(t *testing.T) {
	origPipePath, origFilePath := *pipePath, *filePath
	t.Cleanup(func() {
		*pipePath, *filePath = origPipePath, origFilePath
	})

	cases := []struct {
		name     string
		pipePath string
		filePath string
		wantErr  bool
	}{
		{name: "pipe path set", pipePath: "/tmp/stream_exporter.pipe"},
		{name: "only file path set", filePath: "/var/log/some.log", wantErr: true},
		{name: "nothing set", wantErr: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			*pipePath, *filePath = c.pipePath, c.filePath

			in, err := newNamedPipeInput()
			if c.wantErr {
				if err == nil {
					t.Fatalf("Expected error, got input %v", in)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := in.(NamedPipeInput).path; got != c.pipePath {
				t.Fatalf("Mismatch! Expected path %q, got %q", c.pipePath, got)
			}
		})
	}
}
