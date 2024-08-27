package options

import (
	"flag"
)

type Options struct {
	Version bool

	ListenAddr   string
	Debug        bool
	ColorizeLogs bool
}

func Init() *Options {
	opts := new(Options)
	flag.BoolVar(&opts.Version, "v", false, "prints the version")
	flag.BoolVar(&opts.Version, "version", false, "prints the version")
	flag.BoolVar(&opts.Debug, "debug", false, "enable debug logging")
	flag.BoolVar(&opts.ColorizeLogs, "colorize-logs", false, "colorize log messages")
	flag.StringVar(&opts.ListenAddr, "address", ":29900", "server/bind address in format [host]:port")
	flag.Parse()
	return opts
}
