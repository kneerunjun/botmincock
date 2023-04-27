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
		return fmt.Errorf("error checking for account duplicate %d: %s", ua.TelegID, err)
	}
	if duplicate != 0 {
		return fmt.Errorf("account with id %d already registered", ua.TelegID)
	}
	if err := iadp.GetCount(&UserAccount{Email: ua.Email, Archived: false}, &duplicate); err != nil {
		return fmt.Errorf("error checking for account duplicate %d: %s", ua.TelegID, err)
	}
	if duplicate != 0 {
		return fmt.Errorf("account with email %s already registered", ua.Email)
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
// sends back the account details after having elevated it
// Errors incase: account does not exists,requested elevation not within limits,query to update account fails,query to get the updated account details fails
func ElevateAccount(ua *UserAccount, iadp DbAdaptor) error {
	// Checking to see if account with same telegid exists and is not archived
	exists := 0
	if err := iadp.GetCount(&UserAccount{TelegID: ua.TelegID, Archived: false}, &exists); err != nil {
		return fmt.Errorf("error checking for account duplicate %d: %s", ua.TelegID, err)
	}
	if exists == 0 {
		// account does not exists - error
		return fmt.Errorf("account not found registered  %d", ua.TelegID)
	}
	if ua.Elevtn < AccElev(User) && ua.Elevtn > AccElev(Admin) {
		// has to be between permissible limits
		return fmt.Errorf("requested account elevation is invalid, Has to be between [%d,%d]", User, Admin)
	}
	if err := iadp.UpdateOne(&UserAccount{TelegID: ua.TelegID, Elevtn: ua.Elevtn}); err != nil {
		// query error with adaptor, query has failed
		return fmt.Errorf("failed query to update  account %d: %s", ua.TelegID, err)
	}
	// getting the updated account details
	updated, err := iadp.GetOne(&UserAccount{TelegID: ua.TelegID}, reflect.TypeOf(&UserAccount{}))
	if err != nil {
		return fmt.Errorf("failed query to get account %d: %s", ua.TelegID, err)
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
		return fmt.Errorf("failed to check if the account is registered %d", ua.TelegID)
	}
	if exists == 0 {
		return fmt.Errorf("no account %d found registered", ua.TelegID)
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
	exists := 0
	err := iadp.GetCount(&UserAccount{TelegID: ua.TelegID, Archived: false}, &exists)
	if err != nil {
		return fmt.Errorf("failed to check if the account is registered %d", ua.TelegID)
	}
	if exists == 0 {
		return fmt.Errorf("no account %d found registered", ua.TelegID)
	}
	if err := iadp.UpdateOne(&UserAccount{TelegID: ua.TelegID, Archived: true}); err != nil {
		return fmt.Errorf("failed query to update  account %d: %s", ua.TelegID, err)
	}
	return nil
}
