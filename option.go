package fresher

import "time"

type OptionFunc func(f *Fresher)

func BuildPath(buildPath string) OptionFunc {
	return func(f *Fresher) {
		f.opt.buildPath = buildPath
	}
}

func WatchPaths(paths []string) OptionFunc {
	return func(f *Fresher) {
		f.opt.paths = paths
	}
}

func WatchInterval(interval time.Duration) OptionFunc {
	return func(f *Fresher) {
		f.opt.interval = interval
	}
}
