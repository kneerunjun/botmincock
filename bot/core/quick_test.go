package core

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTelegBot(t *testing.T) {
	t.Setenv("BOT_HANDLE", "fdsfdsf")
	t.Setenv("BOT_NAME", "fsdfdsf")
	t.Setenv("PSABADMIN_GRP", "fsfsdfs")
	t.Setenv("MYID", "456546")
	botmincock := NewTeleGBot(&SharedExpensesBotEnv{BotEnv: EnvOrDefaults(), GuestCharges: 150.00}, reflect.TypeOf(&SharedExpensesBot{}))
	assert.NotNil(t, botmincock, "Uexpected nil when creating new bot")
}
