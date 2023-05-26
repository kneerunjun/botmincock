package biz

/* =================
> Estimates are indicative player availability for the upcoming month
> Estimates for each month let the bot decide on how to equitably divide the expenses for each day in the month
> Estimates also help the bot to arrive at the recovery deficit for each day when a player who has promised to play does not play
> Estimates once given can be changed, but only the debits ahead of the change will be affected
================= */

import (
	"errors"
	"fmt"
	"time"

	"github.com/kneerunjun/botmincock/dbadp"
	"github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// UpsertEstimate 	: Inserts or updates an estimate for the player only for the given month
// incase the estimate is already added - the estimate is updated
// incase the playdays are invalid - error
// Incase the db gateway fails - error
func UpsertEstimate(est *Estimate, iadp dbadp.DbAdaptor) error {
	errLoc := "UpsertEstimate"
	est.DtTm = time.Now() // since the estimate is always for the current month only
	// checking to see if the estimate has 0 <= plydays >= max monthly days
	if 0 > est.PlyDys || daysInMonth(est.DtTm.Month(), est.DtTm.Year()) < est.PlyDys {
		// invalid number of play days this needs to send back an error
		return NewDomainError(fmt.Errorf("invalid number of play days in the estimate"), nil).SetLoc(errLoc).SetUsrMsg("Invalid play days for the estimate indicated, kindly check and send again").SetLogEntry(logrus.Fields{
			"playdays": est.PlyDys,
			"tid":      est.TelegID,
		})
	}
	// If no record found for the player we add a new estimate
	days, err := PlayerPlayDays(est.TelegID, iadp)
	if err != nil { // cannot be the case when result.Total  ==0
		if days == 0 {
			if err := iadp.AddOne(est); err != nil {
				return NewDomainError(fmt.Errorf("failed UpsertEstimate"), err).SetLoc(errLoc).SetUsrMsg(failed_query("adding the estimates"))
			}
			return nil
		}
		return err
	}
	// If record found for the player, we update the estimate
	from, to := MonthAsBoundary()
	selectPlayrEst := bson.M{
		"dttm": bson.M{
			"$gte": from,
			"$lte": to,
		},
		"tid": est.TelegID,
	} // for the current month this can select the player estimate
	// If there werent an error we just update
	if err := iadp.UpdateOne(selectPlayrEst, bson.M{"plydys": est.PlyDys}); err != nil {
		return NewDomainError(fmt.Errorf("failed UpsertEstimate"), err).SetLoc(errLoc).SetUsrMsg(failed_query("updating the estimates"))
	}
	return nil
}

// TotalPlayDays 	: for the given month the play day estimates are summed up, this is useful when getting the player contribution ratio
// 0, err 			: no records found, implies for the given month either the poll wasnt published or no one answered the poll
// -1, err			: error in getting records, gateway query failed.
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
	return result.Total, nil
}

// PlayerShare : for any player that has indicated his efforts estimate, this will get share of his contribution for a given month
// 0, err 			: no records found, implies the player has not answered the poll
// -1, err			: error in getting records, gateway query failed.
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
			return 0, NewDomainError(fmt.Errorf("zero PlayerPlayDays"), nil).SetLoc(errLoc).SetUsrMsg(zero_playdays())
		}
		return -1, NewDomainError(fmt.Errorf("failed PlayerPlayDays"), nil).SetLoc(errLoc).SetUsrMsg(failed_query("getting the total monthly playdays"))
	}
	return result.Total, nil
}
