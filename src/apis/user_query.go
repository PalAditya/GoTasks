package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GeoResponse struct {
	X       map[string]interface{} `json:"-"`
	Address Address
}

type Address struct {
	State string                 `json:"state"`
	X     map[string]interface{} `json:"-"`
}

type MongoResponse struct {
	Region      []Regional `bson:"res"`
	LastUpdated string     `bson:"lastUpdated"`
	ModifiedAt  string     `bson:"modifiedAt"`
	RecordDate  string     `bson:"recordDate"`
	Summary     Summary    `bson:"summary"`
}

func FindLatestDoc() (result MongoResponse, e error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := Conn().Database("testing").Collection("covid")
	opts := options.Find()
	opts.SetSort(bson.D{{"recordDate", -1}})
	opts.SetLimit(1)
	cursor, err := collection.Find(ctx, bson.D{}, opts)
	var record []MongoResponse
	if err = cursor.All(ctx, &record); err != nil {
		return MongoResponse{}, err
	}
	return record[0], err
}

func UserResults(c echo.Context) error {
	client := &http.Client{}
	lat := c.Param("lat")
	long := c.Param("long")
	fmt.Println("https://us1.locationiq.com/v1/reverse.php?key=pk.06c68e4b509f2a73643305e760c488eb&lat=" + lat + "&lon=" + long + "&format=json")

	req, err := http.NewRequest("GET", "https://us1.locationiq.com/v1/reverse.php?key=pk.06c68e4b509f2a73643305e760c488eb&lat="+lat+"&lon="+long+"&format=json", nil)
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
	var responseObject GeoResponse
	json.Unmarshal(bodyBytes, &responseObject)
	fmt.Println(responseObject)

	res, err := FindLatestDoc()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	} else {
		for _, element := range res.Region {
			fmt.Println(element)
			if element.Loc == responseObject.Address.State {
				return c.JSON(http.StatusOK, element)
			}
		}
	}
	return c.JSON(http.StatusNotFound, bson.D{{"state", "Not Found"}})
}
