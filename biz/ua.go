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

	"github.com/kneerunjun/botmincock/dbadp"
	log "github.com/sirupsen/logrus"
)

// AccountInfo: gets the account information for given unique id
// ua		: in/out param sends in the teleg id for search and account information on return
// iadp		: db adaptor
// throws an error when account isnt found registered or error in query to get account information
func AccountInfo(ua *UserAccount, iadp dbadp.DbAdaptor) error {
	errLoc := "AccountInfo"
	if iadp == nil {
		return NewDomainError(ERR_DBCONN, nil).SetLoc(errLoc).SetUsrMsg(gateway_fail())
	}
	if ua == nil {
		return NewDomainError(ERR_NILACC, nil).SetLoc(errLoc).SetUsrMsg(invalid_account("getting nil account information"))
	}
	count := 0
	archive := false
	flt := &UserAccount{TelegID: ua.TelegID, Archived: &archive} // account being queried
	if err := iadp.GetCount(flt, &count); err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(failed_query("getting account information")).SetLogEntry(log.Fields{
			"telegid": ua.TelegID,
		})
	}
	if count == 0 {
		// When the account isnt found - this can reveal the telegid though
		return NewDomainError(ERR_ACC404, nil).SetLoc(errLoc).SetUsrMsg(account_notfound(ua.TelegID)).SetLogEntry(log.Fields{
			"telegid": ua.TelegID,
		})
	}
	info, err := iadp.GetOne(flt, reflect.TypeOf(&UserAccount{}))
	if err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(failed_query("getting account information")).SetLogEntry(log.Fields{
			"telegid": ua.TelegID,
		})
	}
	x, _ := info.(*UserAccount)
	*ua = *x
	return nil
}

// RegisterNewAccount: checks for the account duplicates, validates the account values and then just pushes that to the database
// Any account when registered new will be marked archived = false
// any account when registered new will have elevation =0
// Icnase of already registered account but archived its re-enabled
func RegisterNewAccount(ua *UserAccount, iadp dbadp.DbAdaptor) error {
	errLoc := "RegisterNewAccount"
	if iadp == nil {
		return NewDomainError(ERR_DBCONN, nil).SetLoc(errLoc).SetUsrMsg(gateway_fail())
	}
	if ua == nil {
		return NewDomainError(ERR_NILACC, nil).SetLoc(errLoc).SetUsrMsg(invalid_account("getting nil account information"))
	}
	if !REGX_EMAIL.MatchString(ua.Email) || ua.Name == "" {
		return NewDomainError(ERR_ACCFLD, nil).SetLoc(errLoc).SetUsrMsg(invalid_account("trying to register account with invalid email or name")).SetLogEntry(log.Fields{
			"email": ua.Email,
		})
	}
	duplicate := 0
	archive := false
	if err := iadp.GetCount(&UserAccount{TelegID: ua.TelegID, Archived: &archive}, &duplicate); err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(failed_query("getting account information")).SetLogEntry(log.Fields{
			"telegid":  ua.TelegID,
			"archived": archive,
		})
	}
	if duplicate != 0 {
		return NewDomainError(ERR_ACCDUPLC, nil).SetLoc(errLoc).SetUsrMsg(duplc_account(ua.TelegID, ua.Email)).SetLogEntry(log.Fields{
			"telegid":  ua.TelegID,
			"archived": archive,
		})
	}
	if err := iadp.GetCount(&UserAccount{Email: ua.Email, Archived: &archive}, &duplicate); err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(failed_query("getting account information")).SetLogEntry(log.Fields{
			"telegid":  ua.TelegID,
			"archived": archive,
		})
	}
	if duplicate != 0 {
		return NewDomainError(ERR_ACCDUPLC, nil).SetLoc(errLoc).SetUsrMsg(duplc_account(ua.TelegID, ua.Email)).SetLogEntry(log.Fields{
			"telegid":  ua.TelegID,
			"archived": archive,
		})
	}
	archived := 0
	archive = true
	if err := iadp.GetCount(&UserAccount{TelegID: ua.TelegID, Archived: &archive}, &archived); err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(failed_query("getting account information"))
	}
	if archived != 0 {
		// this means the account just needs re-enablement and not registration
		// this happens with updating the email id as well.
		archive = false
		if err := iadp.UpdateOne(&UserAccount{TelegID: ua.TelegID}, &UserAccount{Archived: &archive, Email: ua.Email}); err != nil {
			return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(duplc_account(ua.TelegID, ua.Email)).SetLogEntry(log.Fields{
				"telegid":  ua.TelegID,
				"archived": archive,
			})
		}
		return NewDomainError(ERR_ACCREENABLE, nil).SetLoc(errLoc).SetUsrMsg(acc_renable(ua.TelegID)).SetLogEntry(log.Fields{"telegid": ua.TelegID})
	}
	// defaults when registering new account
	defaultElev := AccElev(User)
	ua.Elevtn = &defaultElev
	archive = false
	ua.Archived = &archive
	if err := iadp.AddOne(ua); err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(failed_query("adding new account information")).SetLogEntry(log.Fields{
			"telegid":  ua.TelegID,
			"archived": archive,
		})
	}
	return nil
}

// ElevateAccount : comes in handy when an account has to be promoted in role
// sends back the account details after having elevated it
// Errors incase: account does not exists,requested elevation not within limits,query to update account fails,query to get the updated account details fails
func ElevateAccount(ua *UserAccount, iadp dbadp.DbAdaptor) error {
	errLoc := "ElevateAccount"
	if iadp == nil {
		return NewDomainError(ERR_DBCONN, nil).SetLoc(errLoc).SetUsrMsg(gateway_fail())
	}
	if ua == nil {
		return NewDomainError(fmt.Errorf("nil account query"), nil).SetLoc(errLoc).SetUsrMsg(invalid_account("getting nil account information"))
	}
	exists := 0
	archive := false
	if err := iadp.GetCount(&UserAccount{TelegID: ua.TelegID, Archived: &archive}, &exists); err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(failed_query("adding new account information")).SetLogEntry(log.Fields{
			"telegid": ua.TelegID,
			"elevatn": ua.Elevtn,
		})
	}
	if exists == 0 {
		return NewDomainError(ERR_ACCMISSIN, nil).SetLoc(errLoc).SetUsrMsg(account_notfound(ua.TelegID)).SetLogEntry(log.Fields{
			"telegid": ua.TelegID,
		})
	}
	if *ua.Elevtn < AccElev(User) || *ua.Elevtn > AccElev(Admin) {
		return NewDomainError(ERR_INVLPARAM, nil).SetLoc(errLoc).SetUsrMsg(elev_ceiling())
	}
	if err := iadp.UpdateOne(&UserAccount{TelegID: ua.TelegID}, &UserAccount{Elevtn: ua.Elevtn}); err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(failed_query("elevating  account role"))
	}
	updated, err := iadp.GetOne(&UserAccount{TelegID: ua.TelegID}, reflect.TypeOf(&UserAccount{}))
	if err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(failed_query("getting account information")).SetLogEntry(log.Fields{
			"telegid": ua.TelegID,
		})
	}
	// sending the updated account detatils
	// FIXME: casting this into a type - this has some problems
	// x ==nil, while updated has the expected value
	x, _ := updated.(*UserAccount)
	*ua = *x
	return nil
}

// UpdateAccountEmail: changes the email attached to the account
// Errors when account not found registered
func UpdateAccountEmail(ua *UserAccount, iadp dbadp.DbAdaptor) error {
	errLoc := "UpdateAccountEmail"
	if iadp == nil {
		return NewDomainError(ERR_DBCONN, nil).SetLoc(errLoc).SetUsrMsg(gateway_fail())
	}
	if ua == nil {
		return NewDomainError(fmt.Errorf("nil account query"), nil).SetLoc(errLoc).SetUsrMsg(invalid_account("getting nil account information"))
	} else if !REGX_EMAIL.MatchString(ua.Email) {
		return NewDomainError(ERR_INVLPARAM, nil).SetLoc(errLoc).SetUsrMsg(invalid_account("updating account with invalid email")).SetLogEntry(log.Fields{"email": ua.Email})
	}
	exists := 0
	archive := false
	err := iadp.GetCount(&UserAccount{TelegID: ua.TelegID, Archived: &archive}, &exists)
	if err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(failed_query("getting account information")).SetLogEntry(log.Fields{"telegid": ua.TelegID})
	}
	if exists == 0 {
		return NewDomainError(ERR_ACCMISSIN, nil).SetLoc(errLoc).SetUsrMsg(account_notfound(ua.TelegID)).SetLogEntry(log.Fields{"telegid": ua.TelegID})
	}
	if err := iadp.UpdateOne(&UserAccount{TelegID: ua.TelegID}, &UserAccount{Email: ua.Email}); err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(failed_query("updating account information")).SetLogEntry(log.Fields{"telegid": ua.TelegID})
	}
	updated, err := iadp.GetOne(&UserAccount{TelegID: ua.TelegID}, reflect.TypeOf(&UserAccount{}))
	if err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc("UpdateAccountEmail").SetUsrMsg(failed_query("getting account information")).SetLogEntry(log.Fields{"telegid": ua.TelegID})
	}
	x, _ := updated.(*UserAccount)
	*ua = *x
	return nil
}

// DeregisterAccount :  flagging the account as archived
// Account information is never deleted since there is financial information connected to the account
// account is marked archived only to omit it from all other searches and operations
// Throws an error incase the db query fails or the account isnt found
func DeregisterAccount(ua *UserAccount, iadp dbadp.DbAdaptor) error {
	errLoc := "DeregisterAccount"
	exists := 0
	archive := false
	err := iadp.GetCount(&UserAccount{TelegID: ua.TelegID, Archived: &archive}, &exists)
	if err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(failed_query("getting account information"))
	}
	if exists == 0 {
		return NewDomainError(ERR_ACCMISSIN, nil).SetLoc(errLoc).SetUsrMsg(account_notfound(ua.TelegID)).SetLogEntry(log.Fields{
			"telegid": ua.TelegID,
		})
	}
	archive = true
	if err := iadp.UpdateOne(&UserAccount{TelegID: ua.TelegID}, &UserAccount{Archived: &archive}); err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc("DeregisterAccount").SetUsrMsg(failed_query("de-registering account")).SetLogEntry(log.Fields{
			"telegid": ua.TelegID,
		})
	}
	return nil
}
