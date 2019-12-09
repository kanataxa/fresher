package fresher

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Option struct {
	target        string
	paths         []*WatcherPath
	globalExclude *GlobalExclude
	exts          Extensions
	interval      time.Duration
}

func defaultOption() *Option {
	return &Option{
		target: "main.go",
		paths: []*WatcherPath{
			{
				Name: ".",
			},
		},
		exts:     Extensions{"go"},
		interval: time.Second * 3,
	}
}

type Fresher struct {
	opt         *Option
	event       *fsnotify.Event
	latestEvent *fsnotify.Event
	rebuild     chan bool
}

func New(fns ...OptionFunc) *Fresher {
	fr := &Fresher{
		opt:     defaultOption(),
		rebuild: make(chan bool, 1),
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
	defer close(done)

	if err := f.run(); err != nil {
		log.Println("failed to build: %w", err)
	}

	go f.publish(watcher)
	go f.subscribe()

	for _, path := range f.opt.paths {
		if err := path.Walk(watcher, f.opt); err != nil {
			return err
		}
	}

	<-done
	fmt.Println("DONE")
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
			f.event = &event
		case err, ok := <-watcher.Errors:
			fmt.Println("ERRR", err)
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func (f *Fresher) buildTarget() error {
	cmd := exec.Command("go", "build", "-o", filepath.Join(os.TempDir(), "fresher_run"), f.opt.target)
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	log.Println(string(output))
	return nil
}

func (f *Fresher) run() error {
	if err := f.buildTarget(); err != nil {
		return err
	}
	cmd := exec.Command(filepath.Join(os.TempDir(), "fresher_run"))
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	fmt.Println("PROCESS START")
	if err := cmd.Start(); err != nil {
		return err
	}

	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	go func() {
		<-f.rebuild
		fmt.Println("KILL PROCESS")
		cmd.Process.Kill()
	}()

	return nil
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
	defer func() {
		f.latestEvent = event
	}()
	f.rebuild <- true
	if err := f.run(); err != nil {
		return err
	}
	return nil
}

func (f *Fresher) subscribe() {
	for {
		if err := f.build(); err != nil {
			log.Println(err)
		}
	}
}
