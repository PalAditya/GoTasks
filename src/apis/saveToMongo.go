package apis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"InShorts/src/db"
	"InShorts/src/models"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//Upsert upserts a document in Mongo with the filter crieteria being that report date is today
func (externalClient OExternal) Upsert(document bson.D, filter bson.D, today string) (result *mongo.UpdateResult, e error) {
	ctx := db.GetCTX()
	collection := db.Conn().Database("testing").Collection("covid")
	opts := options.Update().SetUpsert(true)
	res, err := collection.UpdateOne(ctx, filter, document, opts)
	if res.UpsertedID != nil {
		fmt.Println("Going to cache id entry for today " + today)
		db.SaveToCache(today, res.UpsertedID.(primitive.ObjectID).Hex(), 0)
	}
	return res, err
}

//GetIdForToday Gets the Mongo Id for the document that has report date as today, either from Redis or Mongo
func (externalClient OExternal) GetIdForToday(today string) string {
	dbMethod := db.ODBExternal{}
	val, err := db.IsPresentInCache(today)
	if err == nil {
		return val
	} else { // Not present in cache. Maybe Redis was re-started? Or it never got saved. Query and save
		cursor, err := dbMethod.FindLatestDoc()
		if err != nil {
			log.Println("Unable to fetch latest doc from Mongo")
			return "-"
		}
		ctx := db.GetCTX()
		var record []models.MongoResponseWithId
		if err = cursor.All(ctx, &record); err != nil {
			return "-"
		}
		db.SaveToCache(today, record[0].Id, 0)
		return record[0].Id
	}
}

//FetchCall takes the responsibility of querying an external endpoint and saving the data to mongo datastore
func Fetchcall(c echo.Context, externalClient IExternal) error {

	client := &http.Client{}
	resp, err := externalClient.MakeExternalHTTPRequest(client, "https://api.rootnet.in/covid19-in/stats/latest")
	if err != nil {
		fmt.Print(err.Error())
		return c.JSON(http.StatusInternalServerError, models.ErrorMessage{
			Message: "We are having issues retrieving data, please check later"})
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
		return c.JSON(http.StatusInternalServerError,
			models.ErrorMessage{Message: "Unable to fetch latest Covid Data for saving"})
	}
	var responseObject models.Response
	json.Unmarshal(bodyBytes, &responseObject)

	today := time.Now().Format("2006-01-02")
	query := bson.D{{Key: "res", Value: responseObject.Data.Regional},
		{Key: "lastUpdated", Value: responseObject.LastRefreshed},            //Server update time for API
		{Key: "modifiedAt", Value: time.Now().Format("2006-01-02 15:04:05")}, //last time we updated data on DB
		{Key: "recordDate", Value: today},
		{Key: "summary", Value: models.DBSummary{
			Total:      responseObject.Data.Summary.Indiancases,
			Discharged: responseObject.Data.Summary.Discharged}}}
	update := bson.D{{Key: "$set", Value: query}}
	filter := bson.D{{Key: "recordDate", Value: today}}
	res, err := externalClient.Upsert(update, filter, today)

	if err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusInternalServerError,
			models.ErrorMessage{Message: "Something went wrong while upserting"})
	} else {
		mongoId := externalClient.GetIdForToday(today)
		if res.UpsertedCount == 1 {
			if mongoId == "-" {
				return c.JSON(http.StatusInternalServerError,
					models.ErrorMessage{Message: "Unable to find Id of last upserted record"})
			} else {
				return c.JSON(http.StatusOK, models.ApiResponse{
					LatestRecord: mongoId,
					Inserted:     true,
					Modified:     false})
			}
		} else {
			if mongoId == "-" {
				return c.JSON(http.StatusInternalServerError,
					models.ErrorMessage{Message: "Unable to find Id of last upserted record"})
			} else {
				return c.JSON(http.StatusOK, models.ApiResponse{
					LatestRecord: mongoId,
					Inserted:     false,
					Modified:     true})
			}
		}
	}
}
