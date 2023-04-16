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

func invalid(u *UserAccount) bool {
	if u.Name == "" || u.Email == "" || u.TelegID != int64(0) {
		return true
	}
	return false
}
func GetAccOfID(uid int64, findOneInStore func(flt bson.M) (map[string]interface{}, error)) (*UserAccount, error) {
	if uid == int64(0) || uid < int64(0) {
		// negative account ids signifies all the groups
		// a group cannot have a valid account on botmincock database
		return nil, fmt.Errorf("invalid account id to get, cannot be empty, or negative")
	}
	result, err := findOneInStore(bson.M{"tid": uid, "archive": false})
	if err != nil {
		return nil, fmt.Errorf("failed to get account")
	}
	byt, _ := json.Marshal(result)
	ua := &UserAccount{}
	if json.Unmarshal(byt, ua) != nil {
		return nil, fmt.Errorf("account of id %d could not be read", uid)
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
func RegisterNewUser(name, email string, tid int64, addToStore func(obj interface{}) error, duplicate func([]bson.M) (bool, error)) error {
	ua := &UserAccount{Name: name, TelegID: tid, Email: email, Elevtn: AccElev(0), Archived: false}
	if !invalid(ua) {
		return fmt.Errorf("invalid account details, can you check and send again?")
	}
	// for checking duplicates either account with same name or email should exists
	yes, err := duplicate([]bson.M{{"name": name}, {"email": email}})
	if err != nil {
		return fmt.Errorf("failed t check for account duplicacy")
	}
	if yes {
		return fmt.Errorf("an account with the same name/telegram ID already registered")
	}
	// if you are here then its time to register the account
	return addToStore(ua) // if the callback has an error then ofcourse we send that back to the calling function
}

func PatchAccofID(id int64, email string, update func(flt, patch bson.M) (map[string]interface{}, error)) (*UserAccount, error) {
	result, err := update(bson.M{"tid": id}, bson.M{"$set": bson.M{"email": email}})
	if err != nil {
		return nil, err
	}
	byt, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	ua := UserAccount{}
	err = json.Unmarshal(byt, &ua)
	if err != nil {
		return nil, fmt.Errorf("account of id %d could not be read", id)
	}
	return &ua, nil
}
