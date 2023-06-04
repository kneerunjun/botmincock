package dbadp

/* ==================================
author 		: kneerunjun@gmail.com
time		: April 2023
project		: botmincock
For the purpose of usint testing other packages that want to be agnostic of the logic thats implemented here on the database adaptor we implement a facade of a adaptor
====================================*/
import (
	"reflect"

	"gopkg.in/mgo.v2/bson"
)

// DbAdaptor : Agnostic of the database platform this can
// basic functions any database adaptor needs to implement
type DbAdaptor interface {
	AddOne(interface{}) error
	RemoveOne(interface{}) error
	UpdateOne(interface{}, interface{}) error
	UpdateBulk(selectr, patch interface{}) (int, error)
	GetOne(interface{}, reflect.Type) (interface{}, error)
	GetCount(interface{}, *int) error
	Aggregate(p []bson.M, res interface{}) error
	Switch(string) DbAdaptor // switches the collection and sends out a new adaptor with new underlying collection
}
