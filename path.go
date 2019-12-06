package fresher

import (
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// TODO: support regexp
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
	Files        []string        `yaml:"file"`
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

func (r *RecursiveDir) WalkWithDirName(watcher *fsnotify.Watcher, global *GlobalExclude, dirName string) error {
	if global.IsExcludeDir(r.Name) {
		return nil
	}
	// Ignored global exclude files
	if len(r.Files) == 0 && len(r.ExcludeFiles) == 0 {
		if err := watcher.Add(filepath.Join(dirName, r.Name)); err != nil {
			return err
		}
	} else {
		for _, file := range r.Files {
			isGlobalExclude, err := global.IsExcludeFile(file)
			if err != nil {
				return err
			}
			if isGlobalExclude {
				continue
			}
			isExclude, err := r.IsExcludeFile(file)
			if err != nil {
				return err
			}
			if isExclude {
				continue
			}
			if err := watcher.Add(filepath.Join(dirName, r.Name, file)); err != nil {
				return err
			}
		}
	}

	for _, dir := range r.Dirs {
		if err := dir.WalkWithDirName(watcher, global, filepath.Join(dirName, r.Name)); err != nil {
			return err
		}
	}
	return nil

}
func (r *RecursiveDir) Walk(watcher *fsnotify.Watcher, global *GlobalExclude) error {
	if err := r.WalkWithDirName(watcher, global, "."); err != nil {
		return err
	}
	return nil
}
