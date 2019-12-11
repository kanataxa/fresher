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
	l.Info(l.msg(magenta, fmt.Sprintf("Watching file [%s]", path)))
}

func (l *Log) UpdateFile(path string) {
	l.Info(l.msg(green, fmt.Sprintf("Rebuild to updated watched file [%s]", path)))
}

func (l *Log) IgnoreFile(path string) {
	l.Infof(l.msg(yellow, fmt.Sprintf("Ignore file [%s]", path)))
}

func (l *Log) Building() {
	l.Info(l.msg(yellow, "Building..."))
}

func (l *Log) Info(msg string) {
	l.Logger.Info(l.msg(blue, msg))
}

func (l *Log) Error(v interface{}) {
	l.Info(l.msg(red, fmt.Sprint(v)))
}

func (l *Log) msg(code int, msg string) string {
	return fmt.Sprintf("\033[%dm%s\033[0m", code, fmt.Sprintf("%s: %s", "Fresher Watch", msg))
}

func init() {
	log = &Log{
		Logger: logrus.New(),
	}
	formatter := new(logrus.TextFormatter)
	formatter.ForceColors = true
	log.SetFormatter(formatter)
}
