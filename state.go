package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

const showLineCount = 10

// defines the max line size in bytes.
const defaultReaderSize = 1_000_000

// defines stdOut initial buffer size.
const defaultBufferSize = 1_000_000

type parserHandler func(line []byte) (fields []field)

type UIController struct {
	scrollIndex   int
	Width, Height int
}

func (c *UIController) Scroll(delta, scrollableHeight int) {
	if c.scrollIndex == -1 && delta == 1 {
		c.scrollIndex = scrollableHeight - showLineCount - 2
		if c.scrollIndex < 0 {
			c.scrollIndex = 0
		}
		return
	}
	c.scrollIndex -= delta
	if c.scrollIndex < 0 {
		c.scrollIndex = -1
	}
	if c.scrollIndex > scrollableHeight-showLineCount {
		c.scrollIndex = -1
	}
}

func (c *UIController) ScrollBottom() {
	c.scrollIndex = -1
}

type State struct {
	ctx        context.Context
	buf        []byte
	bufSize    int
	lines      lines
	lineCount  int
	parser     parserHandler
	lastFields []field
	i          indexer
	cmd        *exec.Cmd
	reader     *bufio.Reader
	ui         UIController
}

func (s *State) ReadCmd() tea.Msg {
	// read next message.
	msg, _, err := s.reader.ReadLine()
	switch {
	case err == nil:
		msg = append(msg, '\n')
	case errors.Is(err, io.EOF):
		s.close()
		return tea.Quit
	default:
		msg = []byte(fmt.Sprintf("failed to read: %v", err))
	}
	return msg
}

func (s *State) Init() tea.Cmd {
	msg, _, err := s.reader.ReadLine()
	switch {
	case err == nil:
		msg = append(msg, '\n')
	case errors.Is(err, io.EOF):
		return nil
	default:
		msg = []byte(fmt.Sprintf("failed to read: %v", err))
	}
	s.parser = defineParseStrategy(msg)
	return func() tea.Msg {
		return msg
	}
}

func (s *State) View() string {
	var lines lines
	var start int
	if s.ui.scrollIndex > -1 {
		start = s.ui.scrollIndex
		if start < 0 {
			start = 0
		}
		end := start + showLineCount
		if end > s.lineCount {
			end = s.lineCount
		}
		lines = s.lines[start:end]
	} else {
		start = s.lineCount - showLineCount
		if start < 0 {
			start = 0
		}
		lines = s.lines[start:]
	}
	return fmt.Sprintf(`%s
Scroll: %d to %d of %d`,
		string(s.printLines(lines)),
		start, start+min(showLineCount, s.lineCount), s.lineCount,
	)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *State) close() {
	s.cmd.Process.Kill()
}

func (s *State) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case []byte:
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
		return s, s.ReadCmd
	case tea.WindowSizeMsg:
		s.ui.Width, s.ui.Height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		// The "up" and "k" keys move the cursor up
		case "up", "k":
			s.ui.Scroll(1, s.lineCount)
		// The "down" and "j" keys move the cursor down
		case "down", "j":
			s.ui.Scroll(-1, s.lineCount)
		case "ctrl+c", "q":
			return s, tea.Quit
		case "esc":
			s.ui.ScrollBottom()
		}
	}
	return s, nil
}

var out = make([]byte, 0, 1_000_000)

func (s *State) printLines(l []line) []byte {
	out = out[:0] // clear buffer
	len := len(l)
	for i := 0; i < len; i++ { // show last line first
		line := &l[i]
		end := line.endOffset
		size := line.endOffset - line.startOffset
		if size > s.ui.Width {
			end -= size - s.ui.Width - 1
		}
		out = append(out, s.buf[line.startOffset:end]...)
		if size > s.ui.Width {
			out = append(out, '\n')
		}
	}
	return out
}

func (s *State) renderStats() string {
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
	return &State{
		ctx:    ctx,
		cmd:    cmd,
		reader: reader,
		buf:    make([]byte, 0, defaultBufferSize),
		lines:  make(lines, 0, 100_000),
		ui: UIController{
			scrollIndex: -1,
		},
		i: &mapIndexer{
			keys: make(map[string]node, 1000), // initial capacity for 1k fields.
		},
	}
}
