package dbadp

import (
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type mongoDbAdaptor struct {
	*mgo.Session
}

func (mda *mongoDbAdaptor) GetOne(m bson.M, db, coll string) (bson.M, error) {
	result := bson.M{}
	if err := mda.DB(db).C(coll).Find(m).One(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// NewMongoDBAdaptor :given the configuration this can dialup a mongo connection
// when in error will send a nil adaptor - indicating the connection failed
//
/*
	adp := NewMongoDBAdaptor(map[string]string{"hostip": "localhost:27017", "database":"dbanme"})
	if adp !=nil{
		fmt.Errorf("failed to create new adaptor")
	}
*/
func NewMongoDBAdaptor(cfg map[string]string) DBAdaptor {
	// dial the mongo connection given the configuration
	// a default collection is used here
	dlInf := &mgo.DialInfo{
		Addrs:    []string{cfg["hostip"]},
		Database: cfg["database"],
		Timeout:  5 * time.Second,
	}
	sess, err := mgo.DialWithInfo(dlInf)
	if err != nil || sess == nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Failed to dial mongo session")
		return nil
	}
	return &mongoDbAdaptor{sess}
}
