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

func ExcludePaths(paths []string) OptionFunc {
	return func(f *Fresher) {
		f.opt.excludePaths = paths
	}
}

func Extensions(exts []string) OptionFunc {
	return func(f *Fresher) {
		f.opt.exts = exts
	}
}

func IgnoreTest(ignoreTest bool) OptionFunc {
	return func(f *Fresher) {
		f.opt.ignoreTest = ignoreTest
	}
}

func WatchInterval(interval time.Duration) OptionFunc {
	return func(f *Fresher) {
		f.opt.interval = interval
	}
}
