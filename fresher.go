package fresher

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Fresher struct {
	event       *fsnotify.Event
	latestEvent *fsnotify.Event
}

func New(opt ...func() interface{}) *Fresher {
	return &Fresher{}
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

	err = watcher.Add("./testdata")
	if err != nil {
		return fmt.Errorf("failed to add file or dir: %w", err)
	}
	<-done
	return nil
}

func (f *Fresher) publish(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
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

func (f *Fresher) build() error {
	event := f.event
	defer func() {
		time.Sleep(3 * time.Second)
	}()
	if event == nil {
		return nil
	}
	if event == f.latestEvent {
		return nil
	}

	cmd := exec.Command("go", "run", "./testdata/a.go")
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
