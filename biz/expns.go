package biz

import (
	"fmt"
	"reflect"

	"github.com/kneerunjun/botmincock/dbadp"
	"github.com/sirupsen/logrus"
)

// RecordExpense : recrods a new expense in the database
// Expenses with default date time , and zero value invalid expenses
// any user can record expenses and the telegram id of the sender is considered to be the one expending
// recording expenses on behalf of another user is not possible
func RecordExpense(exp *Expense, iadp dbadp.DbAdaptor) error {
	if exp == nil {
		return NewDomainError(fmt.Errorf("invalid expense object, cannot be nil"), nil)
	}
	if exp.DtTm.IsZero() || exp.INR <= float32(0.0) {
		logrus.WithFields(logrus.Fields{
			"dttm": exp.DtTm,
			"inr":  exp.INR,
		})
		return NewDomainError(fmt.Errorf("one or more arguments for the expense are invalid"), nil)
	}
	// relational check to be done in the calling package not here
	// like checkin if the account against which expense is added is not checked here
	err := iadp.AddOne(exp)
	if err != nil {
		return NewDomainError(fmt.Errorf("failed to add expense"), err)
	}
	// sending back the details of the newly added expense
	// TODO: there has to be an id to identify the expense uniquly, else we would have to use 3 simulteneous fields to pick the expenses
	newExp, err := iadp.GetOne(&Expense{TelegID: exp.TelegID, INR: exp.INR}, reflect.TypeOf(&Expense{}))
	if err != nil {
		return NewDomainError(fmt.Errorf("failed to get details of the newly added expense"), err)
	}
	x, _ := newExp.(*Expense)
	*exp = *x
	return nil
}
