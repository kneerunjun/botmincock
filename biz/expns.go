package biz

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/kneerunjun/botmincock/dbadp"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	INVL_EXPNS     = "one or more arguments of the command to record an expense is invalid. Kindly check & send again"
	FAIL_QRY_EXPNS = "one or more internal operations on expense failed, if this continues we request you to contact the admin"
	FAIL_GET_EXPNS = "failed to find expense, either an operation failed or there weren't any recorded expenses"
)

// daysInMonth: for any month this can give the utmost days in it
// https://stackoverflow.com/questions/35182556/get-last-day-in-month-of-time-time
func daysInMonth(month time.Month, year int) int {
	switch month {
	case time.April, time.June, time.September, time.November:
		return 30
	case time.February:
		// Not all years that are divisible by 4 are leap years
		// Those that are divisible by 4 but not by 100 or divisible by 400 are leap years
		if year%4 == 0 && (year%100 != 0 || year%400 == 0) { // leap year
			return 29
		}
		return 28 // not a leap year
	default:
		return 31
	}
}

// TeamMonthlyExpense: gets the aggregate of monthly expenses for the month
// ue		: in/out param, send in the id and the month for which expenses are expected, gets back with the aggregate of expenses
func TeamMonthlyExpense(ue *MnthlyExpnsQry, iadp dbadp.DbAdaptor) error {
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
		return NewDomainError(fmt.Errorf("failed query getting team monthly expenses"), err).SetLoc("TeamMonthlyExpense").SetUsrMsg(FAIL_QRY_EXPNS)
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
		return NewDomainError(fmt.Errorf("failed query getting user monthly expenses"), err).SetLoc("UserMonthlyExpense").SetUsrMsg(FAIL_QRY_EXPNS)
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
	if exp == nil {
		return NewDomainError(fmt.Errorf("invalid expense object, cannot be nil"), nil).SetLoc("RecordExpense").SetUsrMsg(INVL_EXPNS)
	}
	if exp.DtTm.IsZero() || exp.INR <= float32(0.0) {
		logrus.WithFields(logrus.Fields{
			"dttm": exp.DtTm,
			"inr":  exp.INR,
		})
		return NewDomainError(fmt.Errorf("one or more arguments for the expense are invalid"), nil).SetLoc("RecordExpense").SetUsrMsg(INVL_EXPNS)
	}
	// relational check to be done in the calling package not here
	// like checkin if the account against which expense is added is not checked here
	err := iadp.AddOne(exp)
	if err != nil {
		return NewDomainError(fmt.Errorf("failed to record expense"), err).SetLoc("RecordExpense").SetUsrMsg(FAIL_QRY_EXPNS)
	}
	/*
		Adds as a credit for the same account in the transactions as well
	*/
	transacs := iadp.Switch("transacs")
	trnsc := &Transac{TelegID: exp.TelegID, Credit: exp.INR, Desc: exp.Desc, DtTm: exp.DtTm}
	err = transacs.AddOne(trnsc)
	if err != nil {
		return NewDomainError(fmt.Errorf("failed to record transaction"), err).SetLoc("RecordExpense").SetUsrMsg(FAIL_QRY_EXPNS)
	}
	// sending back the details of the newly added expense
	// TODO: there has to be an id to identify the expense uniquly, else we would have to use 3 simulteneous fields to pick the expenses
	// BUG: to uniquely identify the expense we need to add the date of the expense too..
	newExp, err := iadp.GetOne(&Expense{TelegID: exp.TelegID, INR: exp.INR}, reflect.TypeOf(&Expense{}))
	if err != nil {
		return NewDomainError(fmt.Errorf("failed to get details of the newly added expense"), err).SetLoc("RecordExpense").SetUsrMsg(FAIL_GET_EXPNS)
	}
	x, _ := newExp.(*Expense)
	*exp = *x
	return nil
}
