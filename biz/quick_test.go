package biz

import (
	"fmt"
	"testing"

	"github.com/kneerunjun/botmincock/dbadp"
	"github.com/stretchr/testify/assert"
)

func TestRegisterAccount(t *testing.T) {
	// TEST: happy test, no error
	dataOk := []*UserAccount{
		{TelegID: 5435345, Email: "cayce0@bbb.org", Name: "Conrado Ayce"},
		{TelegID: 5435346, Email: "rscimoni1@paypal.com", Name: "Reagen Scimon"},
		{TelegID: 5435347, Email: "ekornousek2@apple.com", Name: "Elyssa Kornousek"},
	}
	for _, d := range dataOk {
		err := RegisterNewAccount(d, &dbadp.DummyAdaptor{DummyCount: 0})
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
		assert.NotNil(t, err, "Unexpected error when registering new account")
	}
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
