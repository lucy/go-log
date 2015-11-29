package log

import (
	"fmt"
	"io"
	"runtime"
	"sync"
	"time"
)

// Level is a log level.
type Level int

// Available log levels.
const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
)

// LevelStrings are the strings prefixed to each log message based on level.
type LevelStrings [5]string

// DefaultLevelStrings are the default level strings.
var DefaultLevelStrings = LevelStrings{
	"DEBUG",
	"INFO ",
	"WARN ",
	"ERROR",
}

// Flags represents options for the logger.
type Flags int

// Logger options.
const (
	// FlagLongPath prepends the full source file path.
	FlagLongPath = 1 << iota
	// FlagShortPath prepends a shortened source file path.
	FlagShortPath
)

// A Logger is a thread safe logger with level indicators.
type Logger struct {
	sync.Mutex
	out  io.Writer
	buf  []byte
	min  Level
	pre  LevelStrings
	flag Flags
}

// New creates a new logger.
// If pre is nil it uses the default level strings.
func New(out io.Writer, minLevel Level, flags Flags, pre *LevelStrings) *Logger {
	if pre == nil {
		pre = &DefaultLevelStrings
	}
	return &Logger{out: out, min: minLevel, flag: flags, pre: *pre}
}

func itoa(buf *[]byte, i int, wid int) {
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

func (log *Logger) date(now time.Time) {
	hour, minute, second := now.Clock()
	year, month, day := now.Date()
	//nsec := now.Nanosecond()
	itoa(&log.buf, year, 4)
	log.buf = append(log.buf, '-')
	itoa(&log.buf, int(month), 2)
	log.buf = append(log.buf, '-')
	itoa(&log.buf, day, 2)
	log.buf = append(log.buf, 'T')
	itoa(&log.buf, hour, 2)
	log.buf = append(log.buf, ':')
	itoa(&log.buf, minute, 2)
	log.buf = append(log.buf, ':')
	itoa(&log.buf, second, 2)
	//log.buf = append(log.buf, '.')
	//itoa(&log.buf, nsec, 9)
	_, off := now.Zone()
	if off == 0 {
		log.buf = append(log.buf, 'Z')
	} else {
		zone := off / 60
		absoff := off
		if zone < 0 {
			log.buf = append(log.buf, '-')
			absoff = -absoff
			zone = -zone
		} else {
			log.buf = append(log.buf, '+')
		}
		itoa(&log.buf, zone/60, 2)
		log.buf = append(log.buf, ':')
		itoa(&log.buf, zone%60, 2)
	}
}

func (log *Logger) header(level Level, now time.Time, file string, line int) {
	log.buf = append(log.buf, log.pre[int(level)]...)
	log.buf = append(log.buf, ' ')
	//2006-01-02T15:04:05.999999999Z07:00
	log.date(now)
	log.buf = append(log.buf, ' ')
	if log.flag&(FlagShortPath|FlagLongPath) != 0 {
		if log.flag&(FlagShortPath) != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		log.buf = append(log.buf, file...)
		log.buf = append(log.buf, ':')
		itoa(&log.buf, line, -1)
		log.buf = append(log.buf, ": "...)
	}
}

// Output is the generic printing function.
func (log *Logger) Output(l Level, s string) error {
	now := time.Now()
	log.Lock()
	defer log.Unlock()
	if l < log.min {
		return nil
	}
	var file string
	var line int
	if log.flag&(FlagShortPath|FlagLongPath) != 0 {
		log.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(2)
		if !ok {
			file = "?"
			line = 0
		}
		log.Lock()
	}
	log.buf = log.buf[0:0]
	log.header(l, now, file, line)
	log.buf = append(log.buf, s...)
	log.buf = append(log.buf, '\n')
	_, err := log.out.Write(log.buf)
	return err
}

// Log outputs a log message at the specified level.
func (log *Logger) Log(l Level, v ...interface{}) {
	log.Output(l, fmt.Sprint(v...))
}

// Logf outputs a formatted log message at the specified level.
func (log *Logger) Logf(l Level, format string, v ...interface{}) {
	log.Output(l, fmt.Sprintf(format, v...))
}

// Debug is Log at the debug log level.
func (log *Logger) Debug(v ...interface{}) {
	log.Output(LevelDebug, fmt.Sprint(v...))
}

// Debugf is Log at the debug log level.
func (log *Logger) Debugf(format string, v ...interface{}) {
	log.Output(LevelDebug, fmt.Sprintf(format, v...))
}

// Info is Log at the info log level.
func (log *Logger) Info(v ...interface{}) {
	log.Output(LevelInfo, fmt.Sprint(v...))
}

// Infof is Log at the info log level.
func (log *Logger) Infof(format string, v ...interface{}) {
	log.Output(LevelInfo, fmt.Sprintf(format, v...))
}

// Warn is Log at the warn log level.
func (log *Logger) Warn(v ...interface{}) {
	log.Output(LevelWarn, fmt.Sprint(v...))
}

// Warnf is Log at the warn log level.
func (log *Logger) Warnf(format string, v ...interface{}) {
	log.Output(LevelWarn, fmt.Sprintf(format, v...))
}

// Error is Log at the error log level.
func (log *Logger) Error(v ...interface{}) {
	log.Output(LevelError, fmt.Sprint(v...))
}

// Errorf is Log at the error log level.
func (log *Logger) Errorf(format string, v ...interface{}) {
	log.Output(LevelError, fmt.Sprintf(format, v...))
}
