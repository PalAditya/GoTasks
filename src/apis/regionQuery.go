package apis

import (
	"InShorts/src/db"
	"InShorts/src/models"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func isPresentInCache(key string) (userResponse models.UserResponse, e error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client := db.RedisClient()
	val, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Println("Key was not present in cache. Going to fetch from db")
		return models.UserResponse{}, err
	} else if err != nil {
		log.Println("Unable to interact with cache")
		return models.UserResponse{}, err
	} else {
		var resp models.UserResponse
		json.Unmarshal([]byte(val), &resp)
		return resp, err
	}
}

func FindLatestDoc() (result models.MongoResponse, e error) {

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := db.Conn().Database("testing").Collection("covid")
	opts := options.Find()
	opts.SetSort(bson.D{{"recordDate", -1}})
	opts.SetLimit(1)
	cursor, err := collection.Find(ctx, bson.D{}, opts)
	var record []models.MongoResponse
	if err = cursor.All(ctx, &record); err != nil {
		return models.MongoResponse{}, err
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
	var responseObject models.GeoResponse
	json.Unmarshal(bodyBytes, &responseObject)

	cache, err := isPresentInCache(responseObject.Address.State)

	if cache != (models.UserResponse{}) {
		log.Println("Fetching from cache")
		return c.JSON(http.StatusOK, cache) // Cache hit
	}

	//Continue
	log.Println("Not found in cache")
	res, err := FindLatestDoc()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	} else {
		for _, element := range res.Region {
			if element.Loc == responseObject.Address.State {
				resp := models.UserResponse{element.Loc,
					element.Indiancases, res.Summary.Total, res.LastUpdated}
				marshalled, _ := json.Marshal(resp)
				db.SaveUserResponse(responseObject.Address.State, string(marshalled))
				return c.JSON(http.StatusOK, resp)
			}
		}
	}
	return c.JSON(http.StatusNotFound, models.StateNotFound{responseObject.Address.State, "Not Found"})
}
