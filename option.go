package fresher

import "time"

type OptionFunc func(f *Fresher)

func ExecCommand(command *Command) OptionFunc {
	return func(f *Fresher) {
		f.opt.command = command
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
