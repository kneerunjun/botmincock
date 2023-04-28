package dbadp

import "reflect"

// DbAdaptor : Agnostic of the database platform this can
type DbAdaptor interface {
	AddOne(interface{}) error
	RemoveOne(interface{}) error
	UpdateOne(interface{}) error
	GetOne(interface{}, reflect.Type) (interface{}, error)
	GetCount(interface{}, *int) error
}
