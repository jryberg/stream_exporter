package input

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

type DryrunFileInput struct {
	path string
}
type TailingFileInput struct {
	path string
}

var (
	filePath = flag.String("input.file.path", "", "Path to file to read")
	mode     = flag.String("input.file.mode", "tail", "Mode of operation. Valid values are 'tail' and 'dryrun'")
)

func init() {
	registerInput("file", newFileInput)
}

func newFileInput() (StreamInput, error) {
	if *filePath == "" {
		return nil, errors.New("-input.file.path not set")
	}

	switch *mode {
	case "tail":
		return TailingFileInput{
			path: *filePath,
		}, nil
	case "dryrun":
		return DryrunFileInput{
			path: *filePath,
		}, nil
	default:
		return nil, errors.New(fmt.Sprintf("Unknown value %q for -input.file.mode", *mode))
	}
}

func (input DryrunFileInput) StartStream(ch chan<- string) {
	defer close(ch)

	file, err := os.Open(input.path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		ch <- scanner.Text()
	}

	metfam, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		fmt.Println(err)
	}
	out := &bytes.Buffer{}
	for _, met := range metfam {
		if _, err := expfmt.MetricFamilyToText(out, met); err != nil {
			panic(err)
		}
	}
	fmt.Println("Finished reading file, dumping final metrics endpoint output:")
	fmt.Println(out)
}

func (input TailingFileInput) StartStream(ch chan<- string) {
	tailConfig := tail.Config{
		Follow: true,
		ReOpen: true,
	}
	tailer, err := tail.TailFile(input.path, tailConfig)
	if err != nil {
		panic(err)
	}
	defer tailer.Cleanup()

	for line := range tailer.Lines {
		ch <- line.Text
	}
}
