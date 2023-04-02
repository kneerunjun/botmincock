package botcore

import "testing"

func TestMonthlyPoll(t *testing.T) {
	// it all starts with a montly poll being sent to the group
	// a trigger from either a cron can ask the bot to send the poll
	// after a set timeinterval the bot will also collect the results for the poll - 24 hours duration
}
