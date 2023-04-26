package business

import (
	"fmt"
	"reflect"
	"regexp"
)

var (
	REGX_EMAIL = regexp.MustCompile("^$")
)

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

func duplicate_account(ua *UserAccount) bool {
	return false
}

// RegisterNewAccount: checks for the account duplicates, validates the account values and then just pushes that to the database
// Any account when registered new will be marked archived = false
// any account when registered new will have elevation =0
// Icnase of already registered account but archived its re-enabled
func RegisterNewAccount(ua *UserAccount, iadp DbAdaptor) error {
	if !valid_account(ua) {
		return fmt.Errorf("account %s isnt a valid account", ua.Email)
	}
	duplicate := 0
	if err := iadp.GetCount(&UserAccount{TelegID: ua.TelegID}, &duplicate); err != nil {
		return fmt.Errorf("error checking for account duplicate %d: %s", ua.TelegID, err)
	}
	if duplicate != 0 {
		return fmt.Errorf("account with id %d already registered", ua.TelegID)
	}
	archived := 0
	if err := iadp.GetCount(&UserAccount{TelegID: ua.TelegID, Archived: true}, &archived); err != nil {
		return fmt.Errorf("failed to get count of similar accounts but archived %d: %s", ua.TelegID, err)
	}
	if archived != 0 {
		// this means the account just needs re-enablement and not registration
		// this happens with updating the email id as well.
		if err := iadp.UpdateOne(&UserAccount{TelegID: ua.TelegID, Archived: false, Email: ua.Email}); err != nil {
			return fmt.Errorf("failed to re-enabled the account %d", ua.TelegID)
		}
		return fmt.Errorf("user account %d was found already registered, but was archived. Re-enabled it", ua.TelegID)
	}
	// defaults when registering new account
	ua.Elevtn = AccElev(User)
	ua.Archived = false
	if err := iadp.AddOne(ua); err != nil {
		return fmt.Errorf("failed query to add account %s %d: %s", ua.Email, ua.TelegID, err)
	}
	return nil
}

// ElevateAccount : comes in handy when an account has to be promoted in role
func ElevateAccount(ua *UserAccount, iadp DbAdaptor) error {
	if !duplicate_account(ua) {
		return fmt.Errorf("account %s %d could not be found registered", ua.Email, ua.TelegID)
	}
	if ua.Elevtn < AccElev(User) && ua.Elevtn > AccElev(Admin) {
		// has to be between permissible limits
		return fmt.Errorf("requested account elevation is invalid, Has to be between [%d,%d]", User, Admin)
	}
	if err := iadp.UpdateOne(&UserAccount{TelegID: ua.TelegID, Elevtn: ua.Elevtn}); err != nil {
		return fmt.Errorf("failed query to update  account %d: %s", ua.TelegID, err)
	}
	updated, err := iadp.GetOne(&UserAccount{TelegID: ua.TelegID}, reflect.TypeOf(&UserAccount{}))
	if err != nil {
		return fmt.Errorf("failed query to get account %d: %s", ua.TelegID, err)
	}
	x, _ := updated.(*UserAccount)
	*ua = *x
	return nil
}

func UpdateAccountEmail(ua *UserAccount, iadp DbAdaptor) error {
	if !duplicate_account(ua) {
		return fmt.Errorf("account %s %d could not be found registered", ua.Email, ua.TelegID)
	}
	if ua.Email == "" || !REGX_EMAIL.MatchString(ua.Email) {
		// has to be between permissible limits
		return fmt.Errorf("requested account email is invalid")
	}
	if err := iadp.UpdateOne(&UserAccount{TelegID: ua.TelegID, Email: ua.Email}); err != nil {
		return fmt.Errorf("failed query to update  account %d: %s", ua.TelegID, err)
	}
	updated, err := iadp.GetOne(&UserAccount{TelegID: ua.TelegID}, reflect.TypeOf(&UserAccount{}))
	if err != nil {
		return fmt.Errorf("failed query to get account %d: %s", ua.TelegID, err)
	}
	x, _ := updated.(*UserAccount)
	*ua = *x
	return nil
}

func DeregisterAccount(ua *UserAccount, iadp DbAdaptor) error {
	if !duplicate_account(ua) {
		return fmt.Errorf("account %d could not be found registered", ua.TelegID)
	}
	if err := iadp.UpdateOne(&UserAccount{TelegID: ua.TelegID, Archived: true}); err != nil {
		return fmt.Errorf("failed query to update  account %d: %s", ua.TelegID, err)
	}
	return nil
}
