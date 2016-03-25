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

var l *log.Logger

var provided *args

type args struct {
	ConfigEnv   string
	ConfigPath  string
	ConfigFile  string
	ChainExpiry time.Duration
	Verbose     bool
	Version     bool
}

func defaultArgs() *args {
	return &args{
		"XDG_CONFIG_HOME",
		"keter/keterrc",
		"",
		2 * time.Second,
		false,
		false,
	}
}

func (a *args) configPath() string {
	if a.ConfigFile != "" {
		return a.ConfigFile
	}
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
		l.Println("SIGINT")
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
		l.Println(fmt.Sprintf("received signal %v", s))
	}
	if provided.Verbose && msg.Len() != 0 {
		l.Println(msg.String())
	}
}

func parseArgs() {
	c := defaultArgs()

	flag.Usage = func() {
		fmt.Printf("Usage: %s [options]\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(
		&c.ConfigEnv,
		"configEnv",
		c.ConfigEnv,
		"Provide an environment variable to locate the home directory of the configuration file.",
	)

	flag.StringVar(
		&c.ConfigPath,
		"configPath",
		c.ConfigPath,
		"Provide the path of the config file in the config home directory.",
	)

	flag.StringVar(
		&c.ConfigFile,
		"configFile",
		c.ConfigFile,
		"Provide a full path to a configuration file, overrides configEnv & configPath",
	)

	flag.DurationVar(
		&c.ChainExpiry,
		"timeout",
		c.ChainExpiry,
		"Timeout in seconds for the recording of chord chains.",
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
	l = log.New(os.Stderr, "KETER: ", log.Ldate|log.Lmicroseconds)
	pkgVersion = newVersion(packageName, versionTag, versionHash, versionDate)
	parseArgs()
}

func main() {
	chains, err := LoadConfig(provided.configPath())
	if err != nil {
		l.Fatalf("configuration loading error: %s", err.Error())
	}

	hndl, err := NewHandlr("")
	if err != nil {
		l.Fatalf("handler configuration error: %s", err.Error())
	}

	err = Configure(hndl, chains)
	if err != nil {
		l.Fatalf("key chain configuration error: %s", err.Error())
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
