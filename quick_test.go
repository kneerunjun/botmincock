package main

import (
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

func TestGMText(t *testing.T) {
	regx := regexp.MustCompile(`^(?P<cmd>(?i)(good[\s]*morning))$`)
	yes := regx.MatchString("good morning")
	assert.Equal(t, true, yes, "Unexpected failure in recognising command pattern")
	yes = regx.MatchString("good Morning")
	assert.Equal(t, true, yes, "Unexpected failure in recognising command pattern")
	yes = regx.MatchString("goodMorning")
	assert.Equal(t, true, yes, "Unexpected failure in recognising command pattern")
}
