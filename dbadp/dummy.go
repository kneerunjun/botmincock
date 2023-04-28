package dbadp

/* ==================================
author 		: kneerunjun@gmail.com
time		: April 2023
project		: botmincock
For the purpose of usint testing other packages that want to be agnostic of the logic thats implemented here on the database adaptor we implement a facade of a adaptor
====================================*/
import "reflect"

type DummyAdaptor struct {
}

func (da *DummyAdaptor) AddOne(interface{}) error {
	return nil
}
func (da *DummyAdaptor) RemoveOne(interface{}) error {
	return nil
}
func (da *DummyAdaptor) UpdateOne(interface{}) error {
	return nil
}
func (da *DummyAdaptor) GetOne(interface{}, reflect.Type) (interface{}, error) {
	return nil, nil
}
func (da *DummyAdaptor) GetCount(interface{}, *int) error {
	return nil
}
