package dbadp

/* ==================================
author 		: kneerunjun@gmail.com
time		: April 2023
project		: botmincock
For the purpose of usint testing other packages that want to be agnostic of the logic thats implemented here on the database adaptor we implement a facade of a adaptor
====================================*/
import "reflect"

// DbAdaptor : Agnostic of the database platform this can
// basic functions any database adaptor needs to implement
type DbAdaptor interface {
	AddOne(interface{}) error
	RemoveOne(interface{}) error
	UpdateOne(interface{}, interface{}) error
	GetOne(interface{}, reflect.Type) (interface{}, error)
	GetCount(interface{}, *int) error
}
