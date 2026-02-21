package main

import (
	"os"

	"github.com/aiomayo/hdf/cmd"
	"github.com/charmbracelet/log"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	log.SetPrefix("hdf")
	log.SetReportTimestamp(false)
	log.SetLevel(log.InfoLevel)

	cmd.SetVersionInfo(version, commit, date)
	os.Exit(cmd.Execute())
}
