package fresher

import (
	"fmt"
	"os"
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
	log.Info("Start Watching......")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to init watcher: %v\n", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	defer close(done)

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
					log.Error(err)
					continue
				}
			}
			if _, exists := watcherPath.watches[event.Name]; !exists {
				continue
			}
			log.UpdateFile(event.Name)
			f.event = &event
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
	procs := []*os.Process{}
	for _, cmd := range f.opt.build.Commands() {
		proc, err := cmd.Exec()
		if err != nil {
			return err
		}
		if proc == nil {
			continue
		}
		log.Info(fmt.Sprintf("Run Process [%d]", proc.Pid))
		procs = append(procs, proc)
	}
	go func() {
		<-f.rebuild
		for _, proc := range procs {
			log.Info(fmt.Sprintf("Kill Exec Process [%d]", proc.Pid))
			proc.Kill()
		}
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
