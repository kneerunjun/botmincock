package biz

import (
	"errors"
	"time"

	"github.com/kneerunjun/botmincock/dbadp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
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

// IsPlayMarkedToday : for the given Telegram id this can find from the transactions if the player has already marked his attendance
// returns error if the query fails
// also returns an error when player found attended, check for ERR_DUPLTRANSAC for knowing what type of error it is
func IsPlayMarkedToday(iadp dbadp.DbAdaptor, tid int64) (bool, error) {
	errLoc := "IsPlayMarkedToday"
	result := struct {
		Count int `bson:"count"`
	}{}
	fromDt, toDt := TodayAsBoundary()
	err := iadp.Aggregate([]bson.M{
		{"$match": bson.M{"tid": tid, "desc": PLAYDAY_DESC, "dttm": bson.M{
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
	if err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			return false, nil
		}
		return false, NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(failed_query("getting user's attendance"))
	}
	return result.Count > 0, NewDomainError(ERR_DUPLTRANSAC, err).SetLoc(errLoc).SetUsrMsg(duplc_attend())
}

// RecoveryTillNow : gets the monthly debits for the entire team till date only for the playday
// error in case the query to database fails
// this gives us the recovered funds till now
func RecoveryTillNow(iadp dbadp.DbAdaptor, total *float32) error {
	errLoc := "RecoveryTillNow"
	result := struct {
		TotalDebits float32 `bson:"debits"`
	}{}
	from, to := YdayAsBoundary() // all the play debits only till yesterday
	if from.IsZero() || to.IsZero() {
		// incase day today is first of any month recovery would be zero
		// we need not consider any rollover from pervious month
		*total = 0.0
		return nil
	}
	err := iadp.Aggregate([]bson.M{
		{"$match": bson.M{
			"desc": PLAYDAY_DESC,
			"dttm": bson.M{
				"$gte": from,
				"$lte": to,
			},
		}}, // all the transactions for the month marked as playday
		{"$group": bson.M{
			"_id":    0,
			"debits": bson.M{"$sum": "$debit"},
		}}, // summing the debits of all such transactions
	}, &result)
	if err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(failed_query("getting total debits for play day"))
	}
	*total = result.TotalDebits
	return nil
}

// MarkPlayday: For every day that a player sends a certain message as GM - we expect the bot to insert new debit transaction
// tr 		: transaction object that shall determine the date, telegid of the transaction. INR value of the transaction is determined by the playshare and the expenses
func MarkPlayday(tr *Transac, iadp dbadp.DbAdaptor) error {
	errLoc := "MarkPlayday"
	if iadp == nil {
		return NewDomainError(ERR_DBCONN, nil).SetLoc(errLoc).SetUsrMsg(gateway_fail())
	}
	err := iadp.AddOne(tr)
	if err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(FAIL_QRY_EXPNS).SetLogEntry(log.Fields{
			"inr":     tr.Debit,
			"dt":      tr.DtTm,
			"telegid": tr.TelegID,
		})
	}
	return nil
}

// TotalPlaydayDebits : for the given time span, all the debits for playday debits
// Incase there arent any play day debits in the given span then does NOT error but sends back the Debits =0
// Error when query encounters errors
func TotalPlaydayDebits(trq *TransacQ, iadp dbadp.DbAdaptor) error {
	errLoc := "RecoveryToday"
	if iadp == nil {
		return NewDomainError(ERR_DBCONN, nil).SetLoc(errLoc).SetUsrMsg(gateway_fail())
	}
	result := struct {
		TotalDebits float32 `bson:"debits"`
	}{}
	err := iadp.Aggregate([]bson.M{
		{"$match": bson.M{
			"desc": PLAYDAY_DESC,
			"dttm": bson.M{
				"$gte": trq.From,
				"$lte": trq.To,
			},
		}}, // all the transactions for the month marked as playday
		{"$group": bson.M{
			"_id":    0,
			"debits": bson.M{"$sum": "$debit"},
		}}, // summing the debits of all such transactions
	}, &result)
	if err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			trq.Debits = 0.0
			return nil
		}
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(FAIL_QRY_EXPNS)
	}
	trq.Debits = result.TotalDebits
	return nil
}

// AttendedToday : gets the total number of attendees for today
// Error only when the query fails
func AttendedToday(iadp dbadp.DbAdaptor) (int, error) {
	errLoc := "AttendedToday"
	if iadp == nil {
		return 0, NewDomainError(ERR_DBCONN, nil).SetLoc(errLoc).SetUsrMsg(gateway_fail())
	}
	result := struct {
		Count int `bson:"total"`
	}{}
	from, to := TodayAsBoundary()
	err := iadp.Aggregate([]bson.M{
		{"$match": bson.M{
			"desc": PLAYDAY_DESC,
			"dttm": bson.M{
				"$gte": from,
				"$lte": to,
			},
		}}, // all the transactions for the month marked as playday
		{"$group": bson.M{
			"_id":   0,
			"total": bson.M{"$sum": 1},
		}}, // summing the debits of all such transactions
	}, &result)
	if err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			return 0, nil
		}
		return 0, NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(FAIL_QRY_EXPNS)
	}
	return result.Count, nil
}

// AdjustDayDebit : for the given adjustment this will find all the day debits and update their debits
func AdjustDayDebit(trq *TransacQ, iadp dbadp.DbAdaptor) error {
	errLoc := "AdjustDayDebit"
	if iadp == nil {
		return NewDomainError(ERR_DBCONN, nil).SetLoc(errLoc).SetUsrMsg(gateway_fail())
	}
	slctr := bson.M{
		"desc": PLAYDAY_DESC,
		"dttm": bson.M{
			"$gte": trq.From,
			"$lte": trq.To,
		},
	}
	return func() error {
		patch := bson.M{
			"$inc": bson.M{"debit": trq.Debits},
		}
		if _, err := iadp.UpdateBulk(slctr, patch); err != nil {
			return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(FAIL_QRY_EXPNS)
		}
		return nil
	}() // distributing the deficits equally among all the attendees
}
