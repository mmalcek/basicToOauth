package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/kardianos/service"
)

const VERSION = "1.0.2"

type (
	program struct {
		exit chan struct{}
	}
	tFlags struct {
		svc *string
	}
)

var (
	logger    service.Logger
	tokensMap = tTokens{tokens: make(map[string]*tToken)}
)

func main() {
	if err := os.Chdir(filepath.Dir(os.Args[0])); err != nil {
		log.Fatal(err.Error())
	}

	flags := tFlags{}
	flags.svc = flag.String("service", "", "Control the system service (start, stop, install, uninstall)")
	flag.Parse()

	options := make(service.KeyValue)
	options["Restart"] = "on-success"
	options["SuccessExitStatus"] = "1 2 8 SIGKILL"

	svcConfig := &service.Config{
		Name:        "basicToOauth",
		DisplayName: "basicToOauth",
		Description: "basic to oauth auth proxy",
		Option:      options,
	}
	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	if len(*flags.svc) != 0 {
		err := service.Control(s, *flags.svc)
		if err != nil {
			logger.Errorf("Valid actions: %q\n", service.ControlAction)
			logger.Errorf(err.Error())
			os.Exit(1)
		}
		return
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}

}

func (p *program) Start(s service.Service) error {
	p.exit = make(chan struct{})
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	time.Sleep(1 * time.Second)
	logger.Info("Stopped")
	close(p.exit)
	return nil
}
