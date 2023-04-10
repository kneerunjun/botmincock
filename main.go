package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/kneerunjun/botmincock/botcore"

	log "github.com/sirupsen/logrus"
)

var (
	FVerbose, FLogF bool
	logFile         string
	allCommands     = []*regexp.Regexp{
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>registerme)(\s+)(?P<email>[\w\d._]+@[\w]+.[\w\d]+)+$`, os.Getenv("BOT_HANDLE"))),
	}
	textCommands = []*regexp.Regexp{
		regexp.MustCompile(`^(?P<cmd>(?i)gm)$`), // user intends to mark his attendance
		regexp.MustCompile(`^(?P<cmd>(?i)(good[\s]*morning))$`),
	}
)

const (
	MAX_COINC_UPDATES = 10
	BOT_TICK_SECS     = 3 * time.Second
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
	// loading the secrets
	f, err := os.Open("/run/secrets/token_secret")
	if err != nil {
		log.Fatal("failed to load bot access key, kindly check and run again")
	}
	rdr := bufio.NewReader(f)
	byt, _, err := rdr.ReadLine()
	if err != nil {
		log.Fatal("failed to read token for bot, check and run again")
	}
	tok := string(byt)
	log.WithFields(log.Fields{
		"token": tok,
	}).Debug("bot secret token is read..")
	log.Info("Grab your cocks, botmincocks is coming up now..")
	defer log.Warn("botmincock now shutting down")

	var wg sync.WaitGroup
	cancel := make(chan bool)
	defer close(cancel)
	// TODO: private keys cannot be exposed here
	// this has to come from secret files
	botmincock := botcore.NewTeleGBot(&botcore.BotConfig{Token: tok}, reflect.TypeOf(&botcore.SharedExpensesBot{}))
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
	// ---------- now starting to watch periodic updates
	// -------------------------------------------------
	botCallouts := make(chan botcore.BotUpdate, MAX_COINC_UPDATES)
	defer close(botCallouts)
	botCommands := make(chan botcore.BotUpdate, MAX_COINC_UPDATES)
	defer close(botCommands)
	txtMsgs := make(chan botcore.BotUpdate, MAX_COINC_UPDATES)
	defer close(txtMsgs)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// TODO: each filter shall have a dedicated channel that it shall dispatch the update to
		// as the filters execute in sequence, the update message is tested for type
		// once the type of the message is determined also for relevance it can be dispatched to the channel that is relevant
		// a filter can also abort the testing of subsequent filters
		filters := []botcore.BotUpdtFilter{
			&botcore.GrpConvFilter{PassChn: nil},
			&botcore.NonZeroIDFilter{PassChn: nil},
			&botcore.BotCommandFilter{PassChn: botCommands, CommandExprs: allCommands},
			&botcore.BotCalloutFilter{PassChn: botCallouts},
			&botcore.TextMsgCmdFilter{PassChn: txtMsgs, CommandExprs: textCommands},
		}
		botcore.WatchUpdates(cancel, botmincock, BOT_TICK_SECS, filters...)
	}()
	// ----------- now setting up the thread to consume updates
	// ---------------------------------------------------------
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case updt := <-botCallouts:
				// parsing the updates
				log.WithFields(log.Fields{
					"text": updt.Message.Text,
				}).Debug("Received a bot callout ..")
			case updt := <-botCommands:
				// parsing the updates
				log.WithFields(log.Fields{
					"text": updt.Message.Text,
				}).Debug("Received a bot command ..")
			case updt := <-txtMsgs:
				// parsing the updates
				log.WithFields(log.Fields{
					"text": updt.Message.Text,
				}).Debug("Received a text command ..")
			case <-cancel:
				return
			}
		}
	}()
	wg.Wait()

}
