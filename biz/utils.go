package biz

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	INVL_EXPNS     = "one or more arguments of the command to record an expense is invalid. Kindly check & send again"
	FAIL_QRY_EXPNS = "one or more internal operations on expense failed, if this continues we request you to contact the admin"
	FAIL_GET_EXPNS = "failed to find expense, either an operation failed or there weren't any recorded expenses"
	TRY_AGAIN      = "Yikes! your command did not work as expected.Try to run again after a while,or if the problem persists kindly contact a sys-admin."
	DUPL_ACC       = "Account with same ID or email is already registered & active. Cannot register again."
	ACC_REACTIVE   = "Account with same ID found archived. Reactivated the account."
	ACC_MISSN      = "Account you referring to does not exists. Please register first then try again."
	INVLD_PARAM    = "One/more params for the command supplied is invalid, Please consult a sys-admin and then try again."
)

/*====================
a list of commonly used emoticons in telegram as variables
- to add new emoticons : get the unicode from given reference url
- remove the `+` symbol
- convert to 32 bit int after having trimmed the leading \\U
- TrimPrefix looks superflous, but helps quickly copy more codes from the site below
====================*/

var (
	// https://unicode.org/emoji/charts/full-emoji-list.html
	EMOJI_warning, _   = strconv.ParseInt(strings.TrimPrefix("\\U26A0", "\\U"), 16, 32)
	EMOJI_grinface, _  = strconv.ParseInt(strings.TrimPrefix("\\U1F600", "\\U"), 16, 32)
	EMOJI_rofl, _      = strconv.ParseInt(strings.TrimPrefix("\\U1F923", "\\U"), 16, 32)
	EMOJI_redcross, _  = strconv.ParseInt(strings.TrimPrefix("\\U274C", "\\U"), 16, 32)
	EMOJI_redqs, _     = strconv.ParseInt(strings.TrimPrefix("\\U2753 	", "\\U"), 16, 32)
	EMOJI_bikini, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F459", "\\U"), 16, 32)
	EMOJI_greentick, _ = strconv.ParseInt(strings.TrimPrefix("\\U2705", "\\U"), 16, 32)
	EMOJI_clover, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F340", "\\U"), 16, 32)
	EMOJI_meat, _      = strconv.ParseInt(strings.TrimPrefix("\\U1F357", "\\U"), 16, 32)
	EMOJI_robot, _     = strconv.ParseInt(strings.TrimPrefix("\\U1F916", "\\U"), 16, 32)
	EMOJI_copyrt, _    = strconv.ParseInt(strings.TrimPrefix("\\U00A9", "\\U"), 16, 32)
	EMOJI_banana, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F34C", "\\U"), 16, 32)
	EMOJI_garlic, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F9C4", "\\U"), 16, 32)
	EMOJI_email, _     = strconv.ParseInt(strings.TrimPrefix("\\U1F4E7", "\\U"), 16, 32)
	EMOJI_badge, _     = strconv.ParseInt(strings.TrimPrefix("\\U1FAAA", "\\U"), 16, 32)
	EMOJI_sheild, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F6E1", "\\U"), 16, 32)
	EMOJI_recycle, _   = strconv.ParseInt(strings.TrimPrefix("\\U267B", "\\U"), 16, 32)
	EMOJI_wilted, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F940", "\\U"), 16, 32)

	REGX_EMAIL = regexp.MustCompile(`^[\w\d._]+@[\w]+.[\w\d]+$`)
)

/*====================
Utility functions can help create contextual user messages from within the business layer.
It shall they apply the emoticons as required
Context of the error still comes from the biz functions - here we have all the common text on the error messgaes
====================*/

func failed_query(operation string) string {
	return fmt.Sprintf("%c Internal operation: '%s' failed, try after some time%%0AIf this continues you may have to contact an administrator", EMOJI_wilted, operation)
}
func account_notfound(id int64) string {
	return fmt.Sprintf("%c No account found associated with the ID %d%%0A Use /registerme command to register first & then proceed", EMOJI_warning, id)
}
func invalid_account(reason string) string {
	return fmt.Sprintf("%c There was a problem: %s", EMOJI_warning, reason)
}
func duplc_account(id int64, email string) string {
	return fmt.Sprintf("%c Account with ID %d or Email %s is already registered", EMOJI_redcross, id, email)
}

func acc_renable(id int64) string {
	return fmt.Sprintf("%c account with ID %d was found archived, re-enabled it", EMOJI_recycle, id)
}

func elev_ceiling() string {
	return fmt.Sprintf("%c Account is at the highest elevation, cannot elevate any further", EMOJI_redcross)
}
func gateway_fail() string {
	return fmt.Sprintf("%c A gateway has failed, and hence aborting your command for now. Check with an admin", EMOJI_redcross)
}

/*====================
error forming utility functions so that standardised errors are sent across the board when logged
Error messages should be standardised
====================*/

var (
	ERR_DBCONN      = fmt.Errorf("no adaptor connection, check if database is up")
	ERR_NILACC      = fmt.Errorf("trying to query with nil account")
	ERR_QRYFAIL     = fmt.Errorf("failed database query")
	ERR_ACC404      = fmt.Errorf("account not found")
	ERR_ACCFLD      = fmt.Errorf("one/more fields of the account is invalid")
	ERR_ACCDUPLC    = fmt.Errorf("duplicate account")
	ERR_ACCREENABLE = fmt.Errorf("account found archived, now re-enabled")
	ERR_ACCMISSIN   = fmt.Errorf("account not found")
	ERR_INVLPARAM   = fmt.Errorf("one or more params is invalid")
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