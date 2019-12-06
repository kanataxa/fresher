package fresher

import "time"

type OptionFunc func(f *Fresher)

func BuildPath(buildPath string) OptionFunc {
	return func(f *Fresher) {
		f.opt.buildPath = buildPath
	}
}

func WatchPaths(paths []*RecursiveDir) OptionFunc {
	return func(f *Fresher) {
		f.opt.paths = paths
	}
}

func GlobalExcludePath(global *GlobalExclude) OptionFunc {
	return func(f *Fresher) {
		f.opt.globalExclude = global
	}
}

func Extensions(exts []string) OptionFunc {
	return func(f *Fresher) {
		f.opt.exts = exts
	}
}

func WatchInterval(interval time.Duration) OptionFunc {
	return func(f *Fresher) {
		f.opt.interval = interval
	}
}
