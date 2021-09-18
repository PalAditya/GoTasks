package apis

import (
	"InShorts/src/db"
	"InShorts/src/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

func isPresentInCache(key string) (userResponse models.UserResponse, e error) {
	val, err := db.IsPresentInCache(key)
	if err == nil {
		var resp models.UserResponse
		json.Unmarshal([]byte(val), &resp)
		return resp, err
	} else {
		log.Println(err.Error())
		return models.UserResponse{}, err
	}
}

func getLatestDoc() (result models.MongoResponse, e error) {

	cursor, err := db.FindLatestDoc()
	if err != nil {
		log.Println("Unable to fetch latest doc from Mongo")
		return models.MongoResponse{}, err
	}
	ctx := db.GetCTX()
	var record []models.MongoResponse
	if err = cursor.All(ctx, &record); err != nil {
		return models.MongoResponse{}, err
	}
	return record[0], err
}

func retrieveKey() string {

	if _, err := os.Stat("../../key.txt"); err == nil { // Useful locally
		b, err := ioutil.ReadFile("../../key.txt")
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

	if err == nil {
		log.Println("Fetched from cache")
		return c.JSON(http.StatusOK, cache) // Cache hit
	}

	//Continue
	log.Printf("Key %s Not found in cache\n", responseObject.Address.State)
	res, err := getLatestDoc()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorMessage{
			Message: "We are having issues retrieving data, please check later"})
	} else {
		for _, element := range res.Region {
			if element.Loc == responseObject.Address.State {
				resp := models.UserResponse{
					State:             element.Loc,
					TotalCasesByState: element.Indiancases,
					TotalCasesInIndia: res.Summary.Total,
					LastUpdated:       res.LastUpdated}
				marshalled, _ := json.Marshal(resp)
				db.SaveToCache(responseObject.Address.State, string(marshalled), 30)
				return c.JSON(http.StatusOK, resp)
			}
		}
	}
	return c.JSON(http.StatusNotFound, models.StateNotFound{
		State:   responseObject.Address.State,
		Message: "Not Found"})
}
