package biz

import (
	"time"

	"github.com/kneerunjun/botmincock/dbadp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

// ClearDues : adds a simple credit transaction
// date of the transaction has to be the time when you have added it
// all transactions are aggregated for the month - if you are recording ofsetted transaction make sure its for the same month
func ClearDues(tr *Transac, iadp dbadp.DbAdaptor) error {
	errLoc := "ClearDues"
	if iadp == nil {
		return NewDomainError(ERR_DBCONN, nil).SetLoc(errLoc).SetUsrMsg(gateway_fail())
	}
	// checking to see if the account is registered
	accsColl := iadp.Switch("accounts")
	ua := &UserAccount{TelegID: tr.TelegID}
	err := AccountInfo(ua, accsColl)
	if err != nil {
		return NewDomainError(ERR_ACC404, err).SetLoc(errLoc).SetUsrMsg(account_notfound(ua.TelegID)).SetLogEntry(log.Fields{
			"telegid": ua.TelegID,
		})
	}
	// account is registered, we can now proceed to add transaction
	err = iadp.AddOne(tr)
	if err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(failed_query("adding a new transaction")).SetLogEntry(log.Fields{
			"cr":      tr.Credit,
			"dr":      tr.Debit,
			"telegid": tr.TelegID,
		})
	}
	return nil
}

// MyDues: aggregate of all credits and debits on a particular account balanced for either outgoing or incoming
// bl		: in/out object for result and param of query
// iadp		: database or any persistent storage adaptor
func MyDues(bl *Balance, iadp dbadp.DbAdaptor) error {

	errLoc := "MyDues"
	if iadp == nil {
		return NewDomainError(ERR_DBCONN, nil).SetLoc(errLoc).SetUsrMsg(gateway_fail())
	}
	// checking to see if the account is registered
	accsColl := iadp.Switch("accounts")
	ua := &UserAccount{TelegID: bl.TelegID}
	err := AccountInfo(ua, accsColl)
	if err != nil {
		return NewDomainError(ERR_ACC404, err).SetLoc(errLoc).SetUsrMsg(account_notfound(ua.TelegID)).SetLogEntry(log.Fields{
			"telegid": ua.TelegID,
		})
	}
	// account is registered, we can now proceed to add transaction
	temp := bl.DtTm
	yr := temp.Year()
	mn := temp.Month()
	loc := temp.Location()
	fromDt := time.Date(yr, mn, 1, 0, 0, 0, 0, loc)                 //start date for any month is 1
	toDt := time.Date(yr, mn, daysInMonth(mn, yr), 0, 0, 0, 0, loc) // end date for a month needs some extra calc
	match := bson.M{
		"$match": bson.M{
			"tid": bl.TelegID,
			"dttm": bson.M{
				"$gte": fromDt,
				"$lte": toDt,
			},
		},
	} //matching only those transactions for the user and current month
	group := bson.M{
		"$group": bson.M{
			"_id":     nil,
			"debits":  bson.M{"$sum": "$debit"},
			"credits": bson.M{"$sum": "$credit"},
			"tid":     bson.M{"$first": "$tid"},
			"dttm":    bson.M{"$first": "$dttm"},
		},
	} // aggregating credits and debits for the account
	project := bson.M{
		"$project": bson.M{
			"_id":  0,
			"tid":  1,
			"dttm": 1,
			"due": bson.M{
				"$subtract": []interface{}{"$credits", "$debits"},
			},
		},
	} // projecting that to the balance object
	err = iadp.Aggregate([]bson.M{match, group, project}, bl)
	if err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(failed_query("getting aggregate transactions for account"))
	}
	return nil
}

// MarkPlayday: For every day that a player sends a certain message as GM - we expect the bot to insert new debit transaction
// tr 		: transaction object that shall determine the date, telegid of the transaction. INR value of the transaction is determined by the playshare and the expenses
func MarkPlayday(tr *Transac, iadp dbadp.DbAdaptor) error {
	/*====================
	incase no connection to datbase
	incase the account does not exists - playday cannot be marked
	====================*/
	errLoc := "MarkPlayday"
	if iadp == nil {
		return NewDomainError(ERR_DBCONN, nil).SetLoc(errLoc).SetUsrMsg(gateway_fail())
	}
	accsColl := iadp.Switch("accounts")
	ua := &UserAccount{TelegID: tr.TelegID}
	err := AccountInfo(ua, accsColl)
	if err != nil {
		return NewDomainError(ERR_ACC404, err).SetLoc(errLoc).SetUsrMsg(account_notfound(ua.TelegID)).SetLogEntry(log.Fields{
			"telegid": ua.TelegID,
		})
	}
	// TODO: find if duplicate transaction
	result := struct {
		Count int `bson:"count"`
	}{}
	fromDt, toDt := TodayAsBoundary()
	iadp.Aggregate([]bson.M{
		{"$match": bson.M{"tid": tr.TelegID, "desc": PLAYDAY_DESC, "dttm": bson.M{
			"$gte": fromDt,
			"$lte": toDt,
		}}},
		{"$group": bson.M{
			"_id":   nil,
			"count": bson.M{"$sum": 1},
		}},
		{"$project": bson.M{
			"_id": 0,
		}},
	}, &result)
	if result.Count > 0 {
		return NewDomainError(ERR_DUPLTRANSAC, nil).SetLoc(errLoc).SetUsrMsg("You seem to have already marked your attendance. Cannot debit more than once").SetLogEntry(log.Fields{
			"telegid": ua.TelegID,
		})
	}
	// NOTE: transaction date cannot be changed here - since the playday is marked once for any given day
	// While the user might have sent the message on a day, its possible that the bot reads it a day later.
	// The message date is the date of the transaction
	// NOTE: for the debit value the estimates play a role in calculations
	tr.Debit = float32(100.00) // for now we just pegg the debit value to 100, but it must be deduced from the estimates and the expenses
	err = iadp.AddOne(tr)
	if err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(FAIL_QRY_EXPNS).SetLogEntry(log.Fields{
			"inr":     tr.Debit,
			"dt":      tr.DtTm,
			"telegid": tr.TelegID,
		})
	}
	return nil
}
