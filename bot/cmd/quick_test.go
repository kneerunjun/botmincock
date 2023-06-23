package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/kneerunjun/botmincock/biz"
	"github.com/kneerunjun/botmincock/bot/core"
	"github.com/kneerunjun/botmincock/bot/resp"
	"github.com/kneerunjun/botmincock/dbadp"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	TEST_MONGO_HOST = "localhost:37017"
	TEST_MONGO_DB   = "botmincock_test"
	TEST_MONGO_COLL = "accounts"
)

func TestSendPoll(t *testing.T) {
	options := []string{
		"Avva Adams",
		"Alison Tyler",
		"August Ames",
		"Lisa Ann",
		"Little Caprice",
	}
	jOptions, _ := json.Marshal(options)
	url := fmt.Sprintf("https://api.telegram.org/bot6133190482:AAFdMU-49W7t9zDoD5BIkOFmtc-PR7-nBLk/sendPoll?chat_id=-902469479&is_anonymous=False&open_period=20&question=Favorite pornstar&options=%s", jOptions)
	cl := &http.Client{Timeout: 3 * time.Second}
	req, _ := http.NewRequest("POST", url, nil)
	resp, err := cl.Do(req)
	assert.Nil(t, err, "Error when sending poll via http")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Unexpected response status code when sending poll")
}

// TestDebitAdjustment: A cron jb can adjust the daily debits to set recoveries for the day
// but this needs to be tested for all the border cases
func TestDebitAdjustment(t *testing.T) {
	// Setting up the connection and the database
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	accounts := sess.DB("").C("accounts")
	accounts.RemoveAll(bson.M{})

	transacs := sess.DB("").C("transacs")
	transacs.RemoveAll(bson.M{})

	estimates := sess.DB("").C("estimates")
	estimates.RemoveAll(bson.M{})

	expenses := sess.DB("").C("expenses")
	expenses.RemoveAll(bson.M{})
	t.Log("Resetting the database. .")

	// TEST: when the total monthly expense is < 5  - we are all settled up
	anyCmd := &core.AnyBotCmd{MsgId: 454839589, ChatId: 5435435875, SenderId: 5157350442} // message id and charid dont have a relevance here
	cmd := &AdjustPlayDebitBotCmd{AnyBotCmd: anyCmd}
	adp := dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "transacs")
	r := cmd.Execute(core.NewExecCtx().SetDB(adp))
	_, ok := r.(*resp.TxtBotResp)
	assert.True(t, ok, "Unexpected type of bot response %s", reflect.TypeOf(r).String())

	// TEST: expenses !=0 but no one played today
	expData := []*biz.Expense{
		{TelegID: 5157350442, DtTm: time.Now(), Desc: "Court bookings", INR: 10000},
		{TelegID: 5157350442, DtTm: time.Now(), Desc: "Purchase of MAVIS 350", INR: 3850},
	}
	for _, d := range expData {
		expenses.Insert(d)
	}
	// Correspoding transactions for the expenses
	transacData := []*biz.Transac{
		{TelegID: 5157350442, Credit: 10000, Debit: 0.0, Desc: "Court bookings", DtTm: time.Now()},
		{TelegID: 5157350442, Credit: 3850, Debit: 0.0, Desc: "Purchase of MAVIS 350", DtTm: time.Now()},
	}
	for _, d := range transacData {
		transacs.Insert(d)
	}
	t.Log("Added test data for expenses &trasactions")
	r = cmd.Execute(core.NewExecCtx().SetDB(adp))
	_, ok = r.(*resp.TxtBotResp)
	assert.True(t, ok, "Unexpected type of bot response %s", reflect.TypeOf(r).String())
	t.Log("Tested for the case when nobody played on the day")

	plydyTrnsc := []*biz.Transac{
		{TelegID: 5157350442, Credit: 0.0, Debit: 150.0, Desc: biz.PLAYDAY_DESC, DtTm: time.Now()},
		{TelegID: 498116745, Credit: 0.0, Debit: 150.0, Desc: biz.PLAYDAY_DESC, DtTm: time.Now()},
	}
	for _, d := range plydyTrnsc {
		transacs.Insert(d)
	}
	r = cmd.Execute(core.NewExecCtx().SetDB(adp))
	_, ok = r.(*resp.TxtBotResp)
	assert.True(t, ok, "Unexpected type of bot response %s", reflect.TypeOf(r).String())
	t.Log("Tested for the case when nobody played on the day")
}

// TestAttendCmd : since this command involves calling multiple business layer functions, it needs a thorough test
func TestAttendCmd(t *testing.T) {
	t.Log("setting up the test..")
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	accounts := sess.DB("").C("accounts")
	accounts.RemoveAll(bson.M{})

	transacs := sess.DB("").C("transacs")
	transacs.RemoveAll(bson.M{})

	estimates := sess.DB("").C("estimates")
	estimates.RemoveAll(bson.M{})

	expenses := sess.DB("").C("expenses")
	expenses.RemoveAll(bson.M{})

	// Adding expenses
	expenseData := []*biz.Expense{
		{TelegID: 5157350442, DtTm: time.Now(), Desc: "Court bookings", INR: 10000},
		{TelegID: 5157350442, DtTm: time.Now(), Desc: "Purchase MAVIS350", INR: 3800},
	}
	for _, d := range expenseData {
		expenses.Insert(d)
	}
	anyCmd := &core.AnyBotCmd{MsgId: 454839589, ChatId: 5435435875, SenderId: 5157350442}
	cmd := &AttendanceBotCmd{AnyBotCmd: anyCmd}
	adp := dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "transacs")

	// TEST: for no account MarkAttendance
	t.Log("Now testing marking attendance when account does not exists")
	resp := cmd.Execute(&core.CmdExecCtx{DBAdp: adp})
	assert.Equal(t, "*resp.ErrBotResp", reflect.TypeOf(resp).String(), "Unexpected type of response")

	// TEST: testing for player already marked attendance
	t.Log("Now testing for user already marked for the day")
	archv := false
	develev := biz.AccElev(uint8(2))
	acc := biz.UserAccount{TelegID: 5157350442, Email: "", Name: "", Elevtn: &develev, Archived: &archv}
	accounts.Insert(&acc)
	// create a new debit for the day for the user
	trnsc := biz.Transac{TelegID: 5157350442, Credit: 0.0, Debit: 150.00, Desc: biz.PLAYDAY_DESC, DtTm: time.Now()}
	transacs.Insert(trnsc) // now the user is already marked for the day
	resp = cmd.Execute(&core.CmdExecCtx{DBAdp: adp})
	assert.Equal(t, "*resp.ErrBotResp", reflect.TypeOf(resp).String(), "Unexpected type of response")
	// removing the transaction for further tests
	transacs.RemoveAll(bson.M{})

	// TEST: everyone has opted out of play, or no one has answered the poll
	t.Log("Now testing when there arent any estimates at all ..")
	resp = cmd.Execute(&core.CmdExecCtx{DBAdp: adp})
	assert.Equal(t, "*resp.ErrBotResp", reflect.TypeOf(resp).String(), "Unexpected type of response")

	// TEST: now testing player attendance whose estimates arent found
	t.Log("Now testing when zero or no estimates for player, default attendance")
	accounts.Insert(biz.UserAccount{TelegID: 961044876, Email: "someone@chutiya.com", Name: "Kallu Bhosadiwala", Archived: &archv})
	ests := []*biz.Estimate{
		{TelegID: 5157350442, PlyDys: 30, DtTm: time.Now()},
		{TelegID: 498116745, PlyDys: 15, DtTm: time.Now()},
		{TelegID: 5116645118, PlyDys: 30, DtTm: time.Now()},
	}
	for _, d := range ests {
		estimates.Insert(d)
	}
	anyCmd = &core.AnyBotCmd{MsgId: 454839589, ChatId: 5435435875, SenderId: 961044876}
	cmd = &AttendanceBotCmd{AnyBotCmd: anyCmd}
	resp = cmd.Execute(&core.CmdExecCtx{DBAdp: adp})
	assert.Equal(t, "*resp.TxtBotResp", reflect.TypeOf(resp).String(), "Unexpected type of response")

	// TEST: recovery till now should give back 0 since its 01-JUN
	transacs.RemoveAll(bson.M{}) // clearing all transactions

	anyCmd = &core.AnyBotCmd{MsgId: 454839589, ChatId: 5435435875, SenderId: 5157350442}
	cmd = &AttendanceBotCmd{AnyBotCmd: anyCmd}
	resp = cmd.Execute(&core.CmdExecCtx{DBAdp: adp})
	assert.Equal(t, "*resp.TxtBotResp", reflect.TypeOf(resp).String(), "Unexpected type of response")
	t.Log(resp.UserMessage())
}
