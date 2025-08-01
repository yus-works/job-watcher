package logging

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/sirupsen/logrus"
)

const DefaultTimeFormat = "06-01-02 15:04:05"

// fileHook writes entries to a writer with its own formatter (no colors)
type fileHook struct {
	w         io.Writer
	formatter logrus.Formatter
	levels    []logrus.Level
}

func (h *fileHook) Levels() []logrus.Level { return h.levels }
func (h *fileHook) Fire(e *logrus.Entry) error {
	b, err := h.formatter.Format(e)
	if err != nil {
		return err
	}
	_, err = h.w.Write(b)
	return err
}

const (
	Reset = "\x1b[0m"

	FGBlack       = "\x1b[30m"
	FGWhite       = "\x1b[37m"
	FGWhiteBright = "\x1b[97m"

	BGRed      = "\x1b[41m"
	BGYellow   = "\x1b[43m"
	BGGreen    = "\x1b[42m"
	BGBlue     = "\x1b[44m"
	BGCyan     = "\x1b[46m"
	BGMagenta  = "\x1b[45m"
	BGGrayDark = "\x1b[48;5;236m"
)

type BlockFormatter struct {
	TimeFormat string
	WithColor  bool

	FieldBG   string
	FieldFG   string
	PadInside bool

	LevelBG string
	LevelFG string
	TimeBG  string
	TimeFG  string
}

func levelAutoBG(l logrus.Level) string {
	switch l {
	case logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel:
		return BGRed
	case logrus.WarnLevel:
		return BGYellow
	case logrus.InfoLevel:
		return BGGreen
	case logrus.DebugLevel:
		return BGBlue
	case logrus.TraceLevel:
		return BGGrayDark
	default:
		return BGGrayDark
	}
}

func contrastFG(bg string) string {
	switch bg {
	case BGGrayDark:
		return FGWhite
	case BGGreen, BGBlue, BGRed, BGYellow:
		return FGBlack
	default:
		return FGBlack
	}
}

func (f *BlockFormatter) block(b *bytes.Buffer, bg, fg, text string) {
	if f.PadInside {
		text = " " + text + " "
	}
	fmt.Fprint(b, bg, fg, text, Reset)
}

func (f *BlockFormatter) Format(e *logrus.Entry) ([]byte, error) {
	var b bytes.Buffer

	lvl := strings.ToUpper(e.Level.String())
	if e.Level == logrus.WarnLevel {
		lvl = "WARN"
	}

	ts := e.Time.Format(f.TimeFormat)

	src := ""
	if e.HasCaller() && e.Caller != nil {
		src = fmt.Sprintf("%s:%d", filepath.Base(e.Caller.File), e.Caller.Line)
	}

	if f.WithColor {
		// LEVEL block
		lbg := f.LevelBG
		if lbg == "" {
			lbg = levelAutoBG(e.Level)
		}
		lfg := f.LevelFG
		if lfg == "" {
			lfg = contrastFG(lbg)
		}
		f.block(&b, lbg, lfg, fmt.Sprintf("%-5s", lvl))
		b.WriteByte(' ')

		// TIME block
		tbg := f.TimeBG
		tfg := f.TimeFG
		if tbg == "" {
			tbg = BGGrayDark
		}
		if tfg == "" {
			tfg = FGWhiteBright
		}
		f.block(&b, tbg, tfg, ts)
	} else {
		fmt.Fprintf(&b, "%-5s [%s]", lvl, ts)
	}
	// source and message
	if src != "" {
		fmt.Fprintf(&b, " %s", src)
	}
	if e.Message != "" {
		fmt.Fprintf(&b, " %s", e.Message)
	}

	// Fields as blocks (or plain when WithColor=false)
	if len(e.Data) > 0 {
		keys := make([]string, 0, len(e.Data))
		for k := range e.Data {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			val := e.Data[k]
			if f.WithColor && (f.FieldBG != "" || f.FieldFG != "") {
				b.WriteByte(' ')
				bg, fg := f.FieldBG, f.FieldFG
				if bg == "" {
					bg = BGGrayDark
				}
				if fg == "" {
					fg = FGWhiteBright
				}
				f.block(&b, bg, fg, fmt.Sprintf("%s=%v", k, val))
			} else {
				fmt.Fprintf(&b, " %s=%v", k, val)
			}
		}
	}
	b.WriteByte('\n')
	return b.Bytes(), nil
}

// New builds a logger that prints colorized logs to stdout and
// simultaneously writes plain-text logs to `logPath`.
// Return values: logger, closer (for the file), error.

// New builds a logger that prints colorized logs to stdout and
// simultaneously writes plain-text logs to `logPath`.

func New(level logrus.Level, logPath string) (*logrus.Logger, io.Closer, error) {
	baseOut := os.Stdout
	var console io.Writer = baseOut
	if runtime.GOOS == "windows" {
		console = colorable.NewColorableStdout()
	}
	useColor := isatty.IsTerminal(baseOut.Fd()) || isatty.IsCygwinTerminal(baseOut.Fd())

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, err
	}

	log := logrus.New()
	log.SetOutput(console)
	log.SetReportCaller(true)
	log.SetLevel(level)

	// Console: level/time/fields as background blocks
	log.SetFormatter(&BlockFormatter{
		TimeFormat: DefaultTimeFormat,
		WithColor:  useColor,

		// Per-level background is automatic if LevelBG is empty.
		LevelBG: "", // auto: error=red, warn=yellow, info=green, debug=blue, trace=gray
		LevelFG: "", // auto-contrast

		// Time block style (neutral)
		TimeBG: BGGrayDark,
		TimeFG: FGWhiteBright,

		// Field blocks
		FieldBG:   BGGrayDark,
		FieldFG:   FGWhiteBright,
		PadInside: true,
	})

	// File: same layout, but no color (plain text)
	log.AddHook(&fileHook{
		w: file,
		formatter: &BlockFormatter{
			TimeFormat: DefaultTimeFormat,
			WithColor:  false,
		},
		levels: logrus.AllLevels,
	})

	return log, file, nil
}
