package biz

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kneerunjun/botmincock/dbadp"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	TEST_MONGO_HOST = "localhost:37017"
	TEST_MONGO_DB   = "botmincock_test"
	TEST_MONGO_COLL = "accounts"
)

// TestAdjustDayDebit : to check if the debit can be altered
func TestAdjustDayDebit(t *testing.T) {
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C("transacs")
	coll.RemoveAll(bson.M{})
	data := []*Transac{
		{TelegID: 5157350442, Debit: 150.00, Desc: PLAYDAY_DESC, Credit: 0.0, DtTm: time.Now()},
		{TelegID: 498116745, Debit: 150.00, Desc: PLAYDAY_DESC, Credit: 0.0, DtTm: time.Now()},
		{TelegID: 5116645118, Debit: 150.00, Desc: PLAYDAY_DESC, Credit: 0.0, DtTm: time.Now()},
		{TelegID: 961044876, Debit: 150.00, Desc: PLAYDAY_DESC, Credit: 0.0, DtTm: time.Now()},
	}
	// Before inserting the transactions, we can test for when no play debits
	for _, d := range data {
		coll.Insert(d)
	}
	c, _ := coll.Count()
	t.Logf("There are about %d test transactions in the database", c)
	from, to := TodayAsBoundary()
	trq := &TransacQ{Desc: PLAYDAY_DESC, From: from, To: to, Debits: 100.00}
	adp := dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "transacs")
	err := AdjustDayDebit(trq, adp)
	assert.Nil(t, err, "Unexpected error when AdjustDayDebit")
	if err != nil {
		t.Error(err)
		return
	}
}

func TestErrType(t *testing.T) {
	err := ERR_ACC404
	assert.True(t, errors.Is(err, ERR_ACC404), "Error comparison")
}

func TestTotalPlaydayDebits(t *testing.T) {
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C("transacs")
	coll.RemoveAll(bson.M{})
	data := []*Transac{
		{TelegID: 5157350442, Debit: 150.00, Desc: PLAYDAY_DESC, Credit: 0.0, DtTm: time.Now()},
		{TelegID: 498116745, Debit: 150.00, Desc: PLAYDAY_DESC, Credit: 0.0, DtTm: time.Now()},
		{TelegID: 5116645118, Debit: 150.00, Desc: PLAYDAY_DESC, Credit: 0.0, DtTm: time.Now()},
		{TelegID: 961044876, Debit: 150.00, Desc: PLAYDAY_DESC, Credit: 0.0, DtTm: time.Now()},
	}
	// Before inserting the transactions, we can test for when no play debits
	t.Log("No testing when no play debits")
	from, to := TodayAsBoundary()
	trq := TransacQ{TelegID: 5157350442, From: from, To: to}
	err := TotalPlaydayDebits(&trq, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "transacs"))
	assert.Nil(t, err, "Unexpected error when getting the total play day debits for today")
	assert.Equal(t, float32(0.0), trq.Debits, "Unexpected debits value when getting TotalPlaydayDebits")

	for _, d := range data {
		coll.Insert(d)
	}
	t.Log("Inserting test playday debit transactions")

	trq = TransacQ{TelegID: 5157350442, From: from, To: to}
	err = TotalPlaydayDebits(&trq, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "transacs"))
	assert.Nil(t, err, "Unexpected error when getting the total play day debits for today")
	t.Logf("Total debits for today %.2f", trq.Debits)
	// TEST: TODO: more tests for negation as well .. once this is tested it will relieve us of most of the function calls

}

func TestPlayerEstimates(t *testing.T) {
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C("estimates")
	defer func() {
		coll.RemoveAll(bson.M{}) // removing all the estimates inserted for the test purpose
	}()
	adp := dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "estimates")
	now := time.Now()
	data := []*Estimate{
		{TelegID: 5157350442, PlyDys: daysInMonth(now.Month(), now.Year())},
	}
	for _, d := range data {
		err := UpsertEstimate(d, adp)
		assert.Nil(t, err, "Uexpected error when upserting estimate")
		if err != nil {
			return
		}
		//TEST: here we can test getting the player estimate as well
		days, err := PlayerPlayDays(d.TelegID, adp)
		assert.Nil(t, err, "Uexpected error when getting estimate")
		assert.Equal(t, d.PlyDys, days, "Unexpected play days for the user")
		// If estimate is already added, checking if it can be updated
		d.PlyDys = 16
		err = UpsertEstimate(d, adp)
		assert.Nil(t, err, "Uexpected error when updating estimate")
		days, err = PlayerPlayDays(d.TelegID, adp)
		assert.Nil(t, err, "Uexpected error when getting estimate")
		assert.Equal(t, d.PlyDys, days, "Unexpected play days for the user")
	}

	// TEST: now testing for data that is not that ok
	dataNotOK := []*Estimate{
		{TelegID: 5157350442, PlyDys: 32},
		{TelegID: 5157350442, PlyDys: -1},
	}
	for _, d := range dataNotOK {
		err := UpsertEstimate(d, adp)
		assert.NotNil(t, err, "Uexpected nil error when upserting estimate")
	}
	// TEST: when the connection fails query error
	noConnect := dbadp.NewMongoAdpator("loclhos:37017", TEST_MONGO_DB, "estimates")
	err := UpsertEstimate(dataNotOK[0], noConnect)
	assert.NotNil(t, err, "Unexpected nil err when connecting with bad connection")
}

// TestAggrePlayerShare : from the estimates when we need the percentage of player contribution on any given day
func TestAggrePlayerShare(t *testing.T) {
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C("estimates")
	// Setting up the seed data
	done_seeding := func() bool {
		byt, err := readFromJsonF("../seeds/estimates.json")
		if err != nil {
			log.Error(err)
			return false
		}
		data := struct {
			Data []Estimate `json:"data"`
		}{}
		err = json.Unmarshal(byt, &data)
		if err != nil {
			return false
		}
		for idx, d := range data.Data {
			if err := coll.Insert(d); err != nil {
				log.Errorf("failed to insert data :%d", idx)
				return false
			}
		}
		return true
	}
	if !done_seeding() {
		t.Error("failed to seed database")
		return
	}
	result := struct {
		Total int `bson:"total"`
	}{}
	totalPlayDays := func() int {
		// this where we make calls to the database
		from, to := MonthAsBoundary()
		err := coll.Pipe([]bson.M{
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
		}).One(&result)
		if err != nil {
			if errors.Is(err, mgo.ErrNotFound) {
				return 0
			}
			return -1
		}
		return result.Total
	}
	tpd := totalPlayDays()
	assert.True(t, tpd > 0, "unexpected value for the total play days")
	t.Logf("Total playdays from the database %d", tpd)
	// getting the committed hours for only the player
	myPlayDays := func(tID int64) int {
		from, to := MonthAsBoundary()
		err := coll.Pipe([]bson.M{
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
		}).One(&result)
		if err != nil {
			if errors.Is(err, mgo.ErrNotFound) {
				return 0
			}
			return -1
		}
		return result.Total
	}
	mpd := myPlayDays(1633242782)
	assert.True(t, mpd > 0, "Unexpected number of the playdays for the player")
	if mpd > 0 && tpd > 0 {
		share := float32(mpd) / float32(tpd)
		t.Logf("share of player in playday %f", share)
	}
	func() {
		// flush test setup
		coll.RemoveAll(bson.M{})
	}()
}

func TestEstimatesInsert(t *testing.T) {
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C("estimates") // getting some dummy estimates into the database for this month
	// before we begin with any test, clearing off the data from the previous test
	coll.RemoveAll(bson.M{})
	okData := []Estimate{
		{TelegID: 5157350442, PlyDys: 31, DtTm: time.Now()},
		{TelegID: 498116745, PlyDys: 15, DtTm: time.Now()},
		{TelegID: 5116645118, PlyDys: 10, DtTm: time.Now()},
		{TelegID: 961044876, PlyDys: 31, DtTm: time.Now()},
	}
	for _, d := range okData {
		coll.Insert(d)
	}
}

/*
====================
TEST: this was to test getting the transactions based on date - direct query matching the date wasnt working
coll.Find(bson.M{"dttm": time.Now()}) will not give expected results of date matching
what works though is the comparison of date within a boundary $lte $gte
If now we want all the transactions of a day we define a boundary from the start of the day time to end of the day time
comparing the date field to be within the doundary works fine
====================
*/
func TestMarkPlayDay(t *testing.T) {
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	accounts := sess.DB("").C("accounts")
	accounts.RemoveAll(bson.M{}) // clearning the accounts before inserting them
	expenses := sess.DB("").C("expenses")
	expenses.RemoveAll(bson.M{})
	estimates := sess.DB("").C("estimates")
	estimates.RemoveAll(bson.M{})
	transacs := sess.DB("").C("transacs")
	transacs.RemoveAll(bson.M{})

	// Inserting accounts to the database
	archv := false
	elev := AccElev(User)
	data := []UserAccount{
		{Elevtn: &elev, Name: "Parker Sunman", Archived: &archv, Email: "jdurrans0@slate.com", TelegID: 498116745},
		{Elevtn: &elev, Name: "Helenelizabeth Grunson", Archived: &archv, Email: "aseller1@patch.com", TelegID: 5157350442},
		{Elevtn: &elev, Name: "Marcella Haggath", Archived: &archv, Email: "gsellner2@ning.com", TelegID: 5116645118},
		{Elevtn: &elev, Name: "Eldridge Baldinotti", Archived: &archv, Email: "gstarling3@earthlink.net", TelegID: 961044876},
	}

	for _, d := range data {
		accounts.Insert(d)
	}
	count, _ := accounts.Count()
	t.Log(infoMessage(fmt.Sprintf("We have about %d accounts in the test database", count)))
	// Adding some expenses to the database

	expData := []Expense{
		{TelegID: 5157350442, DtTm: time.Now(), Desc: "Court bookings", INR: 10000},
		{TelegID: 5157350442, DtTm: time.Now(), Desc: "Shuttles purchase", INR: 3800},
	}
	for _, d := range expData {
		expenses.Insert(d)
	}
	count, _ = expenses.Count()
	t.Log(infoMessage(fmt.Sprintf("We have about %d expenses in the test database", count)))
	// Now adding the estimates to the setup

	estData := []Estimate{
		{TelegID: 5157350442, PlyDys: 31, DtTm: time.Now()},
		{TelegID: 498116745, PlyDys: 15, DtTm: time.Now()},
		{TelegID: 5116645118, PlyDys: 10, DtTm: time.Now()},
		{TelegID: 961044876, PlyDys: 0, DtTm: time.Now()},
	}
	for _, d := range estData {
		estimates.Insert(d)
	}
	count, _ = estimates.Count()
	t.Log(infoMessage(fmt.Sprintf("We have about %d estimates in the test database", count)))

	// Now we can proceed to mark the play days:
	transacAdp := dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "transacs")
	now := time.Now()
	testTransacs := []*Transac{
		{TelegID: 5157350442, Desc: PLAYDAY_DESC, DtTm: now},
		{TelegID: 498116745, Desc: PLAYDAY_DESC, DtTm: now},
		{TelegID: 5116645118, Desc: PLAYDAY_DESC, DtTm: now},
	}
	for _, d := range testTransacs {
		err := MarkPlayday(d, transacAdp)
		assert.Nil(t, err, "Unexpected  error when marking the play day ")
	}
	// TEST: Previous recovery fromplaydays
	// Inserting a few transactions from the previous day
	// this will help us test to know if previous recovery is calculated correctly,
	// all the transactions until the previous  day are to be considered for recovery
	transacs.RemoveAll(bson.M{}) // removing playday markings from the previous test
	today := time.Now()
	yesterday := today.Add(-24 * time.Hour)
	previousData := []Transac{
		{TelegID: 5157350442, Credit: 0.0, Debit: 100, Desc: PLAYDAY_DESC, DtTm: yesterday},
		{TelegID: 498116745, Credit: 0.0, Debit: 100, Desc: PLAYDAY_DESC, DtTm: yesterday},
		{TelegID: 5116645118, Credit: 0.0, Debit: 100, Desc: PLAYDAY_DESC, DtTm: yesterday},
	}
	for _, d := range previousData {
		transacs.Insert(d)
	}
	count, _ = transacs.Count()
	t.Log(infoMessage(fmt.Sprintf("We have about %d previous transactions", count)))
	for _, d := range testTransacs {
		// this would be different from the previous since there has been some recovery
		err := MarkPlayday(d, transacAdp)
		assert.Nil(t, err, "Unexpected  error when marking the play day ")
	}
	// TEST: account does not exists
	noAccount := []*Transac{
		// accounts here arent registered at all
		{TelegID: 6167350449, Credit: 0.0, Debit: 100, Desc: PLAYDAY_DESC, DtTm: time.Now()},
	}
	for _, d := range noAccount {
		// this would be different from the previous since there has been some recovery
		err := MarkPlayday(d, transacAdp)
		assert.NotNil(t, err, "Unexpected nil error when marking playday - no account registry")
	}
	// TEST: player is already marked for the playday
	for _, d := range testTransacs {
		// wouldnt allow any transactions as the player is already marked for attendance
		err := MarkPlayday(d, transacAdp)
		assert.NotNil(t, err, "Unexpected nil error whe marking duplicate transactions")
	}
	// TEST: when there arent any etimates or estimates == 0
	// this happens when either no one has answered the polls or everyone has opted out of play for the month
	estimates.RemoveAll(bson.M{})
	transacs.RemoveAll(bson.M{})
	for _, d := range testTransacs {
		// wouldnt allow any transactions as the player is already marked for attendance
		err := MarkPlayday(d, transacAdp)
		t.Log(err.Error())
		assert.NotNil(t, err, "Unexpected nil error whe marking duplicate transactions")
	}
	// TEST: Now testing for when one of the player has opted for zero play days but has still sought to play in the month
	for _, d := range estData {
		estimates.Insert(d)
	}
	// 961044876  has opted out of play hence we try to add plyday transaction for this player
	err := MarkPlayday(&Transac{
		TelegID: 961044876,
		Credit:  0.0,
		Desc:    PLAYDAY_DESC,
	}, transacAdp)
	assert.NotNil(t, err, "Unexpected nil err when player estimate is zero")
}

func TestTotalMonthlyPlayDebits(t *testing.T) {
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C("transacs")
	okData := []int64{
		498116745,
		5157350442,
		5116645118,
		961044876,
	}
	checkTotal := float32(0)
	for _, d := range okData {
		coll.Insert(&Transac{TelegID: d, Debit: float32(100.00), Desc: PLAYDAY_DESC, DtTm: time.Now()})
		checkTotal += float32(100.00)
	}
	total := float32(0)
	err := RecoveryTillNow(dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "transacs"), &total)
	assert.Nil(t, err, "Unexpected error when TotalMonthlyPlayDebits")
	assert.Equal(t, checkTotal, total, "Totals of the play debits do not match")
	coll.RemoveAll(bson.M{})

}

func TestAccountBalance(t *testing.T) {
	/*====================
	Setup and seeding the database
	====================*/
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C("transacs")
	seed := []Transac{
		{TelegID: 5157350442, Credit: 320, Debit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350442, Credit: 420, Debit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350442, Credit: 329, Debit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350442, Credit: 320, Debit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350442, Credit: 320, Debit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350442, Credit: 325, Debit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350442, Credit: 330, Debit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350442, Credit: 322, Debit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350442, Debit: 321, Credit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350442, Debit: 100, Credit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350442, Debit: 100, Credit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350442, Debit: 100, Credit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350442, Debit: 100, Credit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350442, Debit: 100, Credit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350442, Debit: 100, Credit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350442, Debit: 100, Credit: 0.0, Desc: "sample test", DtTm: time.Now()},
	}
	credits := float32(0)
	debits := float32(0)
	for _, d := range seed {
		if coll.Insert(d) == nil {
			credits += d.Credit
			debits += d.Debit
		}
	}
	/*====================
	Simple balance test for one account
	====================*/
	checkBal := credits - debits // check the result from the database against this value
	t.Log(infoMessage(fmt.Sprintf("Expected balance of the account is %f", checkBal)))
	t.Log(infoMessage("Now testing a simple account balance.."))
	bl := &Balance{TelegID: 5157350442, DtTm: time.Now()}
	err := MyDues(bl, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "transacs"))
	assert.Nil(t, err, "Unexpected error when getting simple account balance")
	assert.Equal(t, checkBal, bl.Due, "Checkbalance test failed")

	/*====================
	now adding some noise to the data
	====================*/
	noise := []Transac{
		{TelegID: 5157350449, Credit: 420, Debit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350449, Credit: 329, Debit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350449, Credit: 320, Debit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350449, Credit: 320, Debit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350449, Credit: 325, Debit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350449, Credit: 330, Debit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350449, Credit: 322, Debit: 0.0, Desc: "sample test", DtTm: time.Now()},
		{TelegID: 5157350449, Debit: 321, Credit: 0.0, Desc: "sample test", DtTm: time.Now()},
	}
	for _, d := range noise {
		coll.Insert(d)
	}
	t.Log(infoMessage(fmt.Sprintf("Expected balance of the account is %f", checkBal)))
	t.Log(infoMessage("Now testing a simple account balance with data noise"))
	bl = &Balance{TelegID: 5157350442, DtTm: time.Now()}
	err = MyDues(bl, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "transacs"))
	assert.Nil(t, err, "Unexpected error when getting simple account balance")
	assert.Equal(t, checkBal, bl.Due, "Checkbalance test failed")

	/*====================
	Cleaning up the database
	====================*/
	t.Log(warnMessage("now clearing the database.."))
	coll.RemoveAll(bson.M{})

}

func TestTeamMonthlyExpense(t *testing.T) {
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C("expenses")
	// Inserting some test expenses
	okData := []*Expense{
		{TelegID: 5157350442, Desc: "Dimorphocarpa wislizeni", INR: 300, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "Apocynum L", INR: 300, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "Schedonorus giganteus", INR: 300, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "Pertusaria carneopallida", INR: 300, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "Desmatodon latifolius", INR: 300, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "Stenaria mullerae", INR: 300, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "Muhlenbergia arenacea", INR: 300, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "Illosporium carneum Fr", INR: 300, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "Bouchetia erecta DC", INR: 300, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "Chromolaena corymbosa", INR: 300, DtTm: time.Now()},
	}
	testSum := float32(0.0)
	for _, d := range okData {
		if coll.Insert(d) == nil {
			testSum += d.INR
		}
	}
	/*====================
	Actual test with only one user expenses
	====================*/
	t.Log(infoMessage("now testing the team's aggregate monthly expenses.."))
	mnthExp := &MnthlyExpnsQry{Dttm: time.Now()}
	err := TeamMonthlyExpense(mnthExp, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "expenses"))
	assert.Nil(t, err, "unexpected err when getting the team monthly expense")
	assert.Equal(t, testSum, mnthExp.Total, "total of the team monthly expense does not match")

	/*
		We then add some noise in the database -
		expenses from other months and test to see if get the same total
	*/
	noiseData := []*Expense{
		{TelegID: 5116645118, Desc: "noise", INR: 100, DtTm: time.Date(2022, time.May, 20, 0, 0, 0, 0, time.Local)},
		{TelegID: 5116645118, Desc: "noise", INR: 100, DtTm: time.Date(2023, time.April, 20, 0, 0, 0, 0, time.Local)},
	}
	for _, d := range noiseData {
		coll.Insert(d)
	}
	t.Log(infoMessage("now testing the team's aggregate monthly expenses with noise in the data"))
	mnthExp = &MnthlyExpnsQry{Dttm: time.Now()}
	err = TeamMonthlyExpense(mnthExp, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "expenses"))
	assert.Nil(t, err, "unexpected err when getting the team monthly expense")
	assert.Equal(t, testSum, mnthExp.Total, "total of the team monthly expense does not match")
	/*====================
	cleanup
	====================*/
	t.Log(warnMessage("now clearing the database.."))
	coll.RemoveAll(bson.M{})
}

func TestUserMonthlyExpense(t *testing.T) {
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C("expenses")
	// Inserting some test expenses
	okData := []*Expense{
		{TelegID: 5157350442, Desc: "testexpense1", INR: 303, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "testexpense1", INR: 303, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "testexpense1", INR: 306, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "testexpense1", INR: 308, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "testexpense1", INR: 309, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "testexpense1", INR: 305, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "testexpense1", INR: 303, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "testexpense1", INR: 302, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "testexpense1", INR: 301, DtTm: time.Now()},
		{TelegID: 5157350442, Desc: "testexpense1", INR: 300, DtTm: time.Now()},
	}
	testSum := float32(0.0)
	for _, d := range okData {
		if coll.Insert(d) == nil {
			testSum += d.INR
		}
	}
	/*====================
	Actual test with only one user expenses
	====================*/
	t.Log(infoMessage("now testing the aggregate monthly expenses.."))
	mnthExp := &MnthlyExpnsQry{TelegID: 5157350442, Dttm: time.Now()}
	err := UserMonthlyExpense(mnthExp, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "expenses"))
	assert.Nil(t, err, "unexpected error when aggregating user monthly expense")
	assert.Equal(t, testSum, mnthExp.Total, "total of the user monthly expense does not match")

	// TEST: test when TelegID is not uniform - we need to see if user id is matched correcly
	// We insert new data and then test with same tests above to know if 5157350442 is matched correctly
	noiseData := []*Expense{
		{TelegID: 5116645118, Desc: "testexpense1", INR: 303, DtTm: time.Now()},
		{TelegID: 5116645118, Desc: "testexpense1", INR: 303, DtTm: time.Now()},
		{TelegID: 5116645118, Desc: "testexpense1", INR: 306, DtTm: time.Now()},
		{TelegID: 5116645118, Desc: "testexpense1", INR: 308, DtTm: time.Now()},
		{TelegID: 5116645118, Desc: "testexpense1", INR: 309, DtTm: time.Now()},
		{TelegID: 5116645118, Desc: "testexpense1", INR: 305, DtTm: time.Now()},
		{TelegID: 5116645118, Desc: "testexpense1", INR: 303, DtTm: time.Now()},
	}
	for _, d := range noiseData {
		coll.Insert(d)
	}

	/*====================
	test with multiple user data in the expenses database
	the test is the same but this time database has data for another user as well
	====================*/
	t.Log(infoMessage("now testing the aggregate monthly expenses with data noise"))
	mnthExp = &MnthlyExpnsQry{TelegID: 5157350442, Dttm: time.Now()}
	err = UserMonthlyExpense(mnthExp, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "expenses"))
	assert.Nil(t, err, "unexpected error when aggregating user monthly expense")
	assert.Equal(t, testSum, mnthExp.Total, "total of the user monthly expense does not match")

	/*====================
	cleanup
	====================*/
	t.Log(warnMessage("now clearing the database.."))
	coll.RemoveAll(bson.M{})
}

func TestAddNewExpense(t *testing.T) {
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C("expenses")
	defer coll.RemoveAll(bson.M{})
	defer t.Log(warnMessage("now clearing the database.."))
	t.Log(infoMessage("now testing for one sanple expense"))
	d := &Expense{INR: 1055.00, TelegID: 5157350442, DtTm: time.Now(), Desc: "test expense, purchase of shuttles"}
	err := RecordExpense(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "expenses"))
	assert.Nil(t, err, "unexpected error when recording an expense")

	// TEST: for negative test cases
	dataNotOK := []*Expense{
		nil, // nil expense is outrightly rejected
		{TelegID: 5157350442, INR: 0.0},
		{TelegID: 5157350442},
	}
	t.Log(infoMessage("now testing negative cases"))
	for _, d := range dataNotOK {
		err := RecordExpense(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, "expenses"))
		assert.NotNil(t, err, "unexpected nil err when data not ok")
	}
}

func TestRegisterAccount(t *testing.T) {
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C(TEST_MONGO_COLL)
	// TEST: happy test, no error
	dataOk := []*UserAccount{
		{TelegID: 5435345, Email: "cayce0@bbb.org", Name: "Conrado Ayce"},
		{TelegID: 5435346, Email: "rscimoni1@paypal.com", Name: "Reagen Scimon"},
		{TelegID: 5435347, Email: "ekornousek2@apple.com", Name: "Elyssa Kornousek"},
	}
	for _, d := range dataOk {
		err := RegisterNewAccount(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
		assert.Nil(t, err, "Unexpected error when registering new account")
		// After having inserted the accounts we need to test if the account has elevation 0 and archive set to false
		acc := UserAccount{}
		coll.Find(bson.M{"tid": d.TelegID}).One(&acc)
		assert.Equal(t, AccElev(User), *acc.Elevtn, fmt.Sprintf("Unexpected elevation for the account %d", d.TelegID))
		assert.False(t, *acc.Archived, fmt.Sprintf("Unexpected Archive flag for the accounts %d", d.TelegID))
	}

	t.Log(infoMessage("Done testing happy path accounts"))
	// ===========================
	// TEST: email of the account being registered isnt valid
	dataNotOk := []*UserAccount{
		{TelegID: 5435345, Email: "cayce0@bbb.org.cm", Name: "Conrado Ayce"},
		{TelegID: 5435346, Email: "rsci%^$%^moni1@paypal.com", Name: "Reagen Scimon"},
		{TelegID: 5435347, Email: "@apple.com", Name: "Elyssa Kornousek"},
	}
	for _, d := range dataNotOk {
		err := RegisterNewAccount(d, &dbadp.DummyAdaptor{DummyCount: 0})
		assert.NotNil(t, err, "Unexpected nil error when registering new account")
	}
	t.Log(infoMessage("Done testing for accounts with invalid email"))
	// ===========================
	// TEST:  testing for duplicate accounts
	dataDuplicate := []*UserAccount{
		{TelegID: 5435345, Email: "cayce0@bbb.org", Name: "Conrado Ayce"},
		{TelegID: 5435348, Email: "rscimoni1@paypal.com", Name: "Reagen Scimon"}, // email is repeated
		{TelegID: 5435347, Email: "ekornousek2@apple.com", Name: "Elyssa Kornousek"},
	}
	for _, d := range dataDuplicate {
		err := RegisterNewAccount(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
		assert.NotNil(t, err, "Unexpected nil error when registering new account")
	}
	t.Log(infoMessage("Done testing for accounts with duplicate ID/email"))
	// ===========================
	// TEST: testing for archived accounts - does it re-register the acocunt
	// ============= Setting up archived accounts

	archive := true
	coll.Update(&UserAccount{TelegID: dataOk[0].TelegID}, bson.M{"$set": &UserAccount{Archived: &archive}})

	err := RegisterNewAccount(dataOk[0], dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
	assert.NotNil(t, err, "Unexpected nil error when registering archived account")
	t.Log(infoMessage("Done testing for accounts that are archived"))
	// ========================
	// cleaning up the test
	t.Log(warnMessage("Now cleannig up the database.."))
	coll.RemoveAll(bson.M{})
}

// https://github.com/fatih/color/blob/f4c431696a22e834b83444f720bd144b2dbbccff/color.go#L67
func warnMessage(m string) string {
	return fmt.Sprintf("\x1b[%dm!%s\x1b[0m", 36, m)
}
func infoMessage(m string) string {
	return fmt.Sprintf("\x1b[%dm>%s\x1b[0m", 34, m)
}

func TestElevateAcc(t *testing.T) {
	/*
		setting up the test
		- database connection
		- inserting the data
	*/
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C(TEST_MONGO_COLL)

	archive := false
	userElev := AccElev(User)
	dataOk := []*UserAccount{
		{Elevtn: &userElev, TelegID: 5435345, Email: "cayce0@bbb.org", Name: "Conrado Ayce", Archived: &archive},
		{Elevtn: &userElev, TelegID: 5435346, Email: "rscimoni1@paypal.com", Name: "Reagen Scimon", Archived: &archive},
		{Elevtn: &userElev, TelegID: 5435347, Email: "ekornousek2@apple.com", Name: "Elyssa Kornousek", Archived: &archive},
	}
	for _, d := range dataOk {
		coll.Insert(d)
	}
	t.Log(infoMessage("Database setup complete"))
	for _, d := range dataOk {
		*d.Elevtn = AccElev(Manager)
		err := ElevateAccount(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
		assert.Nil(t, err, warnMessage("failed elevate account"))
	}
	// TEST: testing for data not ok
	overElev := AccElev(uint(5)) // over elevation of an account should not be possible
	for _, d := range dataOk {
		d.Elevtn = &overElev
		err := ElevateAccount(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
		assert.NotNil(t, err, "Unexpected not nil error when testing with over elevation of the account")
	}

	/*
		Cleaning up the test database
	*/
	t.Cleanup(func() {
		t.Log(warnMessage("Now clearing up the test database.."))
		coll.RemoveAll(bson.M{})
	})
}

func TestUpdateAccountEmail(t *testing.T) {
	/*
		Setting up ithe database for email modification test
	*/
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C(TEST_MONGO_COLL)
	// cleaning up the test database

	archive := false
	userElev := AccElev(User)
	dataOk := []*UserAccount{
		{Elevtn: &userElev, TelegID: 5435345, Email: "cayce0@bbb.org", Name: "Conrado Ayce", Archived: &archive},
		{Elevtn: &userElev, TelegID: 5435346, Email: "rscimoni1@paypal.com", Name: "Reagen Scimon", Archived: &archive},
		{Elevtn: &userElev, TelegID: 5435347, Email: "ekornousek2@apple.com", Name: "Elyssa Kornousek", Archived: &archive},
	}
	for _, d := range dataOk {
		coll.Insert(d)
	}
	// TEST: positive test for error less update on the email of the account
	newEmails := []string{
		"sglancy0@github.io",
		"msach1@samsung.com",
		"agosnell2@lulu.com",
	}
	t.Log(infoMessage("Now testing for valid accounts.."))
	for idx, d := range dataOk {
		d.Email = newEmails[idx]
		err := UpdateAccountEmail(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
		assert.Nil(t, err, warnMessage(fmt.Sprintf("failed to update email for %s", d.Email)))
	}
	// TEST: account is nil or the email isnt valid
	invalidData := []*UserAccount{
		{TelegID: 5435345, Email: "mkonmann3@who.int.gov"},
		{TelegID: 5435345, Email: ""},
		nil,
	}
	t.Log(infoMessage("Now testing for invalid accounts.."))
	for _, d := range invalidData {
		err := UpdateAccountEmail(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
		assert.NotNil(t, err, warnMessage("unexpected nil error when updating invalid email and accounts"))
	}
	t.Cleanup(func() {
		t.Log(warnMessage("Now clearing up the test database.."))
		coll.RemoveAll(bson.M{})
	})
}

func TestDeregisterAccount(t *testing.T) {
	sess, _ := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{TEST_MONGO_HOST},
		Timeout:  4 * time.Second,
		Database: TEST_MONGO_DB,
	})
	coll := sess.DB("").C(TEST_MONGO_COLL)
	// cleaning up the test database

	archive := false
	userElev := AccElev(User)
	dataOk := []*UserAccount{
		{Elevtn: &userElev, TelegID: 5435345, Email: "cayce0@bbb.org", Name: "Conrado Ayce", Archived: &archive},
		{Elevtn: &userElev, TelegID: 5435346, Email: "rscimoni1@paypal.com", Name: "Reagen Scimon", Archived: &archive},
		{Elevtn: &userElev, TelegID: 5435347, Email: "ekornousek2@apple.com", Name: "Elyssa Kornousek", Archived: &archive},
	}
	for _, d := range dataOk {
		coll.Insert(d)
	}
	// TEST: positive tests , account is marked as archive
	t.Log(infoMessage("Now testing for valid accounts.."))
	for _, d := range dataOk {
		err := DeregisterAccount(d, dbadp.NewMongoAdpator(TEST_MONGO_HOST, TEST_MONGO_DB, TEST_MONGO_COLL))
		assert.Nil(t, err, warnMessage(fmt.Sprintf("failed to deregister email for %s", d.Email)))
	}
	t.Cleanup(func() {
		t.Log(warnMessage("Now clearing up the test database.."))
		coll.RemoveAll(bson.M{})
	})
}

func TestEmailRegx(t *testing.T) {
	data := []string{
		"bdensumbe0@youku.com",
		"mroth3@soup.io",
		"rlambregts4@godaddy.com",
		"jhucker.be9@virginia.edu",
		"lgh_iona@newyorker.com",
		"kneerunjun@gmail.com",
		"awatiniranjan@gmail.com",
	}
	dataNotOk := []string{
		"ahubatsch8@bbc.co.uk",
		"fpetrazzib@state.tx.us",
		"@state.tx.us",
		"fpetrazzibstate.us",
		"#fpetrazzib@state.us",
	}
	for _, d := range data {
		assert.True(t, REGX_EMAIL.MatchString(d), fmt.Sprintf("email %s should have passed the test", d))
	}
	for _, d := range dataNotOk {
		assert.False(t, REGX_EMAIL.MatchString(d), fmt.Sprintf("email %s shouldn't have passed the test", d))
	}
}
