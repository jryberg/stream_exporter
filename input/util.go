package input

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"
)

var (
	maxLineSize = flag.Int("input.max-line-size", 1024*1024, "Maximum length in bytes of a line read from stdin, a socket, a named pipe or a file in dryrun mode. Longer lines are skipped")

	SkippedLines = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "stream_exporter",
			Subsystem: "line_processing",
			Name:      "lines_skipped_total",
			Help:      "Number of lines skipped because they were longer than -input.max-line-size",
		},
	)
)

// LinesProcessed is received from by the main loop only between lines. Since the line channel is
// unbuffered, sending a line only means the main loop got it, not that its metrics are updated, but a
// send on LinesProcessed completes only once every line sent before it has been fully processed
var LinesProcessed = make(chan struct{})

// waitForProcessing blocks until all lines sent so far are reflected in the metrics
func waitForProcessing() {
	LinesProcessed <- struct{}{}
}

// readLines sends each line read from r to ch until EOF, skipping lines longer than -input.max-line-size
func readLines(r io.Reader, ch chan<- string, source string) error {
	limit := *maxLineSize
	return readLinesWithLimit(r, ch, limit, func(size int) {
		log.Warnf("Skipped line of %d bytes from %s, longer than -input.max-line-size %d", size, source, limit)
		SkippedLines.Inc()
	})
}

// readLinesWithLimit never buffers more than limit bytes of a line, so a line without newlines can't exhaust memory
func readLinesWithLimit(r io.Reader, ch chan<- string, limit int, onSkip func(size int)) error {
	reader := bufio.NewReader(r)
	line := make([]byte, 0, min(limit, 64*1024))
	size := 0
	for {
		// ReadLine strips the line ending, and returns long lines in several parts with isPrefix set
		part, isPrefix, err := reader.ReadLine()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		size += len(part)
		if size <= limit {
			line = append(line, part...)
		}
		if isPrefix {
			continue
		}

		if size <= limit {
			ch <- string(line)
		} else {
			onSkip(size)
		}
		line = line[:0]
		size = 0
	}
}

func writeMetrics(w io.Writer) {
	metfam, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		log.Fatal(err)
	}
	out := &bytes.Buffer{}
	for _, met := range metfam {
		if _, err := expfmt.MetricFamilyToText(out, met); err != nil {
			log.Fatal(err)
		}
	}
	if _, err := fmt.Fprintln(w, out); err != nil {
		log.Fatal(err)
	}
}
