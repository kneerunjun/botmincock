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
	TelegID  int64   `bson:"tid,omitempty" json:"tid"`         // telegram chat id for personal conversations, unique
	Email    string  `bson:"email,omitempty" json:"email"`     // email of the user, for reports, unique
	Name     string  `bson:"name,omitempty" json:"name"`       // name for addressing the user in any conversation, unique
	Elevtn   AccElev `bson:"elevtn,omitempty" json:"elevtn"`   // priveleges for the user account 0-user, 1-manager, 2-admin, from specified enumerated values only
	Archived *bool   `bson:"archive,omitempty" json:"archive"` // default setting is false here
}

func (ua *UserAccount) ToMsgTxt() string {
	// https://stackoverflow.com/questions/3871729/transmitting-newline-character-n
	return fmt.Sprintf("Hi,%s%%0AYou are registered with us%%0AEmail: %s%%0ATelegramID: %d", ua.Name, ua.Email, ua.TelegID)
}

// Transac :towards account maintenance - each row is a credit / debit attributed to the account
// denormalized this has to be aggregated to see the balance of the account
type Transac struct {
	TelegID int64     `bson:"tid" json:"tid"`
	Credit  int       `bson:"credit" json:"credit"`
	Debit   int       `bson:"debit" json:"debit"`
	Desc    string    `bson:"desc" json:"desc"`
	DtTm    time.Time `bson:"date" json:"date"`
}

// Start of each month the playdays are estimated from polls in the group
// each month - each acccount the play days help you arrive at the share holding of the account in the monthly expenditure
// Estimates are pivotal to calculating the monthly paybacks or dues
type Estimate struct {
	TelegID int64  `bson:"tid" json:"tid"`
	PlyDys  uint8  `bson:"plydys" json:"plydys"`
	Month   uint8  `bson:"mnth" json:"mnth"`
	Year    uint16 `bson:"yr" json:"yr"`
}

// Any purchases on behalf of the bot as a manager by any account towards goods/services shared by the group is recorded as an expense
type Expense struct {
	TelegID int64  `bson:"tid" json:"tid"`
	Month   uint8  `bson:"mnth" json:"mnth"`
	Year    uint16 `bson:"yr" json:"yr"`
	Desc    string `bson:"desc" json:"desc"`
	INR     int    `bson:"inr" json:"inr"`
}
