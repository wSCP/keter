package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/wSCP/keter/keys"
)

var (
	setEnv       string = "XDG_CONFIG_HOME"
	setPath      string = "keter/keterrc"
	setFile      string
	chainExpiry  time.Duration = 2 * time.Second
	verbose      bool
	printVersion bool
)

func parseArgs() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [options]\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(
		&setEnv,
		"env",
		setEnv,
		"Provide an environment variable to locate the home directory of the configuration file.",
	)

	flag.StringVar(
		&setPath,
		"path",
		setPath,
		"Provide the path of the config file in the config home directory.",
	)

	flag.StringVar(
		&setFile,
		"file",
		setFile,
		"Provide a full path to a configuration file, overrides setEnv & setPath",
	)

	flag.DurationVar(
		&chainExpiry,
		"timeout",
		chainExpiry,
		"Timeout in seconds for the recording of chord chains.",
	)

	flag.BoolVar(
		&verbose,
		"verbose",
		verbose,
		"Verbose logging messages.",
	)

	flag.BoolVar(
		&printVersion,
		"version",
		printVersion,
		"Print the compiled program version and exit.",
	)

	flag.Parse()

	if printVersion {
		fmt.Printf(pkgVersion.Fmt())
		os.Exit(0)
	}
}

func init() {
	pkgVersion = newVersion(packageName, versionTag, versionHash, versionDate)
	parseArgs()
}

func main() {
	k := keys.New(
		keys.SetSettings(setEnv, setPath, setFile, chainExpiry, verbose),
	)

	cErr := k.Configure(k)
	if cErr != nil {
		k.Fatalf("configuration error: %s", cErr.Error())
	}

	go func() {
		k.Manage(k.Conn(), k.Pre, k.Post, k.Quit)
	}()

EVENTLOOP:
	for {
		select {
		case <-k.Pre:
			<-k.Post
		case sig := <-k.Sys:
			k.SignalHandler(sig)
		case <-k.Quit:
			break EVENTLOOP
		}
	}
}
