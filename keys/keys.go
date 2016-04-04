package keys

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Keys struct {
	Configuration
	*log.Logger
	*Settings
	Loader
	*loop
	Handle
}

func New(cnf ...Config) *Keys {
	return &Keys{
		Configuration: newConfiguration(cnf...),
		loop:          newLoop(),
	}
}

type loop struct {
	Pre  chan struct{}
	Post chan struct{}
	Quit chan struct{}
	Comm chan string
	Sys  chan os.Signal
}

func newLoop() *loop {
	l := &loop{
		make(chan struct{}, 0),
		make(chan struct{}, 0),
		make(chan struct{}, 0),
		make(chan string, 0),
		make(chan os.Signal, 0),
	}

	signal.Notify(
		l.Sys,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGCHLD,
		syscall.SIGPIPE,
	)

	return l
}

func (k *Keys) SignalHandler(s os.Signal) {
	msg := new(bytes.Buffer)
	switch s {
	case syscall.SIGINT:
		k.Println("SIGINT")
		os.Exit(0)
	case syscall.SIGHUP:
		msg.WriteString("Got signal SIGHUP, reconfiguring....\n")
		err := loadAndConfigureChains(k)
		if err != nil {
			msg.WriteString(err.Error())
		}
	default:
		k.Println(fmt.Sprintf("received signal %v", s))
	}
	if k.Verbose && msg.Len() != 0 {
		k.Println(msg.String())
	}
}
