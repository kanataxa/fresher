package fresher

import (
	"time"
)

type OptionFunc func(f *Fresher)

func ExecTarget(target string) OptionFunc {
	return func(f *Fresher) {
		f.opt.target = target
	}
}

func WatchConfigs(configs []*WatcherConfig) OptionFunc {
	return func(f *Fresher) {
		f.opt.configs = configs
	}
}

func GlobalExcludePath(global *GlobalExclude) OptionFunc {
	return func(f *Fresher) {
		f.opt.globalExclude = global
	}
}

func ExtensionPaths(exts Extensions) OptionFunc {
	return func(f *Fresher) {
		f.opt.exts = exts
	}
}

func WatchInterval(interval time.Duration) OptionFunc {
	return func(f *Fresher) {
		f.opt.interval = interval
	}
}
