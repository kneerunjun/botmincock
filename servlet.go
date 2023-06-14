package main

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type RunConfig struct {
	WtGrp  *sync.WaitGroup
	Cancel chan bool // send your interrupt signal on this channel
}

type Servlet interface {
	Run(*RunConfig) error
	Halt(context.Context)
}

// RunServlet : For any servlet this will 2 task ,
// on one will watch cancel interruption and halt the  server
// on the other will call the servlet.Run()
// servlet.Run() is implmented for each of servlet specific
func RunServlet(srvlt Servlet, rc *RunConfig) error {
	defer rc.WtGrp.Done()
	go func() {
		<-rc.Cancel
		log.Warn("Closing down servlet")
		ctx, cncl := context.WithTimeout(context.Background(), 3*time.Second)
		defer cncl()
		srvlt.Halt(ctx)
	}()
	return srvlt.Run(rc)
}
