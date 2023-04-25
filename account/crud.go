package account

import (
	"fmt"

	"github.com/kneerunjun/botmincock/dbadp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2/bson"
)

const(
	ERR_EMPTYRESULT = ""
)

func GetAccountofID(uid int64, mngAdp dbadp.DBAdaptor) (*UserAccount, error) {
	result, err := mngAdp.GetOne(bson.M{"tid": uid, "archive": false})
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("GetAccountofID: query failed")
		return nil, fmt.Errorf("failed to get Account")
	}
	byt, _ := bson.Marshal(result)
	ua := UserAccount{}
	err = bson.Unmarshal(byt, &ua)
	if err != nil {
		return nil, fmt.Errorf(ERR_EMPTYRESULT)
	}
	return nil, nil
}

func GetAccountOfEmail(email string, mngAdp dbadp.DBAdaptor) (*UserAccount, error) {
	return nil, nil
}

func EditAccount(uid int64, mngAdp dbadp.DBAdaptor) (*UserAccount, error) {
	return nil, nil
}

func PostAccount(uid int64, name, email string, mngAdp dbadp.DBAdaptor) error {
	return nil
}
