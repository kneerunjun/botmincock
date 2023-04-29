package dbadp

/* ==================================
author 		: kneerunjun@gmail.com
time		: April 2023
project		: botmincock
- Adaptors specific implementation for mongo as a service
====================================*/
import (
	"reflect"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// mongoAdaptor : database adaptor for mongo specific implementation
type mongoAdaptor struct {
	*mgo.Collection
}

func (ma *mongoAdaptor) AddOne(m interface{}) error {
	return ma.Insert(m)
}

func (ma *mongoAdaptor) RemoveOne(m interface{}) error {
	return ma.Remove(m)
}

func (ma *mongoAdaptor) UpdateOne(selectr, patch interface{}) error {
	return ma.Update(selectr, bson.M{"$set": patch})
}

func (ma *mongoAdaptor) GetOne(m interface{}, t reflect.Type) (interface{}, error) {
	result := reflect.New(t.Elem()).Interface()
	err := ma.Find(m).One(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ma *mongoAdaptor) GetCount(m interface{}, c *int) error {
	// byt, _ := bson.Marshal(m)
	// flt := bson.M{}
	// bson.Unmarshal(byt, &flt)
	count, err := ma.Find(m).Count()
	if err != nil {
		return err
	}
	*c = count
	return nil
}

func NewMongoAdpator(ipport, database, coll string) DbAdaptor {
	sess, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{ipport},
		Timeout:  4 * time.Second,
		Database: database,
	})
	if err != nil || sess == nil {
		return nil
	}
	return &mongoAdaptor{Collection: sess.DB("").C(coll)}
}
