package biz

import (
	"fmt"
	"time"
)

/* ==================================
author 		: kneerunjun@gmail.com
time		: April 2023
project		: botmincock
Account objects that face the database are required
Account objects represent objects that store the basic facts about the account
For all other collections Account info serves as the base for getting unique ids
+ CRUD functions for the account
====================================*/

type AccElev uint8

func (e AccElev) Stringify() string {
	if e == AccElev(User) {
		return "User"
	} else if e == AccElev(Manager) {
		return "Manager"
	} else if e == AccElev(Admin) {
		return "Admin"
	}
	return "Unknown"
}

// Account elevation as enumeration
const (
	User = uint8(0) + iota
	Manager
	Admin
)

// UserAccount : this has the basic user account facts
// refer to this to know the details of the user
// NOTE: we arent using bson.ObjectIds here since telegID is sufficent Primary key
type UserAccount struct {
	// IMP: the omitempty is important since when the object is passed into mongo as filters partially filled.
	// All the fields that have 0 / empty values to them would be omitted when  marshalling to bson
	// it is then easy to confuse the default value and the set value (if they are same). Omit empty will actually leave it out
	// Hence we then take it to be pointer to datatypes
	TelegID  int64    `bson:"tid,omitempty" json:"tid"`         // telegram chat id for personal conversations, unique
	Email    string   `bson:"email,omitempty" json:"email"`     // email of the user, for reports, unique
	Name     string   `bson:"name,omitempty" json:"name"`       // name for addressing the user in any conversation, unique
	Elevtn   *AccElev `bson:"elevtn,omitempty" json:"elevtn"`   // priveleges for the user account 0-user, 1-manager, 2-admin, from specified enumerated values only
	Archived *bool    `bson:"archive,omitempty" json:"archive"` // default setting is false here
}

func (ua *UserAccount) ToMsgTxt() string {
	/*
		%%0A 	: new line character
		%%09	: tab space
		%c		: emoticon to appear correctly in chat
	*/
	return fmt.Sprintf("Hi, %s%%0A%c Registered with us%%0A%c%%09%s%%0A%c%%09%d%%0A%c%%09%s", ua.Name, EMOJI_greentick, EMOJI_email, ua.Email, EMOJI_badge, ua.TelegID, EMOJI_sheild, (*ua.Elevtn).Stringify())
}

// Transac :towards account maintenance - each row is a credit / debit attributed to the account
// denormalized this has to be aggregated to see the balance of the account
type Transac struct {
	TelegID int64     `bson:"tid" json:"tid"`
	Credit  float32   `bson:"credit" json:"credit"`
	Debit   float32   `bson:"debit" json:"debit"`
	Desc    string    `bson:"desc,omitempty" json:"desc"`
	DtTm    time.Time `bson:"dttm" json:"dttm"`
}

func (t *Transac) ToMsgTxt() string {
	return fmt.Sprintf("%.2f(Cr) %0.2f(Dr) for account %d", t.Credit, t.Debit, t.TelegID)
}

// TransacQ : when querying on the transactions collection we often need a time span to query from
// The summation of credits and debits in separate fields too
type TransacQ struct {
	Credits float32 `bson:"credits"`
	Debits  float32 `bson:"debits"`
	TelegID int64
	Desc    string
	From    time.Time
	To      time.Time
}

type Balance struct {
	TelegID int64     `bson:"tid" json:"tid"`
	Due     float32   `bson:"due"`
	DtTm    time.Time `bson:"dttm"`
}

func (b *Balance) ToMsgTxt() string {
	return fmt.Sprintf("Account balance %0.2f", b.Due)
}

// Start of each month the playdays are estimated from polls in the group
// each month - each acccount the play days help you arrive at the share holding of the account in the monthly expenditure
// Estimates are pivotal to calculating the monthly paybacks or dues
type Estimate struct {
	TelegID int64     `bson:"tid" json:"tid"`
	PlyDys  int       `bson:"plydys" json:"plydys"`
	DtTm    time.Time `bson:"dttm"`
}

// Any purchases on behalf of the bot as a manager by any account towards goods/services shared by the group is recorded as an expense
type Expense struct {
	TelegID int64     `bson:"tid,omitempty" json:"tid"`
	DtTm    time.Time `bson:"dttm,omitempty" json:"dttm"`
	Desc    string    `bson:"desc,omitempty" json:"desc"`
	INR     float32   `bson:"inr,omitempty" json:"inr"`
}

func (exp *Expense) ToMsgTxt() string {
	return fmt.Sprintf("total expense %.2f for account %d", exp.INR, exp.TelegID)
}

// MnthlyExpnsQry : when querying for the monthly expense aggregates this serves as the flywheel object
type MnthlyExpnsQry struct {
	TelegID int64     `bson:"tid"`   // relevant only when querying for user aggregate expenses
	Total   float32   `bson:"total"` // gets the aggregate monthly expense as a result
	Dttm    time.Time // used only for querying the monthly expense has no bson counter
}

func (ume *MnthlyExpnsQry) ToMsgTxt() string {
	// https://stackoverflow.com/questions/3871729/transmitting-newline-character-n
	if ume.TelegID != int64(0) {
		return fmt.Sprintf("Account: %d%%0ATotal expenditure for %s: %c%.2f", ume.TelegID, ume.Dttm.Month().String(), EMOJI_rupee, ume.Total)
	} else {
		// when what is needed is total expenses for the team
		return fmt.Sprintf("Total team expenditure for %s: %c%.2f", ume.Dttm.Month().String(), EMOJI_rupee, ume.Total)
	}

}
