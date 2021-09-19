package apis

import (
	"InShorts/src/models"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type IExternal interface {
	GetLatestDoc() (models.MongoResponse, error)
	MakeExternalHTTPRequest(client *http.Client, url string) (*http.Response, error)
	SaveToCache(string, string, int)
	IsPresentInCache(key string, query func(string) (string, error)) (models.UserResponse, error)
	Upsert(document bson.D, filter bson.D, today string) (*mongo.UpdateResult, error)
	GetIdForToday(today string) string
}

type OExternal struct {
}

func (externalClient OExternal) MakeExternalHTTPRequest(client *http.Client,
	url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Print(err.Error())
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	return client.Do(req)
}
