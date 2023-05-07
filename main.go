package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/kneerunjun/botmincock/bot/cmd"
	"github.com/kneerunjun/botmincock/bot/core"
	"github.com/kneerunjun/botmincock/bot/resp"
	"github.com/kneerunjun/botmincock/bot/updt"
	"github.com/kneerunjun/botmincock/dbadp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

var (
	FVerbose, FLogF bool
	logFile         string
	allCommands     = []*regexp.Regexp{
		// register new user
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>registerme)(\s+)(?P<email>[\w\d._]+@[\w]+.[\w\d]+)+$`, os.Getenv("BOT_HANDLE"))),
		// email edit
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>editme)(\s+)(?P<email>[\w\d._]+@[\w]+.[\w\d]+)+$`, os.Getenv("BOT_HANDLE"))),
		// getting user info
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>myinfo)$`, os.Getenv("BOT_HANDLE"))),
		// archiving the user account
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>deregisterme)$`, os.Getenv("BOT_HANDLE"))),
		// elevating the account to higher roles
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>elevateacc)(\s+)(?P<accid>[\d]+)$`, os.Getenv("BOT_HANDLE"))),
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>addexpense)(\s+)(?P<inr>[0-9]+)(\s+)(?P<desc>[^!@#\$%%\^&\*\(\\)\[\]\<\\>]*)$`, os.Getenv("BOT_HANDLE"))),
	}
	textCommands = []*regexp.Regexp{
		regexp.MustCompile(`^(?P<cmd>(?i)gm)$`), // user intends to mark his attendance
		regexp.MustCompile(`^(?P<cmd>(?i)(good[\s]*morning))$`),
	}
)

const (
	MAX_COINC_UPDATES = 10
	BOT_TICK_SECS     = 3 * time.Second
	STD_REQ_TIMEOUT   = 5 * time.Second
	MONGO_ADDRS       = "mongostore:27017"
	DB_NAME           = "botmincock"
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

	/* ======================
	parsing command line flags and setting up log
	- set verbose=true to change the level of logging to maximum and include all the noise
	- set the flog=true to change the direction of logging to file instead of the terminal
	=========================*/

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
	/*===============================
	Getting the mongo connection on
	- https://gist.github.com/345161974/4f2048f90584a64891cf07997bfd9e23
	- after geting the handle to the session the DB is passed as execution context
	=================================*/
	storeDialInfo := &mgo.DialInfo{
		Addrs:    []string{MONGO_ADDRS},
		Timeout:  10 * time.Second,
		Database: DB_NAME,
	}
	mongoSession, err := mgo.DialWithInfo(storeDialInfo)
	if err != nil || mongoSession == nil {
		log.Fatalf("failed to dial connection with store: %s\n", err)
	}
	mongoSession.SetMode(mgo.Monotonic, true)
	log.Info("Now connected to the database..")
	/* ============================
	loading the secrets
	- from files on the local repository to container secrets
	- you can store the token of the bot here
	===============================*/
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
	// Starting the bot
	log.Info("Grab your cocks, botmincocks is coming up now..")
	defer log.Warn("botmincock now shutting down")

	/* =============================
	Initializing the bot
	- making the interrupt channel
	- creating an instance of the bot with the configuration
	- listening for signals for interruptions
	================================*/
	var wg sync.WaitGroup
	cancel := make(chan bool)
	defer close(cancel)
	// TODO: private keys cannot be exposed here
	// this has to come from secret files
	botmincock := core.NewTeleGBot(&core.BotConfig{Token: tok, MaxCoincUpdates: MAX_COINC_UPDATES}, reflect.TypeOf(&core.SharedExpensesBot{}))
	log.WithFields(log.Fields{
		"bot token": botmincock.Token(),
	}).Debug("just checking in on the bot token configuration")
	if botmincock == nil {
		log.Fatal("failed to instantiate bot..")
	}
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
	/* ====================
	now starting to watch periodic updates
	- list of filters
	- as the filters execute in sequence, the update message is tested for type
	- once the type of the message is determined also for relevance it can be dispatched to the channel that is relevant
	- a filter can also abort the testing of subsequent filters, typically when it has found content that is relevant to it
	=======================*/
	botCallouts := make(chan updt.BotUpdate, MAX_COINC_UPDATES)
	defer close(botCallouts)
	botCommands := make(chan updt.BotUpdate, MAX_COINC_UPDATES)
	defer close(botCommands)
	txtMsgs := make(chan updt.BotUpdate, MAX_COINC_UPDATES)
	defer close(txtMsgs)
	wg.Add(1)
	go func() {
		defer wg.Done()
		filters := []updt.BotUpdtFilter{
			&updt.GrpConvFilter{PassChn: nil},
			&updt.NonZeroIDFilter{PassChn: nil},
			&updt.BotCommandFilter{PassChn: botCommands, CommandExprs: allCommands},
			&updt.BotCalloutFilter{PassChn: botCallouts},
			&updt.TextMsgCmdFilter{PassChn: txtMsgs, CommandExprs: textCommands},
		}
		updt.WatchUpdates(cancel, botmincock, BOT_TICK_SECS, filters...)
	}()
	// ----------- now setting up the thread to consume updates
	// ---------------------------------------------------------
	// whatever the bot action it sends back the response on this channel
	//
	respChn := make(chan resp.BotResponse, MAX_COINC_UPDATES)
	defer close(respChn)
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
				respChn <- resp.NewTextResponse("Thats whats called as a bot callout.. tag me and I see that as an opportunity to serve you", updt.Message.Chat.Id, updt.Message.Id)
			case updt := <-botCommands:
				// handling bot commands on separate coroutine
				go func() {
					commnd, err := cmd.ParseBotCmd(updt, allCommands)
					if err != nil {
						respChn <- resp.NewErrResponse(err, "ParseBotCmd", "Did not quite understand the command, can you try again?", updt.Message.Chat.Id, updt.Message.Id)
					} else {
						cmdcoll, ok := commnd.(cmd.CmdForColl)
						if !ok {
							respChn <- resp.NewErrResponse(fmt.Errorf("failed to read collection name for the command"), "ParseBotCmd", "Some internal error could not parse your command", updt.Message.Id, updt.Message.Id)
						} else {
							respChn <- commnd.Execute(cmd.NewExecCtx().SetDB(dbadp.NewMongoAdpator(MONGO_ADDRS, DB_NAME, cmdcoll.CollName())))
						}
					}
				}()
			case updt := <-txtMsgs:
				// parsing the updates
				log.WithFields(log.Fields{
					"text": updt.Message.Text,
				}).Debug("Received a text command ..")
				respChn <- resp.NewTextResponse("Thats a text command. Certain text I can recognise as commands for me", updt.Message.Chat.Id, updt.Message.Id)
			case resp := <-respChn:
				go func() {
					resp.Log()
					cl := http.Client{Timeout: STD_REQ_TIMEOUT}
					url := fmt.Sprintf("%s%s", botmincock.UrlBot(), resp.SendMsgUrl())
					req, err := http.NewRequest("POST", url, nil)
					if err != nil {
						log.WithFields(log.Fields{
							"err": err,
						}).Error("send message request could not be created")
					} else {
						resp, err := cl.Do(req)
						if err != nil {
							log.WithFields(log.Fields{
								"status": resp.StatusCode,
								"err":    err,
							}).Error("error sending message over http")
						} else {
							log.WithFields(log.Fields{
								"status": resp.StatusCode,
							}).Info("responded..")
						}
					}

				}()
			case <-cancel:
				return
			}
		}
	}()
	wg.Wait()

}
