package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Response struct {
	Success          bool `json:"success"`
	Data             Data
	LastRefreshed    string `json:"lastRefreshed"`
	LastOriginUpdate string `json:"lastOriginUpdate"`
}

type Data struct {
	Regional []Regional
	Summary  Summary
	X        map[string]interface{} `json:"-"`
}

type Summary struct {
	Indiancases int64                  `json:"confirmedCasesIndian"`
	Discharged  int64                  `json:"discharged"`
	X           map[string]interface{} `json:"-"`
}

type Regional struct {
	Loc           string `json:"loc"`
	Indiancases   int64  `json:"confirmedCasesIndian"`
	Foreigncases  int64  `json:"confirmedCasesForeign"`
	Discharged    int64  `json:"discharged"`
	Deaths        int64  `json:"deaths"`
	Totalconfimed int64  `json:"totalConfirmed"`
}

func Conn() (client *mongo.Client) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

func Upsert(document bson.D, filter bson.D) (result *mongo.UpdateResult, e error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := Conn().Database("testing").Collection("covid")
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
	}
	var responseObject Response
	json.Unmarshal(bodyBytes, &responseObject)

	today := time.Now().Format("2006-01-02")
	query := bson.D{{"res", responseObject.Data.Regional},
		{"lastUpdated", responseObject.LastRefreshed},            //Server update time for API
		{"modifiedAt", time.Now().Format("2006-01-02 15:04:05")}, //last time we updated data on DB
		{"recordDate", today},                                    //While not needed, we will store each day's records separately
		{"summary", bson.D{{"total", responseObject.Data.Summary.Indiancases}, {"discharged", responseObject.Data.Summary.Discharged}}}}
	update := bson.D{{"$set", query}}
	filter := bson.D{{"recordDate", today}}
	res, err := Upsert(update, filter)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	} else {
		return c.JSON(http.StatusOK, res)
	}
}
