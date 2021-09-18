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

func Upsert(document bson.D, filter bson.D, today string) (result *mongo.UpdateResult, e error) {
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

func getIdForToday(today string) string {
	val, err := db.IsPresentInCache(today)
	if err == nil {
		return val
	} else { // Not present in cache. Maybe Redis was re-started? Or it never got saved. Query and save
		cursor, err := db.FindLatestDoc()
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

func Fetchcall(c echo.Context) error {

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.rootnet.in/covid19-in/stats/latest", nil)
	if err != nil {
		fmt.Print(err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
		return c.JSON(http.StatusInternalServerError,
			models.ErrorMessage{Message: "Unable to fetch latest Covid Data for ssaving"})
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
	res, err := Upsert(update, filter, today)

	if err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusInternalServerError,
			models.ErrorMessage{Message: "Something went wrong while upserting"})
	} else {
		mongoId := getIdForToday(today)
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
