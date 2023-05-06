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

func TestAddNewExpense(t *testing.T) {
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C("expenses")
	defer coll.RemoveAll(bson.M{})
	defer t.Log(warnMessage("now clearing the database.."))
	t.Log(infoMessage("now testing for one sanple expense"))
	d := &Expense{INR: 1055.00, TelegID: 5157350442, DtTm: time.Now(), Desc: "test expense, purchase of shuttles"}
	err := RecordExpense(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "expenses"))
	assert.Nil(t, err, "unexpected error when recording an expense")

	// TEST: for negative test cases
	dataNotOK := []*Expense{
		nil, // nil expense is outrightly rejected
		{TelegID: 5157350442, INR: 0.0},
		{TelegID: 5157350442},
	}
	t.Log(infoMessage("now testing negative cases"))
	for _, d := range dataNotOK {
		err := RecordExpense(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "expenses"))
		assert.NotNil(t, err, "unexpected nil err when data not ok")
	}
}

func TestRegisterAccount(t *testing.T) {
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C(TEST_MONGO_COLL)
	// TEST: happy test, no error
	dataOk := []*UserAccount{
		{TelegID: 5435345, Email: "cayce0@bbb.org", Name: "Conrado Ayce"},
		{TelegID: 5435346, Email: "rscimoni1@paypal.com", Name: "Reagen Scimon"},
		{TelegID: 5435347, Email: "ekornousek2@apple.com", Name: "Elyssa Kornousek"},
	}
	for _, d := range dataOk {
		err := RegisterNewAccount(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
		assert.Nil(t, err, "Unexpected error when registering new account")
		// After having inserted the accounts we need to test if the account has elevation 0 and archive set to false
		acc := UserAccount{}
		coll.Find(bson.M{"tid": d.TelegID}).One(&acc)
		assert.Equal(t, AccElev(User), *acc.Elevtn, fmt.Sprintf("Unexpected elevation for the account %d", d.TelegID))
		assert.False(t, *acc.Archived, fmt.Sprintf("Unexpected Archive flag for the accounts %d", d.TelegID))
	}

	t.Log(infoMessage("Done testing happy path accounts"))
	// ===========================
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
	t.Log(infoMessage("Done testing for accounts with invalid email"))
	// ===========================
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
	t.Log(infoMessage("Done testing for accounts with duplicate ID/email"))
	// ===========================
	// TEST: testing for archived accounts - does it re-register the acocunt
	// ============= Setting up archived accounts

	archive := true
	coll.Update(&UserAccount{TelegID: dataOk[0].TelegID}, bson.M{"$set": &UserAccount{Archived: &archive}})

	err := RegisterNewAccount(dataOk[0], dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
	assert.NotNil(t, err, "Unexpected nil error when registering archived account")
	t.Log(infoMessage("Done testing for accounts that are archived"))
	// ========================
	// cleaning up the test
	t.Log(warnMessage("Now cleannig up the database.."))
	coll.RemoveAll(bson.M{})
}

// https://github.com/fatih/color/blob/f4c431696a22e834b83444f720bd144b2dbbccff/color.go#L67
func warnMessage(m string) string {
	return fmt.Sprintf("\x1b[%dm!%s\x1b[0m", 36, m)
}
func infoMessage(m string) string {
	return fmt.Sprintf("\x1b[%dm>%s\x1b[0m", 34, m)
}

func TestElevateAcc(t *testing.T) {
	/*
		setting up the test
		- database connection
		- inserting the data
	*/
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C(TEST_MONGO_COLL)

	archive := false
	userElev := AccElev(User)
	dataOk := []*UserAccount{
		{Elevtn: &userElev, TelegID: 5435345, Email: "cayce0@bbb.org", Name: "Conrado Ayce", Archived: &archive},
		{Elevtn: &userElev, TelegID: 5435346, Email: "rscimoni1@paypal.com", Name: "Reagen Scimon", Archived: &archive},
		{Elevtn: &userElev, TelegID: 5435347, Email: "ekornousek2@apple.com", Name: "Elyssa Kornousek", Archived: &archive},
	}
	for _, d := range dataOk {
		coll.Insert(d)
	}
	t.Log(infoMessage("Database setup complete"))
	for _, d := range dataOk {
		*d.Elevtn = AccElev(Manager)
		err := ElevateAccount(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
		assert.Nil(t, err, warnMessage("failed elevate account"))
	}
	// TEST: testing for data not ok
	overElev := AccElev(uint(5)) // over elevation of an account should not be possible
	for _, d := range dataOk {
		d.Elevtn = &overElev
		err := ElevateAccount(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
		assert.NotNil(t, err, "Unexpected not nil error when testing with over elevation of the account")
	}

	/*
		Cleaning up the test database
	*/
	t.Cleanup(func() {
		t.Log(warnMessage("Now clearing up the test database.."))
		coll.RemoveAll(bson.M{})
	})
}

func TestUpdateAccountEmail(t *testing.T) {
	/*
		Setting up ithe database for email modification test
	*/
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C(TEST_MONGO_COLL)
	// cleaning up the test database

	archive := false
	userElev := AccElev(User)
	dataOk := []*UserAccount{
		{Elevtn: &userElev, TelegID: 5435345, Email: "cayce0@bbb.org", Name: "Conrado Ayce", Archived: &archive},
		{Elevtn: &userElev, TelegID: 5435346, Email: "rscimoni1@paypal.com", Name: "Reagen Scimon", Archived: &archive},
		{Elevtn: &userElev, TelegID: 5435347, Email: "ekornousek2@apple.com", Name: "Elyssa Kornousek", Archived: &archive},
	}
	for _, d := range dataOk {
		coll.Insert(d)
	}
	// TEST: positive test for error less update on the email of the account
	newEmails := []string{
		"sglancy0@github.io",
		"msach1@samsung.com",
		"agosnell2@lulu.com",
	}
	t.Log(infoMessage("Now testing for valid accounts.."))
	for idx, d := range dataOk {
		d.Email = newEmails[idx]
		err := UpdateAccountEmail(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
		assert.Nil(t, err, warnMessage(fmt.Sprintf("failed to update email for %s", d.Email)))
	}
	// TEST: account is nil or the email isnt valid
	invalidData := []*UserAccount{
		{TelegID: 5435345, Email: "mkonmann3@who.int.gov"},
		{TelegID: 5435345, Email: ""},
		nil,
	}
	t.Log(infoMessage("Now testing for invalid accounts.."))
	for _, d := range invalidData {
		err := UpdateAccountEmail(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
		assert.NotNil(t, err, warnMessage("unexpected nil error when updating invalid email and accounts"))
	}
	t.Cleanup(func() {
		t.Log(warnMessage("Now clearing up the test database.."))
		coll.RemoveAll(bson.M{})
	})
}

func TestDeregisterAccount(t *testing.T) {
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C(TEST_MONGO_COLL)
	// cleaning up the test database

	archive := false
	userElev := AccElev(User)
	dataOk := []*UserAccount{
		{Elevtn: &userElev, TelegID: 5435345, Email: "cayce0@bbb.org", Name: "Conrado Ayce", Archived: &archive},
		{Elevtn: &userElev, TelegID: 5435346, Email: "rscimoni1@paypal.com", Name: "Reagen Scimon", Archived: &archive},
		{Elevtn: &userElev, TelegID: 5435347, Email: "ekornousek2@apple.com", Name: "Elyssa Kornousek", Archived: &archive},
	}
	for _, d := range dataOk {
		coll.Insert(d)
	}
	// TEST: positive tests , account is marked as archive
	t.Log(infoMessage("Now testing for valid accounts.."))
	for _, d := range dataOk {
		err := DeregisterAccount(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
		assert.Nil(t, err, warnMessage(fmt.Sprintf("failed to deregister email for %s", d.Email)))
	}
	t.Cleanup(func() {
		t.Log(warnMessage("Now clearing up the test database.."))
		coll.RemoveAll(bson.M{})
	})
}

func TestEmailRegx(t *testing.T) {
	data := []string{
		"bdensumbe0@youku.com",
		"mroth3@soup.io",
		"rlambregts4@godaddy.com",
		"jhucker.be9@virginia.edu",
		"lgh_iona@newyorker.com",
		"kneerunjun@gmail.com",
		"awatiniranjan@gmail.com",
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
