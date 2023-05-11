package biz

/* ==================================
author 		: kneerunjun@gmail.com
time		: April 2023
project		: botmincock
package sends out errors from functions defined.
error needs to carry a bit more information than just error messages
====================================*/
import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// DomainError : error object from the business layer
type DomainError struct {
	Err      error      // error emanating on the location
	Internal error      // lower level error
	Loc      string     // origin of the error
	UserMsg  string     // typically a message that is fit for user consumption, not much of server details
	Entry    *log.Entry // so that we have systematic debug information
}

func (de *DomainError) Error() string {
	if de.Internal != nil {
		return fmt.Sprintf("%s %s:%s", de.Loc, de.Err, de.Internal)
	} else {
		return fmt.Sprintf("%s %s", de.Loc, de.Err)
	}
}
func (de *DomainError) SetLoc(l string) *DomainError {
	de.Loc = l
	return de
}
func (de *DomainError) SetUsrMsg(m string) *DomainError {
	de.UserMsg = m
	return de
}
func (de *DomainError) SetLogEntry(f log.Fields) *DomainError {
	de.Entry = log.WithFields(f)
	return de
}

// LogE : logs the error on the server depending debugging information available
func (de *DomainError) LogE() {
	if de.Entry == nil {
		// without debugging information
		log.Error(de.Error())
	} else {
		// with debugging information
		de.Entry.Error(de.Error())
	}
}
func NewDomainError(err, internal error) *DomainError {
	return &DomainError{
		Err:      err,
		Internal: internal,
	}
}
