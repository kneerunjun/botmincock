package biz

import (
	"errors"
	"fmt"

	"github.com/kneerunjun/botmincock/dbadp"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// PlayerShare : for any player that has indicated his efforts estimate, this will get share of his contribution for a given month
// tID 		: share of the estimates any player for the current month
// this assumes you have valid data from the poll else error

func PlayerShare(tID int64, iadp dbadp.DbAdaptor) (float32, error) {
	errLoc := "PlayerShare"
	result := struct {
		Total int `bson:"total"`
	}{}
	totalPlayDays := func() int {
		// this where we make calls to the database
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
				return 0
			}
			return -1
		}
		return result.Total
	}
	tpd := totalPlayDays()
	if tpd < 0 {
		return 0.0, NewDomainError(fmt.Errorf("failed to get total play days"), nil).SetLoc(errLoc).SetUsrMsg("Wasnt possible to get the estimated playdays, please report this to an admin").SetLogEntry(logrus.Fields{
			"total_play_days": tpd,
		})
	}
	myPlayDays := func(tID int64) int {
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
				return 0
			}
			return -1
		}
		return result.Total
	}
	mpd := myPlayDays(tID)
	if mpd < 0 {
		return 0.0, NewDomainError(fmt.Errorf("failed to get player play days"), nil).SetLoc(errLoc).SetUsrMsg("Wasnt possible to get the estimated playdays, please report this to an admin").SetLogEntry(logrus.Fields{
			"my_play_days": mpd,
		})
	}
	return float32(mpd) / float32(tpd), nil
}
