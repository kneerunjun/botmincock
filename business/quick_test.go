package business

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailRegx(t *testing.T) {
	data := []string{
		"bdensumbe0@youku.com",
		"mroth3@soup.io",
		"rlambregts4@godaddy.com",
		"jhucker.be9@virginia.edu",
		"lgh_iona@newyorker.com",
		"kneerunjun@gmail.com",
	}
	dataNotOk := []string{
		"ahubatsch8@bbc.co.uk",
		"fpetrazzib@state.tx.us",
		"@state.tx.us",
		"fpetrazzibstate.us",
		"#fpetrazzib@state.us",
	}
	for _, d := range data {
		assert.True(t, REGX_EMAIL.MatchString(d), fmt.Sprintf("email %s should have passed the test", d))
	}
	for _, d := range dataNotOk {
		assert.False(t, REGX_EMAIL.MatchString(d), fmt.Sprintf("email %s shouldn't have passed the test", d))
	}
}
