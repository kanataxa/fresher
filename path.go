package fresher

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"

	"github.com/fsnotify/fsnotify"
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

type WatcherPath struct {
	Name     string   `yaml:"name"`
	Excludes []string `yaml:"exclude"`
	Includes []string `yaml:"include"`
}

func (r *WatcherPath) UnmarshalYAML(b []byte) error {
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

func (r *WatcherPath) IsExclude(path string) (bool, error) {
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

func (r *WatcherPath) IsInclude(path string) (bool, error) {
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

func (r *WatcherPath) WalkWithDirName(watcher *fsnotify.Watcher, opt *Option, dirName string) error {
	global := opt.globalExclude
	isExclude, err := global.IsExclude(r.Name)
	if err != nil {
		return err
	}
	if isExclude {
		return nil
	}
	if err := filepath.Walk(filepath.Join(dirName, r.Name), func(path string, info os.FileInfo, err error) error {
		if err != nil && err != filepath.SkipDir {
			return err
		}
		if info.IsDir() && info.Name() != r.Name {
			isInclude, err := r.IsInclude(info.Name())
			if err != nil {
				return err
			}
			if isInclude {
				return nil
			}
			return filepath.SkipDir
		}

		isGlobalExclude, err := global.IsExclude(info.Name())
		if err != nil {
			return err
		}
		if isGlobalExclude {
			return nil
		}
		isExclude, err := r.IsExclude(info.Name())
		if err != nil {
			return err
		}
		if isExclude {
			return nil
		}
		isInclude, err := r.IsInclude(info.Name())
		if err != nil {
			return err
		}
		if !isInclude {
			return nil
		}
		if !opt.exts.IsIncludeSameExt(path) {
			return nil
		}
		fmt.Println("Watched File:", path)
		if err := watcher.Add(path); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil

}
func (r *WatcherPath) Walk(watcher *fsnotify.Watcher, opt *Option) error {
	if err := r.WalkWithDirName(watcher, opt, "."); err != nil {
		return err
	}
	return nil
}
