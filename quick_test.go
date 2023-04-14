package main

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegMeCommand(t *testing.T) {
	// When the user would want to register himself -
	regx := regexp.MustCompile(`^@psabadminton_bot(\s+)\/(?P<cmd>registerme)(\s+)(?P<email>[\w\d._]+@[\w]+.[\w\d]+)+$`)
	okdata := []string{
		"@psabadminton_bot /registerme kneerunjun@gmail.com",
		"@psabadminton_bot /registerme kneerunjun@gmail.co",
		"@psabadminton_bot /registerme someone@yahoo.co",
		"@psabadminton_bot /registerme someone@yahoo.co1",
		"@psabadminton_bot /registerme someon_1e@yahoo.co1",
		"@psabadminton_bot /registerme som.eon_1e@yahoo.co1",
	}
	for _, d := range okdata {
		yes := regx.MatchString(d)
		assert.Equal(t, true, yes, "Unexpected failure in recognising command pattern")
		matches := regx.FindStringSubmatch(d)
		result := map[string]string{}
		for i, name := range regx.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = matches[i]
			}
		}
		//and here we can check the result
		t.Log(result)
	}

}

// TestParseCommand : parse command is crucial when it comes to translating textual command to a command in botmincock
func TestParseCommand(t *testing.T) {
	// after the bot update has been recognised as a command it may be due to parse the command
	// here we test such commands on the bot
	t.Setenv("BOT_HANDLE", "@psabadminton_bot") // only for the purposes of testing
	testOkData := []string{
		"@psabadminton_bot /registerme kneerunjun@gmail.com",
		"@psabadminton_bot /registerme kneerunjun@gmail.co",
		"@psabadminton_bot /registerme someone@yahoo.co",
		"@psabadminton_bot /registerme someone@yahoo.co1",
		"@psabadminton_bot /registerme someon_1e@yahoo.co1",
		"@psabadminton_bot /registerme som.eon_1e@yahoo.co1",
	}
	t.Log("now testing all postive test cases ")
	for _, d := range testOkData {
		t.Logf("Now testing for the message \n %s", d)
		updt := map[string]interface{}{
			"id": 343435,
			"message": map[string]interface{}{
				"id":   43535,
				"text": d,
				"from": map[string]interface{}{
					"id":         543676857,
					"username":   "NiruAwati",
					"first_name": "Niranjan",
					"last_name":  "Awati",
				},
				"chat": map[string]interface{}{
					"id": 100923298,
				},
			},
		}
		byt, _ := json.Marshal(updt)
		bupdt := BotUpdate{}
		json.Unmarshal(byt, &bupdt)
		cmd, err := ParseBotCmd(bupdt)
		assert.Nil(t, err, "Unexpected error when parsing bot commnd")
		assert.NotNil(t, cmd, "Unexpected nil command")

	}
	// ==============
	// negative test data parsing such a command shall lead to error
	// nil botcommand returned
	// ===============
	testErrData := []string{
		"/registerme kneerunjun@gmail.com",             // missing bot callout
		"@psabadminton_bot /registerme",                // missing email for registration
		"@psabadminton_bot /register someone@yahoo.co", // command mispelt
	}
	t.Log("Now negative testing.. ")
	for _, d := range testErrData {
		t.Logf("Now testing for the message \n %s", d)
		updt := map[string]interface{}{
			"id": 343435,
			"message": map[string]interface{}{
				"id":   43535,
				"text": d,
				"from": map[string]interface{}{
					"id":         543676857,
					"username":   "NiruAwati",
					"first_name": "Niranjan",
					"last_name":  "Awati",
				},
				"chat": map[string]interface{}{
					"id": 100923298,
				},
			},
		}
		byt, _ := json.Marshal(updt)
		bupdt := BotUpdate{}
		json.Unmarshal(byt, &bupdt)
		_, err := ParseBotCmd(bupdt)
		assert.NotNil(t, err, "Unexpected nil error when parsing invalid command")
	}
	t.Cleanup(func() {
		t.Log("Exiting the test")
	})
}

func TestGMText(t *testing.T) {
	regx := regexp.MustCompile(`^(?P<cmd>(?i)(good[\s]*morning))$`)
	yes := regx.MatchString("good morning")
	assert.Equal(t, true, yes, "Unexpected failure in recognising command pattern")
	yes = regx.MatchString("good Morning")
	assert.Equal(t, true, yes, "Unexpected failure in recognising command pattern")
	yes = regx.MatchString("goodMorning")
	assert.Equal(t, true, yes, "Unexpected failure in recognising command pattern")
}
