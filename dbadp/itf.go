package dbadp

import "gopkg.in/mgo.v2/bson"

// DBAdaptor : for each of the databases we can have an concrete adaptors that can help execute queries in context
type DBAdaptor interface {
	GetOne(m bson.M, db, coll string) (bson.M, error)
}

type NosqlDB interface {
}
