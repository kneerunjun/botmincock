package biz

import (
	"errors"
	"fmt"

	"github.com/kneerunjun/botmincock/dbadp"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func TotalPlayDays(iadp dbadp.DbAdaptor) (int, error) {
	errLoc := "TotalPlayDays"
	result := struct {
		Total int `bson:"total"`
	}{}
	from, to := MonthAsBoundary()
	err := iadp.Aggregate([]bson.M{
		{"$match": bson.M{
			"dttm": bson.M{
				"$gte": from,
				"$lte": to,
			},
		}},
		{"$group": bson.M{
			"_id": nil,
			"total": bson.M{
				"$sum": "$plydys",
			},
		}},
		{"$project": bson.M{
			"_id": 0,
		}},
	}, &result)
	if err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			return 0, NewDomainError(fmt.Errorf("zero TotalPlayDays"), nil).SetLoc(errLoc).SetUsrMsg(zero_playdays())
		}
		return -1, NewDomainError(fmt.Errorf("failed TotalPlayDays"), nil).SetLoc(errLoc).SetUsrMsg(failed_query("getting the total monthly playdays"))
	}
	if result.Total > 0 {
		return result.Total, nil
	} else {
		return result.Total, NewDomainError(fmt.Errorf("zero TotalPlayDays"), nil).SetLoc(errLoc).SetUsrMsg(zero_playdays())
	}

}

// PlayerShare : for any player that has indicated his efforts estimate, this will get share of his contribution for a given month
// tID 		: share of the estimates any player for the current month
// this assumes you have valid data from the poll else error
func PlayerPlayDays(tID int64, iadp dbadp.DbAdaptor) (int, error) {
	errLoc := "PlayerPlayDays"
	result := struct {
		Total int `bson:"total"`
	}{}
	from, to := MonthAsBoundary()
	err := iadp.Aggregate([]bson.M{
		{"$match": bson.M{
			"dttm": bson.M{
				"$gte": from,
				"$lte": to,
			},
			"tid": tID, // this will get us only the player playdays
		}},
		{"$group": bson.M{
			"_id": nil,
			"total": bson.M{
				"$sum": "$plydys",
			},
		}},
		{"$project": bson.M{
			"_id": 0,
		}},
	}, &result)
	if err != nil {
		if errors.Is(err, mgo.ErrNotFound) {
			return 0, nil
		}
		return -1, NewDomainError(fmt.Errorf("failed PlayerPlayDays"), nil).SetLoc(errLoc).SetUsrMsg(failed_query("getting the total monthly playdays"))
	}
	if result.Total > 0 {
		return result.Total, nil
	} else {
		return result.Total, NewDomainError(fmt.Errorf("zero PlayerPlayDays"), nil).SetLoc(errLoc).SetUsrMsg(zero_playdays())
	}

}
