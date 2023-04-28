package biz

/* ==================================
author 		: kneerunjun@gmail.com
time		: April 2023
project		: botmincock
Domain logic behind the user account implemented here
Once the account is regisetred it has some basic maintenance tasks
All the functions here talk to interfaces and the models
====================================*/

import (
	"fmt"
	"reflect"
	"regexp"
)

// Set of standard user messages that can be reused in multiple locations in the business functions
const (
	TRY_AGAIN    = "Yikes! your command did not work as expected.Try to run again after a while,or if the problem persists kindly contact a sys-admin"
	DUPL_ACC     = "Account with same ID or email is already registered & active. Cannot register again."
	ACC_REACTIVE = "Account with same ID found archived. Reactivated the account."
	ACC_MISSN    = "Account you referring to does not exists. Please register first then try again."
	INVLD_PARAM  = "One/more params for the command supplied is invalid, Please consult a sys-admin and then try again."
)

var (
	REGX_EMAIL = regexp.MustCompile(`^[\w\d._]+@[\w]+.[\w\d]+$`)
)

// DbAdaptor : Agnostic of the database platform this can
type DbAdaptor interface {
	AddOne(interface{}) error
	RemoveOne(interface{}) error
	UpdateOne(interface{}) error
	GetOne(interface{}, reflect.Type) (interface{}, error)
	GetCount(interface{}, *int) error
}

func valid_account(ua *UserAccount) bool {
	return true
}

// RegisterNewAccount: checks for the account duplicates, validates the account values and then just pushes that to the database
// Any account when registered new will be marked archived = false
// any account when registered new will have elevation =0
// Icnase of already registered account but archived its re-enabled
func RegisterNewAccount(ua *UserAccount, iadp DbAdaptor) error {
	if !valid_account(ua) {
		return fmt.Errorf("account %s isnt a valid account", ua.Email)
	}
	// Checking to see if live accounts with the same teled id already registered
	// either accont with same telegid  or the email can cause the duplicate flag to be raised
	duplicate := 0
	if err := iadp.GetCount(&UserAccount{TelegID: ua.TelegID, Archived: false}, &duplicate); err != nil {
		return NewDomainError(fmt.Errorf("failed: checking account duplicate %d", ua.TelegID), err).SetLoc("RegisterNewAccount").SetUsrMsg(TRY_AGAIN)
	}
	if duplicate == 0 {
		return NewDomainError(fmt.Errorf("account with same id %d already registered", ua.TelegID), nil).SetLoc("RegisterNewAccount").SetUsrMsg(DUPL_ACC)
	}
	if err := iadp.GetCount(&UserAccount{Email: ua.Email, Archived: false}, &duplicate); err != nil {
		return NewDomainError(fmt.Errorf("failed: checking account duplicate %d", ua.TelegID), err).SetLoc("RegisterNewAccount").SetUsrMsg(TRY_AGAIN)
	}
	if duplicate != 0 {
		return NewDomainError(fmt.Errorf("account with email %s already registered", ua.Email), nil).SetLoc("RegisterNewAccount").SetUsrMsg(DUPL_ACC)
	}
	archived := 0
	if err := iadp.GetCount(&UserAccount{TelegID: ua.TelegID, Archived: true}, &archived); err != nil {
		return NewDomainError(fmt.Errorf("failed to get count of similar archived accounts %d", ua.TelegID), err).SetLoc("RegisterNewAccount").SetUsrMsg(TRY_AGAIN)
	}
	if archived != 0 {
		// this means the account just needs re-enablement and not registration
		// this happens with updating the email id as well.
		if err := iadp.UpdateOne(&UserAccount{TelegID: ua.TelegID, Archived: false, Email: ua.Email}); err != nil {
			return NewDomainError(fmt.Errorf("failed to re-enabled the account %d", ua.TelegID), err).SetLoc("RegisterNewAccount").SetUsrMsg(TRY_AGAIN)
		}
		return NewDomainError(fmt.Errorf("user account %d was found already registered, but was archived. Re-enabled it", ua.TelegID), nil).SetLoc("RegisterNewAccount").SetUsrMsg(ACC_REACTIVE)
	}
	// defaults when registering new account
	ua.Elevtn = AccElev(User)
	ua.Archived = false
	if err := iadp.AddOne(ua); err != nil {
		return NewDomainError(fmt.Errorf("failed: to add account %s %d", ua.Email, ua.TelegID), err).SetLoc("RegisterNewAccount").SetUsrMsg(TRY_AGAIN)
	}
	return nil
}

// ElevateAccount : comes in handy when an account has to be promoted in role
// sends back the account details after having elevated it
// Errors incase: account does not exists,requested elevation not within limits,query to update account fails,query to get the updated account details fails
func ElevateAccount(ua *UserAccount, iadp DbAdaptor) error {
	// Checking to see if account with same telegid exists and is not archived
	exists := 0
	if err := iadp.GetCount(&UserAccount{TelegID: ua.TelegID, Archived: false}, &exists); err != nil {
		return NewDomainError(fmt.Errorf("failed: checking account duplicate %d", ua.TelegID), err).SetLoc("ElevateAccount").SetUsrMsg(TRY_AGAIN)
	}
	if exists == 0 {
		// account does not exists - error
		return NewDomainError(fmt.Errorf("account not found registered  %d", ua.TelegID), nil).SetLoc("ElevateAccount").SetUsrMsg(ACC_MISSN)
	}
	if ua.Elevtn < AccElev(User) && ua.Elevtn > AccElev(Admin) {
		// has to be between permissible limits
		return NewDomainError(fmt.Errorf("requested account elevation is invalid %d", ua.Elevtn), nil).SetLoc("ElevateAccount").SetUsrMsg(INVLD_PARAM)
	}
	if err := iadp.UpdateOne(&UserAccount{TelegID: ua.TelegID, Elevtn: ua.Elevtn}); err != nil {
		// query error with adaptor, query has failed
		return NewDomainError(fmt.Errorf("failed query to update  account %d", ua.TelegID), err).SetLoc("ElevateAccount").SetUsrMsg(TRY_AGAIN)
	}
	// getting the updated account details
	updated, err := iadp.GetOne(&UserAccount{TelegID: ua.TelegID}, reflect.TypeOf(&UserAccount{}))
	if err != nil {
		return NewDomainError(fmt.Errorf("failed query to get account %d", ua.TelegID), err).SetLoc("ElevateAccount").SetUsrMsg(TRY_AGAIN)
	}
	// sending the updated account detatils
	x, _ := updated.(*UserAccount)
	*ua = *x
	return nil
}

// UpdateAccountEmail: changes the email attached to the account
// Errors when account not found registered
func UpdateAccountEmail(ua *UserAccount, iadp DbAdaptor) error {
	exists := 0
	err := iadp.GetCount(&UserAccount{TelegID: ua.TelegID, Archived: false}, &exists)
	if err != nil {
		return NewDomainError(fmt.Errorf("failed to check if the account is registered %d", ua.TelegID), err).SetLoc("UpdateAccountEmail").SetUsrMsg(TRY_AGAIN)
	}
	if exists == 0 {
		return NewDomainError(fmt.Errorf("no account %d found registered", ua.TelegID), nil).SetLoc("UpdateAccountEmail").SetUsrMsg(ACC_MISSN)
	}
	if ua.Email == "" || !REGX_EMAIL.MatchString(ua.Email) {
		// has to be between permissible limits
		return NewDomainError(fmt.Errorf("requested account email is invalid"), nil).SetLoc("UpdateAccountEmail").SetUsrMsg(INVLD_PARAM)
	}
	if err := iadp.UpdateOne(&UserAccount{TelegID: ua.TelegID, Email: ua.Email}); err != nil {
		return NewDomainError(fmt.Errorf("failed query to update  account %d", ua.TelegID), err).SetLoc("UpdateAccountEmail").SetUsrMsg(TRY_AGAIN)
	}
	updated, err := iadp.GetOne(&UserAccount{TelegID: ua.TelegID}, reflect.TypeOf(&UserAccount{}))
	if err != nil {
		return NewDomainError(fmt.Errorf("failed query to get account %d", ua.TelegID), err).SetLoc("UpdateAccountEmail").SetUsrMsg(TRY_AGAIN)
	}
	x, _ := updated.(*UserAccount)
	*ua = *x
	return nil
}

// DeregisterAccount :  flagging the account as archived
// Account information is never deleted since there is financial information connected to the account
// account is marked archived only to omit it from all other searches and operations
func DeregisterAccount(ua *UserAccount, iadp DbAdaptor) error {
	exists := 0
	err := iadp.GetCount(&UserAccount{TelegID: ua.TelegID, Archived: false}, &exists)
	if err != nil {
		return NewDomainError(fmt.Errorf("failed to check if the account is registered %d", ua.TelegID), err).SetLoc("DeregisterAccount").SetUsrMsg(TRY_AGAIN)
	}
	if exists == 0 {
		return NewDomainError(fmt.Errorf("no account %d found registered", ua.TelegID), nil).SetLoc("DeregisterAccount").SetUsrMsg(TRY_AGAIN)
	}
	if err := iadp.UpdateOne(&UserAccount{TelegID: ua.TelegID, Archived: true}); err != nil {
		return NewDomainError(fmt.Errorf("failed query to update  account %d", ua.TelegID), err).SetLoc("DeregisterAccount").SetUsrMsg(TRY_AGAIN)
	}
	return nil
}
