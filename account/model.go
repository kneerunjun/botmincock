package account

import "fmt"

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
	TelegID  int64   `bson:"tid" json:"tid"`         // telegram chat id for personal conversations, unique
	Email    string  `bson:"email" json:"email"`     // email of the user, for reports, unique
	Name     string  `bson:"name" json:"name"`       // name for addressing the user in any conversation, unique
	Elevtn   AccElev `bson:"elevtn" json:"elevtn"`   // priveleges for the user account 0-user, 1-manager, 2-admin, from specified enumerated values only
	Archived bool    `bson:"archive" json:"archive"` // default setting is false here
}

func (ua *UserAccount) ToMsgTxt() string {
	// https://stackoverflow.com/questions/3871729/transmitting-newline-character-n
	return fmt.Sprintf("Hi,%s%%0AYou are registered with us%%0AEmail: %s%%0ATelegramID: %d", ua.Name, ua.Email, ua.TelegID)
}
