package fresher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/goccy/go-yaml"
)

type Extensions []string

func (e Extensions) IsIncludeSameExt(fileName string) bool {
	for _, ext := range e {
		if fmt.Sprintf(".%s", ext) == filepath.Ext(fileName) {
			return true
		}
	}
	return false
}

type GlobalExclude []string

func (g GlobalExclude) IsExclude(path string) (bool, error) {
	for _, d := range g {
		ok, err := filepath.Match(d, path)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

type WatcherConfig struct {
	Name     string   `yaml:"name"`
	Excludes []string `yaml:"exclude"`
	Includes []string `yaml:"include"`
}

func (r *WatcherConfig) UnmarshalYAML(b []byte) error {
	s := struct {
		Name     string   `yaml:"name"`
		Excludes []string `yaml:"exclude"`
		Includes []string `yaml:"include"`
	}{}
	if err := yaml.Unmarshal(b, &s); err != nil {
		var name string
		if err := yaml.Unmarshal(b, &name); err != nil {
			return err
		}
		r.Name = name
		return nil
	}
	r.Name = s.Name
	r.Excludes = s.Excludes
	r.Includes = s.Includes
	return nil
}

func (r *WatcherConfig) IsExclude(path string) (bool, error) {
	for _, f := range r.Excludes {
		ok, err := filepath.Match(f, path)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func (r *WatcherConfig) IsInclude(path string) (bool, error) {
	if len(r.Includes) == 0 {
		return true, nil
	}
	for _, f := range r.Includes {
		ok, err := filepath.Match(f, path)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func (r *WatcherConfig) ShouldWatchFile(path string, opt *Option) (bool, error) {
	shouldWatch, err := r.shouldWatch(path, opt)
	if err != nil {
		return false, err
	}
	if !shouldWatch {
		return false, nil
	}
	if !opt.exts.IsIncludeSameExt(path) {
		return false, nil
	}
	return true, nil
}

func (r *WatcherConfig) shouldWatch(path string, opt *Option) (bool, error) {
	path = filepath.Base(path)
	isGlobalExclude, err := opt.globalExclude.IsExclude(path)
	if err != nil {
		return false, nil
	}
	if isGlobalExclude {
		return false, nil
	}
	isExclude, err := r.IsExclude(path)
	if err != nil {
		return false, nil
	}
	if isExclude {
		return false, nil
	}
	isInclude, err := r.IsInclude(path)
	if err != nil {
		return false, nil
	}
	if !isInclude {
		return false, nil
	}
	return true, nil
}

func (r *WatcherConfig) SkipDir(dirName string, opt *Option) (bool, error) {
	shouldWatch, err := r.shouldWatch(dirName, opt)
	if err != nil {
		return false, err
	}
	return !shouldWatch, nil
}

func (r *WatcherConfig) WalkWithDirName(watcher *fsnotify.Watcher, opt *Option, dirName string) (*WatcherPath, error) {
	watcherPath := NewWatcherPath(opt.configs, opt)
	if _, err := os.Stat(filepath.Join(dirName, r.Name)); err != nil {
		if os.IsNotExist(err) {
			log.IgnoreFile(filepath.Join(dirName, r.Name))
			return watcherPath, nil
		}
	}

	global := opt.globalExclude
	isExclude, err := global.IsExclude(r.Name)
	if err != nil {
		return nil, err
	}
	if isExclude {
		return watcherPath, nil
	}
	if err := filepath.Walk(filepath.Join(dirName, r.Name), func(path string, info os.FileInfo, err error) error {
		if err != nil && err != filepath.SkipDir {
			return err
		}
		if info.IsDir() {
			if info.Name() == r.Name {
				if err := watcher.Add(path); err != nil {
					return err
				}
				return nil
			}
			skipDir, err := r.SkipDir(path, opt)
			if err != nil {
				return err
			}
			if !skipDir {
				if err := watcher.Add(path); err != nil {
					return err
				}
				return nil
			}
			return filepath.SkipDir
		}

		shouldWatch, err := r.ShouldWatchFile(path, opt)
		if err != nil {
			return err
		}
		if !shouldWatch {
			log.IgnoreFile(path)
			watcherPath.ignores[path] = struct{}{}
			return nil
		}
		log.WatchFile(path)
		watcherPath.watches[path] = struct{}{}
		return nil
	}); err != nil {
		return nil, err
	}
	return watcherPath, nil

}

func (r *WatcherConfig) Walk(watcher *fsnotify.Watcher, opt *Option) (*WatcherPath, error) {
	watcherPath, err := r.WalkWithDirName(watcher, opt, ".")
	if err != nil {
		return nil, err
	}
	return watcherPath, nil
}

type WatcherPath struct {
	ignores map[string]struct{}
	watches map[string]struct{}
	wcs     []*WatcherConfig
	opt     *Option
}

func NewWatcherPath(wcs []*WatcherConfig, option *Option) *WatcherPath {
	return &WatcherPath{
		ignores: map[string]struct{}{},
		watches: map[string]struct{}{},
		wcs:     wcs,
		opt:     option,
	}
}

func (w *WatcherPath) Merge(wp *WatcherPath) {
	for ignore := range wp.ignores {
		w.ignores[ignore] = struct{}{}
	}
	for watch := range wp.watches {
		w.watches[watch] = struct{}{}
	}
}

func (w *WatcherPath) AddIfNeeds(path string, watcher *fsnotify.Watcher) error {
	if strings.Contains(path, "tmp___") {
		return nil
	}
	if strings.Contains(path, "old___") {
		return nil
	}
	if _, exists := w.ignores[path]; exists {
		return nil
	}
	if _, exists := w.watches[path]; exists {
		return nil
	}
	for _, wc := range w.wcs {
		shouldWatch, err := wc.ShouldWatchFile(path, w.opt)
		if err != nil {
			return err
		}
		if shouldWatch {
			log.WatchFile(path)
			if err := watcher.Add(path); err != nil {
				return err
			}
			w.watches[path] = struct{}{}
			return nil
		}
	}
	log.IgnoreFile(path)
	w.ignores[path] = struct{}{}
	return nil
}
