package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kneerunjun/botmincock/bot/cmd"
	"github.com/kneerunjun/botmincock/bot/core"
	"github.com/kneerunjun/botmincock/dbadp"
	log "github.com/sirupsen/logrus"
)

// HandlrBotInContext : will inject the bot as an object in the context
func HandlrBotInContext(b core.Bot) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("bot", b)
	}
}

// HndlrPlaydayEstimates : this handles getting http command to send the poll for getting the estimates
// This will trigger sending the poll to the group once every month as per scheduled cron job
// Does not require the command infra  .. can send
func HndlrPlaydayEstimates(c *gin.Context) {
	val, _ := c.Get("bot")
	bot := val.(core.Bot)

	qs := fmt.Sprintf("Availability for %s %%3F", time.Now().AddDate(0, 1, 0).Month().String()) // the question of the poll
	anon := "False"
	options := []string{
		"All days",
		"15 days",
		"Only on weekends",
		"Out for the month",
	}
	jOptions, _ := json.Marshal(options)
	url := bot.(core.BotUrl).SendPollUrl(anon, qs, string(jOptions))
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

type HttpListenServlet struct {
	Srvr *http.Server
	Bot  core.Bot
}

func (hls *HttpListenServlet) Run(*RunConfig) error {
	r := gin.Default()
	r.GET("debits/adjust", HandlrBotInContext(hls.Bot), HandlrDebitAdjustments)
	r.GET("playdays/estimate", HandlrBotInContext(hls.Bot), HndlrPlaydayEstimates)
	hls.Srvr = &http.Server{
		Addr:    ":3333",
		Handler: r,
	}
	return hls.Srvr.ListenAndServe()
}
func (hls *HttpListenServlet) Halt(ctx context.Context) {
	hls.Srvr.Shutdown(ctx)
}

// SendBotHttp: Common service method that can , when given the url POST it to the telegram api server
// Can log errors and returns nil when success
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
				"url":    url,
			}).Error("error sending message over http")
			return err
		} else {
			if resp.StatusCode != http.StatusOK {
				log.WithFields(log.Fields{
					"status": resp.StatusCode,
					"url":    url,
				}).Error("unfavourable reponse from server")
				return fmt.Errorf("unfavourable reponse from server")
			}
		}
		return nil
	}
}
