package fresher

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Option struct {
	buildPath    string
	paths        []string
	excludePaths []string
	exts         []string
	ignoreTest   bool
	interval     time.Duration
}

func defaultOption() *Option {
	return &Option{
		buildPath: "main.go",
		paths:     []string{"."},
		interval:  time.Second * 3,
	}
}

type Fresher struct {
	opt         *Option
	event       *fsnotify.Event
	latestEvent *fsnotify.Event
}

func New(fns ...OptionFunc) *Fresher {
	fr := &Fresher{
		opt: defaultOption(),
	}
	for _, fn := range fns {
		fn(fr)
	}
	return fr
}

func (f *Fresher) Watch() error {
	log.Println("Start Watching......")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to init watcher: %w", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go f.publish(watcher)
	go f.subscribe()

	for _, path := range f.opt.paths {
		if err := watcher.Add(path); err != nil {
			return fmt.Errorf("failed to add file or dir: %w", err)
		}
	}
	<-done
	return nil
}

func (f *Fresher) publish(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			fmt.Println(event)
			if !ok {
				return
			}
			if f.shouldBuild(event.Name) && event.Op&fsnotify.Chmod != fsnotify.Chmod {
				f.event = &event
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func (f *Fresher) shouldBuild(filename string) bool {
	if !f.includePath(filename) {
		return false
	}
	if f.excludePath(filename) {
		return false
	}
	if !f.includeExt(filename) {
		return false
	}
	if f.opt.ignoreTest && strings.Contains(filename, "_test.go") {
		return false
	}
	return true
}

func (f *Fresher) includePath(filename string) bool {
	dir := filepath.Dir(filename)
	for _, p := range f.opt.paths {
		if p == dir {
			return true
		}
	}
	return false
}

func (f *Fresher) excludePath(filename string) bool {
	dir := filepath.Dir(filename)
	for _, p := range f.opt.excludePaths {
		if p == dir {
			return true
		}
	}
	return false
}

func (f *Fresher) includeExt(filename string) bool {
	ext := filepath.Ext(filename)
	for _, p := range f.opt.exts {
		if fmt.Sprintf(".%s", p) == ext {
			return true
		}
	}
	return false
}

func (f *Fresher) build() error {
	event := f.event
	defer func() {
		time.Sleep(f.opt.interval)
	}()
	if event == nil {
		return nil
	}
	if event == f.latestEvent {
		return nil
	}

	cmd := exec.Command("go", "run", f.opt.buildPath)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := io.Copy(os.Stdout, stdout); err != nil {
		return err
	}

	errBuf, err := ioutil.ReadAll(stderr)
	if err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		log.Println(string(errBuf))
		return err
	}
	f.latestEvent = event
	return nil
}

func (f *Fresher) subscribe() {
	for {
		if err := f.build(); err != nil {
			log.Println(err)
		}
	}
}
