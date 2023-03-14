package main

import (
	"flag"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
)

var (
	FVerbose, FLogF bool
	logFile         string
)

func init() {
	/* ======================
	-verbose=true would mean log.Debug can work
	-verbose=false would mean log.Debug will be hidden
	-flog=true: all the log output shall be onto a file
	-flog=false: all the log output shall be on stdout
	- We are setting the default log level to be Info level
	======================= */
	flag.BoolVar(&FVerbose, "verbose", false, "Level of logging messages are set here")
	flag.BoolVar(&FLogF, "flog", false, "Direction in which the log should output")
	// Setting up log configuration for the api
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	log.SetReportCaller(false)
	log.SetOutput(os.Stdout)    // FLogF will set it main, but dfault is stdout
	log.SetLevel(log.InfoLevel) // default level info debug but FVerbose will set it main
	logFile = os.Getenv("LOGF")

}

func main() {
	/*
		parsing command line flags and setting up log
	*/
	flag.Parse() // command line flags are parsed
	log.WithFields(log.Fields{
		"verbose": FVerbose,
		"flog":    FLogF,
	}).Info("Log configuration..")
	if FVerbose {
		log.SetLevel(log.DebugLevel)
	}
	if FLogF {
		lf, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("Failed to connect to log file, kindly check the privileges")
		} else {
			log.Infof("Check log file for entries @ %s", logFile)
			log.SetOutput(lf)
		}
	}
	log.Info("Grab your cocks, botmincocks is coming up now..")
	defer log.Warn("botmincock now shutting down")

	var wg sync.WaitGroup
	cancel := make(chan bool)
	defer close(cancel)
	botmincock := NewTeleGBot(&BotConfig{Token: "6133190482:AAFdMU-49W7t9zDoD5BIkOFmtc-PR7-nBLk"}, reflect.TypeOf(&SharedExpensesBot{}))
	wg.Add(1)
	go func() {
		defer wg.Done()
		// waits for any cancel signal from the terminal and then closes down all tasks
		sigs := make(chan os.Signal, 1)
		defer close(sigs)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		log.Debug("Now starting signal watcher..")
		<-sigs
		close(cancel)
	}()
	wg.Add(1)

	go func() {
		defer wg.Done()
		StartBot(cancel, botmincock)
	}()
	wg.Wait()

}
