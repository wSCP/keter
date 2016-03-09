package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"
)

var (
	provided    *args
	Logger      = log.New(os.Stderr, "KETER: ", log.Ldate|log.Lmicroseconds)
	pkgVersion  *version
	packageName string = "keter"
	versionTag  string = "No version tag supplied with compilation"
	versionHash string
	versionDate string
)

type args struct {
	ConfigEnv   string
	ConfigPath  string
	ChainExpiry time.Duration
	Verbose     bool
	Version     bool
}

func defaultArgs() *args {
	return &args{
		"XDG_CONFIG_HOME",
		"keter/keterrc",
		2 * time.Second,
		false,
		false,
	}
}

func (a *args) configPath() string {
	var pth string
	configHome := os.Getenv(a.ConfigEnv)
	if configHome != "" {
		pth = fmt.Sprintf("%s/%s", configHome, a.ConfigPath)
	} else {
		pth = fmt.Sprintf("%s/%s/%s", os.Getenv("HOME"), ".config", a.ConfigPath)
	}
	return pth
}

func SignalHandler(h Handlr, s os.Signal) {
	msg := new(bytes.Buffer)
	switch s {
	case syscall.SIGINT:
		Logger.Println("SIGINT")
		os.Exit(0)
	case syscall.SIGHUP:
		msg.WriteString("Got signal SIGHUP, reconfiguring....\n")
		chains, err := LoadConfig(provided.configPath())
		if err != nil {
			msg.WriteString(fmt.Sprintf("error while loading config: %s\n", err.Error()))
		}
		err = Configure(h, chains)
		if err != nil {
			msg.WriteString(fmt.Sprintf("error while configuring: %s\n", err))
		}
	default:
		Logger.Println(fmt.Sprintf("received signal %v", s))
	}
	if provided.Verbose && msg.Len() != 0 {
		Logger.Println(msg.String())
	}
}

func parseArgs() {
	c := defaultArgs()

	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] <input directories>\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(
		&c.ConfigEnv,
		"configEnv",
		c.ConfigEnv,
		"Provide an env variable to locate the home directory of the configuration file. Default is 'XDG_CONFIG_HOME.'",
	)

	flag.StringVar(
		&c.ConfigPath,
		"configPath",
		c.ConfigPath,
		"Provide the path of the config file in the config home directory. Default is 'keter/keterrc'",
	)

	flag.DurationVar(
		&c.ChainExpiry,
		"timeout",
		c.ChainExpiry,
		"Timeout in seconds for the recording of chord chains. Default is 2.",
	)

	flag.BoolVar(
		&c.Verbose,
		"verbose",
		c.Verbose,
		"Verbose logging messages.",
	)

	flag.BoolVar(
		&c.Version,
		"version",
		c.Version,
		"Print the compiled program version and exit.",
	)

	flag.Parse()

	if c.Version {
		fmt.Printf(pkgVersion.Fmt())
		os.Exit(0)
	}

	provided = c
}

func init() {
	parseArgs()
	pkgVersion = newVersion(packageName, versionTag, versionHash, versionDate)
}

func main() {
	chains, err := LoadConfig(provided.configPath())
	if err != nil {
		Logger.Fatalf("configuration loading error: %s", err.Error())
	}

	hndl, err := NewHandlr("")
	if err != nil {
		Logger.Fatalf("handler configuration error: %s", err.Error())
	}

	err = Configure(hndl, chains)
	if err != nil {
		Logger.Fatalf("key chain configuration error: %s", err.Error())
	}

	before, after, quit, signals := Loop(hndl)

EVENTLOOP:
	for {
		select {
		case <-before:
			<-after
		case sig := <-signals:
			SignalHandler(hndl, sig)
		case <-quit:
			break EVENTLOOP
		}
	}
}
