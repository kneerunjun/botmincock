package dbadp

/* ==================================
author 		: kneerunjun@gmail.com
time		: April 2023
project		: botmincock
For the purpose of usint testing other packages that want to be agnostic of the logic thats implemented here on the database adaptor we implement a facade of a adaptor
====================================*/
import (
	"reflect"
	"time"

	"gopkg.in/mgo.v2"
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
	return ma.Update(selectr, patch)
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
