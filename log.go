package fresher

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type Log struct {
	*logrus.Logger
}

var log *Log

const (
	black = 30 + iota
	red
	green
	yellow
	blue
	magenta
	cyan
)

func (l *Log) WatchFile(path string) {
	l.Infof(l.msg(magenta, fmt.Sprintf("Watching file [%s]", path)))
}

func (l *Log) UpdateFile(path string) {
	l.Infof(l.msg(green, fmt.Sprintf("Rebuild to updated watched file [%s]", path)))
}

func (l *Log) IgnoreFile(path string) {
	l.Infof(l.msg(yellow, fmt.Sprintf("Ignore file [%s]", path)))
}

func (l *Log) Building() {
	l.Infof(l.msg(yellow, "Building..."))
}

func (l *Log) Info(msg string) {
	l.Infof(l.msg(blue, msg))
}

func (l *Log) Error(v interface{}) {
	l.Infof(l.msg(red, fmt.Sprint(v)))
}

func (l *Log) msg(code int, msg string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", code, fmt.Sprintf("%s: %s", "Watcher", msg))
}

func init() {
	log = &Log{
		Logger: logrus.New(),
	}
}
