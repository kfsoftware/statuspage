package main

import (
	"github.com/kfsoftware/statuspage/cmd"
	"github.com/newrelic/go-agent/v3/integrations/logcontext/nrlogrusplugin"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	log.SetFormatter(nrlogrusplugin.ContextFormatter{})
	//log.SetFormatter(customFormatter)
	if err := cmd.NewCmdStatusPage().Execute(); err != nil {
		os.Exit(1)
	}
}
