package fresher

import (
	"fmt"
	"os"
	"path/filepath"

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

type GlobalExclude struct {
	Files []string `json:"file"`
	Dirs  []string `json:"dir"`
}

func (g *GlobalExclude) IsExcludeDir(dir string) bool {
	for _, d := range g.Dirs {
		if d == dir {
			return true
		}
	}
	return false
}

func (g *GlobalExclude) IsExcludeFile(fileName string) (bool, error) {
	for _, f := range g.Files {
		ok, err := filepath.Match(f, fileName)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

type RecursiveDir struct {
	Name         string          `yaml:"name"`
	ExcludeFiles []string        `yaml:"exclude"`
	Dirs         []*RecursiveDir `yaml:"dir"`
}

func (r *RecursiveDir) IsExcludeFile(fileName string) (bool, error) {
	for _, f := range r.ExcludeFiles {
		ok, err := filepath.Match(f, fileName)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func (r *RecursiveDir) WalkWithDirName(watcher *fsnotify.Watcher, opt *Option, dirName string) error {
	global := opt.globalExclude
	if global.IsExcludeDir(r.Name) {
		return nil
	}
	if err := filepath.Walk(filepath.Join(dirName, r.Name), func(path string, info os.FileInfo, err error) error {
		if err != nil && err != filepath.SkipDir {
			return err
		}
		if info.IsDir() && info.Name() != r.Name {
			return filepath.SkipDir
		}

		isGlobalExclude, err := global.IsExcludeFile(info.Name())
		if err != nil {
			return err
		}
		if isGlobalExclude {
			return nil
		}
		isExclude, err := r.IsExcludeFile(info.Name())
		if err != nil {
			return err
		}
		if isExclude {
			return nil
		}
		if !opt.exts.IsIncludeSameExt(path) {
			return nil
		}
		fmt.Println("Wathed File:", path)
		if err := watcher.Add(path); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	for _, dir := range r.Dirs {
		if err := dir.WalkWithDirName(watcher, opt, filepath.Join(dirName, r.Name)); err != nil {
			return err
		}
	}
	return nil

}
func (r *RecursiveDir) Walk(watcher *fsnotify.Watcher, opt *Option) error {
	if err := r.WalkWithDirName(watcher, opt, "."); err != nil {
		return err
	}
	return nil
}
