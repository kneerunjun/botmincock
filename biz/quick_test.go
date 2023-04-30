package biz

import (
	"fmt"
	"testing"
	"time"

	"github.com/kneerunjun/botmincock/dbadp"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	TEST_MONGO_HOST = "localhost:27017"
	TEST_MONGO_DB   = "botmincock_test"
	TEST_MONGO_COLL = "accounts"
)

func TestRegisterAccount(t *testing.T) {
	// TEST: happy test, no error
	dataOk := []*UserAccount{
		{TelegID: 5435345, Email: "cayce0@bbb.org", Name: "Conrado Ayce"},
		{TelegID: 5435346, Email: "rscimoni1@paypal.com", Name: "Reagen Scimon"},
		{TelegID: 5435347, Email: "ekornousek2@apple.com", Name: "Elyssa Kornousek"},
	}
	for _, d := range dataOk {
		err := RegisterNewAccount(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
		assert.Nil(t, err, "Unexpected error when registering new account")
	}
	// TEST: dummy count =1 hence the account being registered is duplicate
	for _, d := range dataOk {
		err := RegisterNewAccount(d, &dbadp.DummyAdaptor{DummyCount: 1})
		assert.NotNil(t, err, "Unexpected error when registering new account")
	}
	// TEST: email of the account being registered isnt valid
	dataNotOk := []*UserAccount{
		{TelegID: 5435345, Email: "cayce0@bbb.org.cm", Name: "Conrado Ayce"},
		{TelegID: 5435346, Email: "rsci%^$%^moni1@paypal.com", Name: "Reagen Scimon"},
		{TelegID: 5435347, Email: "@apple.com", Name: "Elyssa Kornousek"},
	}
	for _, d := range dataNotOk {
		err := RegisterNewAccount(d, &dbadp.DummyAdaptor{DummyCount: 0})
		assert.NotNil(t, err, "Unexpected nil error when registering new account")
	}

	// TEST:  testing for duplicate accounts
	dataDuplicate := []*UserAccount{
		{TelegID: 5435345, Email: "cayce0@bbb.org", Name: "Conrado Ayce"},
		{TelegID: 5435348, Email: "rscimoni1@paypal.com", Name: "Reagen Scimon"}, // email is repeated
		{TelegID: 5435347, Email: "ekornousek2@apple.com", Name: "Elyssa Kornousek"},
	}
	for _, d := range dataDuplicate {
		err := RegisterNewAccount(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
		assert.NotNil(t, err, "Unexpected nil error when registering new account")
	}
	// TEST: testing for archived accounts - does it re-register the acocunt
	t.Log("Now testing to see if archived accounts are re-enabled")
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C(TEST_MONGO_COLL)
	archive := true
	coll.Update(&UserAccount{TelegID: dataOk[0].TelegID}, bson.M{"$set": &UserAccount{Archived: &archive}})
	err := RegisterNewAccount(dataOk[0], dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
	assert.NotNil(t, err, "Unexpected nil error when registering archived account")
	// cleaning up the test
	coll.RemoveAll(bson.M{})
}

func TestEmailRegx(t *testing.T) {
	data := []string{
		"bdensumbe0@youku.com",
		"mroth3@soup.io",
		"rlambregts4@godaddy.com",
		"jhucker.be9@virginia.edu",
		"lgh_iona@newyorker.com",
		"kneerunjun@gmail.com",
	}
	dataNotOk := []string{
		"ahubatsch8@bbc.co.uk",
		"fpetrazzib@state.tx.us",
		"@state.tx.us",
		"fpetrazzibstate.us",
		"#fpetrazzib@state.us",
	}
	for _, d := range data {
		assert.True(t, REGX_EMAIL.MatchString(d), fmt.Sprintf("email %s should have passed the test", d))
	}
	for _, d := range dataNotOk {
		assert.False(t, REGX_EMAIL.MatchString(d), fmt.Sprintf("email %s shouldn't have passed the test", d))
	}
}
