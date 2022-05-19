package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/rivo/tview"
	"github.com/valyala/fastjson"
)

// defines the max line size in bytes.
const defaultReaderSize = 1_000_000

// defines stdOut initial buffer size.
const defaultBufferSize = 1_000_000

type parserHandler func(line []byte) (fields []field)

type State struct {
	ctx        context.Context
	buf        []byte
	bufSize    int
	lines      lines
	lineCount  int
	readChan   chan []byte
	parser     parserHandler
	lastFields []field
	i          indexer
	cmd        *exec.Cmd
	reader     *bufio.Reader
}

var jsonParser = fastjson.Parser{}

func jsonHandler(line []byte) (fields []field) {
	v, _ := jsonParser.ParseBytes(line)
	obj := v.GetObject()
	obj.Visit(func(key []byte, v *fastjson.Value) {
		fields = append(fields, field{
			key:   key,
			value: v.GetStringBytes(),
		})
	})
	return
}

func textHandler(line []byte) (fields []field) {
	lineLen := len(line)
	cur := 0
	isKey := true
	field := field{}

	for i := 0; i < lineLen; i++ {
		switch line[i] {
		case ' ':
			if isKey {
				field.key = line[cur:i]
			} else {
				field.value = line[cur:i]
			}
			isKey = true
			cur = i + 1
			fields = append(fields, field)
		case '=':
			if isKey {
				isKey = false
				field.key = line[cur:i]
				cur = i + 1
			}
		}
	}
	return
}

func defineParseStrategy(line []byte) parserHandler {
	switch line[0] {
	case '{':
		return jsonHandler
	default:
		return textHandler
	}
}

func (s *State) close() {
	close(s.readChan)
	s.cmd.Process.Kill()
}

func (s *State) start() {
	msg, _, err := s.reader.ReadLine()
	switch {
	case err == nil:
		msg = append(msg, '\n')
	case errors.Is(err, io.EOF):
		return
	default:
		msg = []byte(fmt.Sprintf("failed to read: %v", err))
	}
	s.parser = defineParseStrategy(msg)
	for {
		// check for context cancelation.
		if s.ctx.Err() != nil {
			return
		}
		size := len(msg)
		// parsing.
		fields := s.parser(msg)
		s.lastFields = fields
		// index fields.
		s.i.Index(fields, s.bufSize, size)
		// index last lines.
		s.lines = append(s.lines, line{s.bufSize, s.bufSize + size})
		// update flags.
		s.lineCount++
		s.bufSize += size
		// append to stdOut buffer.
		s.buf = append(s.buf, msg...)
		s.readChan <- msg
		// read next message.
		msg, _, err = s.reader.ReadLine()
		switch {
		case err == nil:
			msg = append(msg, '\n')
		case errors.Is(err, io.EOF):
			s.close()
			return
		default:
			msg = []byte(fmt.Sprintf("failed to read: %v", err))
		}
	}
}

// var out = make([]byte, 0, 1_000_000)

// func (s *State) printLines(l []line) []byte {
// 	out = out[:0] // clear buffer
// 	len := len(l)
// 	for i := 0; i < len; i++ { // show last line first
// 		line := &l[i]
// 		out = append(out, s.buf[line.startOffset:line.endOffset]...)
// 	}
// 	return out
// }

func (s *State) RenderStats() string {
	return fmt.Sprintf(`Buffer Size: %s/%s
Line Count: %d`,
		ByteCountSI(s.bufSize),
		ByteCountSI(cap(s.buf)),
		s.lineCount,
	)
}

func initClog(ctx context.Context, args []string) *State {
	cmd := exec.Command(args[0], args[1:]...)
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil
	}
	reader := bufio.NewReaderSize(pipe, defaultReaderSize)
	if err := cmd.Start(); err != nil {
		return nil
	}
	s := &State{
		ctx:      ctx,
		cmd:      cmd,
		reader:   reader,
		buf:      make([]byte, 0, defaultBufferSize),
		lines:    make(lines, 0, 100_000),
		readChan: make(chan []byte),
		i: &mapIndexer{
			keys: make(map[string]node, 1000), // initial capacity for 1k fields.
		},
	}
	go s.start()
	return s
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	args := os.Args[1:]
	clog := initClog(ctx, args)

	app := tview.NewApplication()
	grid := tview.NewGrid()
	grid.SetColumns(0).SetRows(3, 0)

	statsBox := tview.NewTextView()
	outputBox := tview.NewTextView().
		SetWordWrap(true).
		SetScrollable(true).
		ScrollToEnd().SetChangedFunc(func() {
		statsBox.SetText(clog.RenderStats())
		app.Draw()
	})
	grid.AddItem(statsBox, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(outputBox, 1, 0, 1, 1, 0, 0, true)

	go func() {
		for msg := range clog.readChan {
			fmt.Fprint(outputBox, string(msg))
		}
	}()

	if err := app.SetRoot(grid, true).Run(); err != nil {
		cancel()
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
}
