package apis

import (
	"InShorts/src/db"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func FindLatestDoc() (result MongoResponse, e error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := db.Conn().Database("testing").Collection("covid")
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

func retrieveKey() string {

	if _, err := os.Stat("../key.txt"); err == nil { // Useful locally
		fmt.Println("Hmm")
		b, err := ioutil.ReadFile("../key.txt")
		if err == nil {
			return string(b)
		} else {
			return os.Getenv("apiKey")
		}
	} else {
		fmt.Println(err.Error())
	}
	return os.Getenv("apiKey") // Useful when deployed publically
}

func buildURL(lat string, long string) string {
	key := retrieveKey()
	url := "https://us1.locationiq.com/v1/reverse.php?key=%s&lat=%s&lon=%s&format=json"
	return fmt.Sprintf(url, key, lat, long)
}

func LocResults(c echo.Context) error {
	client := &http.Client{}
	lat := c.Param("lat")
	long := c.Param("long")
	url := buildURL(lat, long)
	fmt.Println("Got URL as " + url)

	req, err := http.NewRequest("GET", url, nil)
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

	res, err := FindLatestDoc()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	} else {
		for _, element := range res.Region {
			if element.Loc == responseObject.Address.State {
				return c.JSON(http.StatusOK, UserResponse{element.Loc,
					element.Indiancases, res.Summary.Total, res.LastUpdated})
			}
		}
	}
	return c.JSON(http.StatusNotFound, StateNotFound{responseObject.Address.State, "Not Found"})
}
