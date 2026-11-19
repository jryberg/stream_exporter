package input

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
)

var (
	writeOnEOF = flag.Bool("input.stdin.write-on-eof", false, "If all metrics should be written to stdout after seeing end of file")
	quitOnEOF  = flag.Bool("input.stdin.quit-on-eof", false, "If the exporter should exit after seeing end of file")
)

type StdinInput struct {
}

func init() {
	registerInput("stdin", newStdinInput)
}

func newStdinInput() (StreamInput, error) {
	return StdinInput{}, nil
}

func (input StdinInput) StartStream(ch chan<- string) {
	if err := readLines(os.Stdin, ch, "stdin"); err != nil {
		log.Errorf("Error reading stdin: %v", err)
	}

	if *writeOnEOF {
		waitForProcessing()
		writeMetrics(os.Stdout)
	}
	if *quitOnEOF {
		close(ch)
	}
}
