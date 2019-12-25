package fresher

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Option struct {
	build         *BuildConfig
	configs       []*WatcherConfig
	globalExclude *GlobalExclude
	exts          Extensions
	interval      time.Duration
}

func defaultOption() *Option {
	return &Option{
		build: &BuildConfig{
			Target: "main.go",
		},
		configs: []*WatcherConfig{
			{
				Name: ".",
			},
		},
		exts:     Extensions{"go"},
		interval: time.Second * 3,
	}
}

type Fresher struct {
	opt     *Option
	event   chan fsnotify.Event
	rebuild chan bool
	timer   *time.Timer
}

func New(fns ...OptionFunc) *Fresher {
	fr := &Fresher{
		opt:     defaultOption(),
		event:   make(chan fsnotify.Event, 1),
		rebuild: make(chan bool, 1),
	}
	for _, fn := range fns {
		fn(fr)
	}
	return fr
}

func (f *Fresher) Watch() error {
	log.Info("Start Watching......")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to init watcher: %v\n", err)
	}
	defer watcher.Close()

	done := make(chan struct{})
	go func() {
		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt)
		<-quit
		f.rebuild <- true
		close(quit)
		close(done)
	}()

	if err := f.run(); err != nil {
		log.Error(fmt.Errorf("failed to build: %v\n", err))
	}

	watcherPath := NewWatcherPath(f.opt.configs, f.opt)
	for _, path := range f.opt.configs {
		wp, err := path.Walk(watcher, f.opt)
		if err != nil {
			return err
		}
		watcherPath.Merge(wp)
	}

	go f.publish(watcher, watcherPath)
	go f.subscribe()

	<-done

	return nil
}

func (f *Fresher) publish(watcher *fsnotify.Watcher, watcherPath *WatcherPath) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				continue
			}
			if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				continue
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				if err := watcherPath.AddIfNeeds(event.Name, watcher); err != nil {
					if err != skipToAddErr {
						log.Error(err)
					}
					continue
				}
			}
			if _, exists := watcherPath.watches[event.Name]; !exists {
				continue
			}
			f.event <- event
		case err, ok := <-watcher.Errors:
			if !ok {
				continue
			}
			log.Error(err)
		}
	}
}

func (f *Fresher) run() error {
	log.Building()
	commands := f.opt.build.Commands()
	for _, cmd := range commands{
		if err := cmd.Exec(); err != nil {
			return err
		}
	}
	go func() {
		<-f.rebuild
		for _, cmd := range commands {
			cmd.Kill()
		}
	}()
	return nil
}

func (f *Fresher) reserve() error {
	event := <-f.event
	log.UpdateFile(event.Name)
	if f.timer == nil || f.timer.Stop() {
		f.timer = time.NewTimer(f.opt.interval)
	} else {
		f.timer.Reset(f.opt.interval)
	}
	go func() {
		<-f.timer.C
		f.rebuild <- true
		if err := f.run(); err != nil {
			log.Error(err)
			return
		}
	}()
	return nil
}
func (f *Fresher) subscribe() {
	for {
		if err := f.reserve(); err != nil {
			log.Println(err)
		}
	}
}
