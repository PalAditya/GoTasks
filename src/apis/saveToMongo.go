package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"InShorts/src/db"
	"InShorts/src/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Upsert(document bson.D, filter bson.D) (result *mongo.UpdateResult, e error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := db.Conn().Database("testing").Collection("covid")
	opts := options.Update().SetUpsert(true)
	res, err := collection.UpdateOne(ctx, filter, document, opts)
	return res, err
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
		return c.JSON(http.StatusInternalServerError, models.ErrorMessage{"Unable to fetch latest Covid Data for ssaving"})
	}
	var responseObject models.Response
	json.Unmarshal(bodyBytes, &responseObject)

	today := time.Now().Format("2006-01-02")
	query := bson.D{{"res", responseObject.Data.Regional},
		{"lastUpdated", responseObject.LastRefreshed},            //Server update time for API
		{"modifiedAt", time.Now().Format("2006-01-02 15:04:05")}, //last time we updated data on DB
		{"recordDate", today},
		{"summary", models.DBSummary{responseObject.Data.Summary.Indiancases, responseObject.Data.Summary.Discharged}}}
	update := bson.D{{"$set", query}}
	filter := bson.D{{"recordDate", today}}
	res, err := Upsert(update, filter)

	if err != nil {
		log.Printf(err.Error())
		return c.JSON(http.StatusInternalServerError, models.ErrorMessage{"Something went wrong while upserting"})
	} else {
		return c.JSON(http.StatusOK, res)
	}
}
