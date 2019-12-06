package fresher

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Option struct {
	command       *Command
	paths         []*RecursiveDir
	globalExclude *GlobalExclude
	exts          Extentions
	interval      time.Duration
}

func defaultOption() *Option {
	return &Option{
		command: &Command{
			Name: "go",
			Args: []string{"version"},
		},
		paths: []*RecursiveDir{
			{
				Name: ".",
			},
		},
		exts:     Extentions{"go"},
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

	if err := f.buildCMD(); err != nil {
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
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

func (f *Fresher) buildCMD() error {
	cmd := exec.Command(f.opt.command.Name, f.opt.command.Args...)
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
	fmt.Println("RUN BUILD CMD")
	if err := f.buildCMD(); err != nil {
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
