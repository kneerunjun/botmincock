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

type DummyAdaptor struct {
	DummyCount  int
	AddError    error
	RemoveError error
	UpdateError error
	GetOneError error
}

func (da *DummyAdaptor) UpdateBulk(selectr, patch interface{}) (int, error) {
	return 0.0, nil
}

func (da *DummyAdaptor) AddOne(interface{}) error {
	return nil
}
func (da *DummyAdaptor) RemoveOne(interface{}) error {
	return nil
}
func (da *DummyAdaptor) UpdateOne(interface{}, interface{}) error {
	return nil
}
func (da *DummyAdaptor) GetOne(interface{}, reflect.Type) (interface{}, error) {
	return nil, nil
}
func (da *DummyAdaptor) GetCount(o interface{}, c *int) error {
	*c = da.DummyCount
	return nil
}
func (da *DummyAdaptor) Aggregate(p []bson.M, res interface{}) error {
	return nil
}

func (da *DummyAdaptor) Switch(string) DbAdaptor {
	return nil
}
