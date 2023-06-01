package cmd

import (
	"reflect"
	"testing"
	"time"

	"github.com/kneerunjun/botmincock/biz"
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
	anyCmd := &AnyBotCmd{MsgId: 454839589, ChatId: 5435435875, SenderId: 5157350442}
	cmd := &AttendanceBotCmd{AnyBotCmd: anyCmd}
	adp := dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "transacs")

	// TEST: for no account MarkAttendance
	t.Log("Now testing marking attendance when account does not exists")
	resp := cmd.Execute(&CmdExecCtx{DBAdp: adp})
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
	resp = cmd.Execute(&CmdExecCtx{DBAdp: adp})
	assert.Equal(t, "*resp.ErrBotResp", reflect.TypeOf(resp).String(), "Unexpected type of response")
	// removing the transaction for further tests
	transacs.RemoveAll(bson.M{})

	// TEST: everyone has opted out of play, or no one has answered the poll
	t.Log("Now testing when there arent any estimates at all ..")
	resp = cmd.Execute(&CmdExecCtx{DBAdp: adp})
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
	anyCmd = &AnyBotCmd{MsgId: 454839589, ChatId: 5435435875, SenderId: 961044876}
	cmd = &AttendanceBotCmd{AnyBotCmd: anyCmd}
	resp = cmd.Execute(&CmdExecCtx{DBAdp: adp})
	assert.Equal(t, "*resp.TxtBotResp", reflect.TypeOf(resp).String(), "Unexpected type of response")

	// TEST: recovery till now should give back 0 since its 01-JUN
	transacs.RemoveAll(bson.M{}) // clearing all transactions

	anyCmd = &AnyBotCmd{MsgId: 454839589, ChatId: 5435435875, SenderId: 5157350442}
	cmd = &AttendanceBotCmd{AnyBotCmd: anyCmd}
	resp = cmd.Execute(&CmdExecCtx{DBAdp: adp})
	assert.Equal(t, "*resp.TxtBotResp", reflect.TypeOf(resp).String(), "Unexpected type of response")
	t.Log(resp.UserMessage())
}
