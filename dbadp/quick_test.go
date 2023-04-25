package dbadp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateMongoAdap(t *testing.T) {
	t.Log("creating mongo adaptor..")
	adp := NewMongoDBAdaptor(map[string]string{
		// when testing the client is dialling from the local machine
		// hence we used the mapped ports and localhost
		"hostip":   "localhost:27017",
		"database": "botmincock",
	})
	assert.NotNil(t, adp, "Unexpected nil adaptor when creating mongo connection")

	// TEST: for the server ip that isnt existing
	adp = NewMongoDBAdaptor(map[string]string{
		// when testing the client is dialling from the local machine
		// hence we used the mapped ports and localhost
		"hostip":   "mongostore:27017",
		"database": "botmincock",
	})
	assert.Nil(t, adp, "Unexpected nil adaptor when creating mongo connection")
	// TEST: for the port that isnt listening
	adp = NewMongoDBAdaptor(map[string]string{
		// when testing the client is dialling from the local machine
		// hence we used the mapped ports and localhost
		"hostip":   "localhost:27018",
		"database": "botmincock",
	})
	assert.Nil(t, adp, "Unexpected nil adaptor when creating mongo connection")
}
