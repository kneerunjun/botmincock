package main

/* ==================================
author 		: kneerunjun@gmail.com
time		: April 2023
project		: botmincock
Account objects that face the database are required
Account objects represent objects that store the basic facts about the account
For all other collections Account info serves as the base for getting unique ids
+ CRUD functions for the account
====================================*/
import (
	"encoding/json"
	"fmt"

	"gopkg.in/mgo.v2/bson"
)

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
	TelegID  int64  `bson:"tid" json:"tid"`         // telegram chat id for personal conversations, unique
	Email    string  `bson:"email" json:"email"`     // email of the user, for reports, unique
	Name     string  `bson:"name" json:"name"`       // name for addressing the user in any conversation, unique
	Elevtn   AccElev `bson:"elevtn" json:"elevtn"`   // priveleges for the user account 0-user, 1-manager, 2-admin, from specified enumerated values only
	Archived bool    `bson:"archive" json:"archive"` // default setting is false here
}

func validate(u *UserAccount) bool {
	// TODO: validation logic here
	return true
}
func duplicate(u *UserAccount) bool {
	return false
}

func GetAccOfID(uid string, findOneInStore func(flt bson.M) (map[string]interface{}, error)) (*UserAccount, error) {
	if uid == "" {
		return nil, fmt.Errorf("invalid account id to get, cannot be empty")
	}
	result, err := findOneInStore(bson.M{"uid": uid, "archive": false})
	if err != nil {
		return nil, fmt.Errorf("failed to get account")
	}
	byt, _ := json.Marshal(result)
	ua := &UserAccount{}
	if json.Unmarshal(byt, ua) != nil {
		return nil, fmt.Errorf("account of id %s could not be read", uid)
	}
	return ua, nil
}

// DeregUser: pulls out the user from all the records, but since the balance tables still do have the information all what we can do is mark this account archived
func DeregUser(uid string, updtStore func(flt bson.M, set bson.M) error) error {
	if uid == "" {
		return fmt.Errorf("invalid accunt id to delete, cannot be empty")
	}
	// Updates the document instead of deleting and sets the archived flag
	return updtStore(bson.M{"uid": uid}, bson.M{"$set": bson.M{"archive": true}})
}

// RegisterNewUser : gets the details from the bot and registers a new user in the database
// UID is generated here as 16bitunique id that can be used to universally track the user
// Default elevation for the user when creating is 0
// To start with the account is never archived i
func RegisterNewUser(name, email string,tid int64, addToStore func(obj interface{}) error) error {
	ua := &UserAccount{Name: name, TelegID: tid, Email: email, Elevtn: AccElev(0), Archived: false}
	if !validate(ua) {
		return fmt.Errorf("invalid account details, can you check and send again?")
	}
	if duplicate(ua) {
		return fmt.Errorf("an account with the same name/telegram ID already registered")
	}
	// if you are here then its time to register the account
	return addToStore(ua) // if the callback has an error then ofcourse we send that back to the calling function
}
