package biz

import (
	"errors"
	"reflect"
	"time"

	"github.com/kneerunjun/botmincock/dbadp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// TeamMonthlyExpense: gets the aggregate of monthly expenses for the month
// ue		: in/out param, send in the id and the month for which expenses are expected, gets back with the aggregate of expenses
func TeamMonthlyExpense(ue *MnthlyExpnsQry, iadp dbadp.DbAdaptor) error {
	errLoc := "TeamMonthlyExpense"
	temp := ue.Dttm
	yr := temp.Year()
	mn := temp.Month()
	loc := temp.Location()
	fromDt := time.Date(yr, mn, 1, 0, 0, 0, 0, loc)                 //start date for any month is 1
	toDt := time.Date(yr, mn, daysInMonth(mn, yr), 0, 0, 0, 0, loc) // end date for a month needs some extra calc
	pipe := []bson.M{
		{"$match": bson.M{"dttm": bson.M{"$gte": fromDt, "$lte": toDt}}}, // only matching all the expenses for the month
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$inr"}}},
		{"$project": bson.M{"_id": 0}},
	}
	err := iadp.Aggregate(pipe, ue)
	if err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			ue.Total = float32(0.0) // since no expenses found for matching i
			return nil
		}
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(FAIL_QRY_EXPNS).SetLogEntry(log.Fields{
			"inr": ue.Total,
			"dt":  ue.Dttm,
		})
	}
	ue.Dttm = temp // setting the date back to what it was
	return nil
}

// UserMonthlyExpense: for the current month this will get sum of all expenses for the
// ue		: in/out param, send in the id and the month for which expenses are expected, gets back with the aggregate of expenses
func UserMonthlyExpense(ue *MnthlyExpnsQry, iadp dbadp.DbAdaptor) error {
	temp := ue.Dttm
	yr := temp.Year()
	mn := temp.Month()
	loc := temp.Location()
	fromDt := time.Date(yr, mn, 1, 0, 0, 0, 0, loc)                 //start date for any month is 1
	toDt := time.Date(yr, mn, daysInMonth(mn, yr), 0, 0, 0, 0, loc) // end date for a month needs some extra calc
	pipe := []bson.M{
		{"$match": bson.M{"tid": ue.TelegID, "dttm": bson.M{"$gte": fromDt, "$lte": toDt}}}, // specific user current month
		{"$group": bson.M{"_id": nil, "total": bson.M{"$sum": "$inr"}, "tid": bson.M{"$first": "$tid"}}},
		{"$project": bson.M{"_id": 0}},
	}
	err := iadp.Aggregate(pipe, ue)
	if err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			ue.Total = float32(0.0) // since no expenses found for matching i
			return nil
		}
		return NewDomainError(ERR_QRYFAIL, err).SetLoc("UserMonthlyExpense").SetUsrMsg(FAIL_QRY_EXPNS).SetLogEntry(log.Fields{
			"inr":     ue.Total,
			"dt":      ue.Dttm,
			"telegid": ue.TelegID,
		})
	}
	// since on the way back the data on object ptr is erased we have to set the month again
	ue.Dttm = temp
	return nil
}

// RecordExpense : recrods a new expense in the database
// Expenses with default date time , and zero value invalid expenses
// any user can record expenses and the telegram id of the sender is considered to be the one expending
// recording expenses on behalf of another user is not possible
func RecordExpense(exp *Expense, iadp dbadp.DbAdaptor) error {
	errLoc := "RecordExpense"
	if exp == nil {
		return NewDomainError(ERR_INVLPARAM, nil).SetLoc(errLoc).SetUsrMsg(INVL_EXPNS)
	}
	if exp.DtTm.IsZero() || exp.INR <= float32(0.0) {
		return NewDomainError(ERR_INVLPARAM, nil).SetLoc(errLoc).SetUsrMsg(INVL_EXPNS).SetLogEntry(log.Fields{
			"inr":     exp.INR,
			"dt":      exp.DtTm,
			"telegid": exp.TelegID,
		})
	}
	// relational check to be done in the calling package not here
	// like checkin if the account against which expense is added is not checked here
	err := iadp.AddOne(exp)
	if err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(FAIL_QRY_EXPNS).SetLogEntry(log.Fields{
			"inr":     exp.INR,
			"dt":      exp.DtTm,
			"telegid": exp.TelegID,
		})
	}
	/*
		Adds as a credit for the same account in the transactions as well
	*/
	transacs := iadp.Switch("transacs")
	trnsc := &Transac{TelegID: exp.TelegID, Credit: exp.INR, Desc: exp.Desc, DtTm: exp.DtTm}
	err = transacs.AddOne(trnsc)
	if err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(FAIL_QRY_EXPNS).SetLogEntry(log.Fields{
			"inr":     exp.INR,
			"dt":      exp.DtTm,
			"telegid": exp.TelegID,
		})
	}
	// sending back the details of the newly added expense
	// TODO: there has to be an id to identify the expense uniquly, else we would have to use 3 simulteneous fields to pick the expenses
	// BUG: to uniquely identify the expense we need to add the date of the expense too..
	newExp, err := iadp.GetOne(&Expense{TelegID: exp.TelegID, INR: exp.INR}, reflect.TypeOf(&Expense{}))
	if err != nil {
		return NewDomainError(ERR_QRYFAIL, err).SetLoc(errLoc).SetUsrMsg(FAIL_GET_EXPNS).SetLogEntry(log.Fields{
			"inr":     exp.INR,
			"dt":      exp.DtTm,
			"telegid": exp.TelegID,
		})
	}
	x, _ := newExp.(*Expense)
	*exp = *x
	return nil
}
