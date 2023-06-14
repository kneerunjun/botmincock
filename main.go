package main

import (
	"bufio"
	"context"
	"encoding/json"
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

	"github.com/gin-gonic/gin"
	"github.com/kneerunjun/botmincock/bot/cmd"
	"github.com/kneerunjun/botmincock/bot/core"
	"github.com/kneerunjun/botmincock/bot/resp"
	"github.com/kneerunjun/botmincock/bot/updt"
	"github.com/kneerunjun/botmincock/dbadp"
	"github.com/kneerunjun/botmincock/seeds"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	FVerbose, FLogF, FSeed bool
	logFile                string
	allCommands            = []*regexp.Regexp{
		/*
			Registering new account
			editing account email
			checking personal information
			de-registering an account
			elevating the account
		*/
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>registerme)(\s+)(?P<email>[\w\d._]+@[\w]+.[\w\d]+)+$`, os.Getenv("BOT_HANDLE"))),
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>editme)(\s+)(?P<email>[\w\d._]+@[\w]+.[\w\d]+)+$`, os.Getenv("BOT_HANDLE"))),
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>myinfo)$`, os.Getenv("BOT_HANDLE"))),
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>deregisterme)$`, os.Getenv("BOT_HANDLE"))),
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>elevateacc)(\s+)(?P<accid>[\d]+)$`, os.Getenv("BOT_HANDLE"))),
		/*
			Adding  expenses
			Check for personal expenses
			Check for entire team expenses
		*/
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>addexpense)(\s+)(?P<inr>[0-9]+)(\s+)(?P<desc>[^!@#\$%%\^&\*\(\\)\[\]\<\\>]*)$`, os.Getenv("BOT_HANDLE"))),
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>paydues)(\s+)(?P<inr>[0-9]+)$`, os.Getenv("BOT_HANDLE"))),
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>mydues)$`, os.Getenv("BOT_HANDLE"))),
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>myexpenses)$`, os.Getenv("BOT_HANDLE"))),
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>allexpenses)$`, os.Getenv("BOT_HANDLE"))),
		/*
			Help listing of all the commands
			calling out the bot and the sending the /help command shall send a list of commands
		*/
		regexp.MustCompile(fmt.Sprintf(`^%s(\s+)\/(?P<cmd>help)$`, os.Getenv("BOT_HANDLE"))),
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
	flag.BoolVar(&FSeed, "seed", false, "Flag if the db needs to be force seeded")
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

	if FSeed {
		log.Warn("Flushing data, and re-seeding it")
		err, accs := seeds.Accounts("seeds/acc.json")
		if err != nil {
			log.Error(err)
		} else {
			mongoSession.DB(DB_NAME).C("accounts").RemoveAll(bson.M{})
			for _, d := range accs {
				mongoSession.DB(DB_NAME).C("accounts").Insert(d)
			}
			log.WithFields(log.Fields{
				"count": len(accs),
			}).Info("Seeded accounts")
		}
		err, ests := seeds.Estimates("seeds/estimates.json")
		if err != nil {
			log.Error(err)
		} else {
			mongoSession.DB(DB_NAME).C("estimates").RemoveAll(bson.M{})
			for _, d := range ests {
				mongoSession.DB(DB_NAME).C("estimates").Insert(d)
			}
		}

		err, expns := seeds.Expenses("seeds/expense.json")
		if err != nil {
			log.Error(err)
		} else {
			mongoSession.DB(DB_NAME).C("expenses").RemoveAll(bson.M{})
			for _, d := range expns {
				mongoSession.DB(DB_NAME).C("expenses").Insert(d)
			}
		}

		err, transacs := seeds.Transactions("seeds/transac.json")
		if err != nil {
			log.Error(err)
		} else {
			mongoSession.DB(DB_NAME).C("transacs").RemoveAll(bson.M{})
			for _, d := range transacs {
				mongoSession.DB(DB_NAME).C("transacs").Insert(d)
			}
		}

	}
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

	benv := core.EnvOrDefaults()
	log.WithFields(log.Fields{
		"handle":  benv.Handle,
		"name":    benv.Name,
		"grpid":   benv.GrpID,
		"oid":     benv.OwnerID,
		"baseurl": benv.BaseURL,
		"token":   benv.Token,
	}).Debug("bot environment=")
	botmincock := core.NewTeleGBot(&core.SharedExpensesBotEnv{BotEnv: benv, GuestCharges: 150.00}, reflect.TypeOf(&core.SharedExpensesBot{}))
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
				respChn <- resp.NewTextResponse("Did you mean to command me? This isn't valid command", updt.Message.Chat.Id, updt.Message.Id)
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
				go func() {
					commnd, err := cmd.ParseTextCmd(updt, textCommands)
					if err != nil {
						respChn <- resp.NewErrResponse(err, "ParseTextCmd", "Did not quite understand the command, can you try again?", updt.Message.Chat.Id, updt.Message.Id)
					} else {
						cmdcoll, ok := commnd.(cmd.CmdForColl)
						if !ok {
							respChn <- resp.NewErrResponse(fmt.Errorf("failed to read collection name for the command"), "ParseTextCmd", "Some internal error could not parse your command", updt.Message.Id, updt.Message.Id)
						} else {
							respChn <- commnd.Execute(cmd.NewExecCtx().SetDB(dbadp.NewMongoAdpator(MONGO_ADDRS, DB_NAME, cmdcoll.CollName())))
						}
					}
				}()
			case resp := <-respChn:
				go func() {
					// resp.Log()
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
	// Starting a small http server so that we can callup from cron jobs
	// Daily cronjobs can call this server to get chores done
	wg.Add(1)
	go func() {
		defer wg.Done()
		r := gin.Default()
		r.GET("debits/adjust", HandlrDebitAdjustments)
		r.GET("playdays/estimate", HndlrPlaydayEstimates)
		srvr := &http.Server{
			Addr:    ":3333",
			Handler: r,
		}
		ctx, cncl := context.WithTimeout(context.TODO(), 3*time.Second)
		defer cncl()
		go func() {
			<-cancel
			srvr.Shutdown(ctx)
		}()
		srvr.ListenAndServe()
		log.Warn("Now closing the http server..")
	}()
	wg.Wait()

}
func SendBotHttp(url string) error {
	cl := http.Client{Timeout: STD_REQ_TIMEOUT}
	// url := fmt.Sprintf("%s%s", baseUrl, resp.SendMsgUrl())
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("send message request could not be created")
		return err
	} else {
		resp, err := cl.Do(req)
		if err != nil {
			log.WithFields(log.Fields{
				"status": resp.StatusCode,
				"err":    err,
			}).Error("error sending message over http")
			return err
		} else {
			log.WithFields(log.Fields{
				"status": resp.StatusCode,
			}).Info("responded..")
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unfavourable reponse from server when sending message from bot")
		}
		return nil
	}
}

// HndlrPlaydayEstimates : this handles getting http command to send the poll for getting the estimates
// This will trigger sending the poll to the group once every month as per scheduled cron job
// Does not require the command infra  .. can send
func HndlrPlaydayEstimates(c *gin.Context) {
	log.Debug("Received request to send poll for estimates")

	qs := fmt.Sprintf("Availability for %s %%3F", time.Now().AddDate(0, 1, 0).Month().String()) // the question of the poll
	anon := "False"
	chatID := -902469479
	// expiry := time.Now().Add(24 * time.Hour).UnixMilli()
	options := []string{
		"All days",
		"15 days",
		"Only on weekends",
		"Out for the month",
	}
	jOptions, _ := json.Marshal(options)
	url := fmt.Sprintf("https://api.telegram.org/bot6133190482:AAFdMU-49W7t9zDoD5BIkOFmtc-PR7-nBLk/sendPoll?chat_id=%d&is_anonymous=%s&question=%s&options=%s", chatID, anon, qs, jOptions)
	if err := SendBotHttp(url); err != nil {
		c.AbortWithStatus(http.StatusBadGateway)
		return
	}
}

func HandlrDebitAdjustments(c *gin.Context) {
	log.Debug("Received request to adjust daily debits")
	// We send in a bot text response whenever the debits are adjusted
	command := cmd.AdjustPlayDebitBotCmd{AnyBotCmd: &cmd.AnyBotCmd{ChatId: -902469479}}
	ctx := cmd.NewExecCtx().SetDB(dbadp.NewMongoAdpator(MONGO_ADDRS, DB_NAME, "transacs"))
	resp := command.Execute(ctx)
	if resp != nil {
		// TODO: here we need the bot url to send the message
		// But unless we have the bot instance getting the url isnt really possible
		// HACK: we are just hardcoding the boturl here for the time being
		if err := SendBotHttp(fmt.Sprintf("%s%s", "https://api.telegram.org/bot6133190482:AAFdMU-49W7t9zDoD5BIkOFmtc-PR7-nBLk", resp.SendMsgUrl())); err != nil {
			c.AbortWithStatus(http.StatusBadGateway)
			return
		}
		c.AbortWithStatus(http.StatusOK)
		return
	}
	c.AbortWithStatus(http.StatusNotFound)
}
