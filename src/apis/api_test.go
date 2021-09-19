package apis

import (
	"InShorts/src/db"
	"InShorts/src/models"
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var (
	region = [1]models.Regional{{
		Loc:           "West Bengal",
		Indiancases:   300,
		Foreigncases:  0,
		Discharged:    0,
		Deaths:        0,
		Totalconfimed: 39875}}
	dbSummary = models.DBSummary{
		Total:      39875,
		Discharged: 0}
	summary = models.Summary{
		Indiancases: 39875,
		Discharged:  0,
	}
	mongoResponse = models.MongoResponse{
		Region:      region[:],
		LastUpdated: "2021-09-16T12:47:21.339Z",
		ModifiedAt:  "2021-09-16 18:33:46",
		RecordDate:  "2021-09-16",
		Summary:     dbSummary,
	}
	address = models.Address{
		State: "West Bengal",
		X:     make(map[string]interface{})}
	geoResponse = models.GeoResponse{
		X:       make(map[string]interface{}),
		Address: address}
	userResponse = models.UserResponse{
		State:             "West Bengal",
		TotalCasesByState: 300,
		TotalCasesInIndia: 39875,
		LastUpdated:       "2021-09-16T12:47:21.339Z",
	}
	data = models.Data{
		Regional: region[:],
		Summary:  summary,
	}
	response = models.Response{
		Success:          true,
		Data:             data,
		LastRefreshed:    "2021-09-16T12:47:21.339Z",
		LastOriginUpdate: "2021-09-16T12:47:21.339Z",
	}
)

type MockedExternalObject struct {
	mock.Mock
}

func (m *MockedExternalObject) GetLatestDoc(dbMethod db.IDBExternal) (models.MongoResponse, error) {
	args := m.Called(dbMethod)
	return args.Get(0).(models.MongoResponse), args.Error(1)
}

func (m *MockedExternalObject) IsPresentInCache(key string, query func(string) (string, error)) (models.UserResponse, error) {
	args := m.Called(key, query)
	return args.Get(0).(models.UserResponse), args.Error(1)
}

func (m *MockedExternalObject) MakeExternalHTTPRequest(client *http.Client, url string) (*http.Response, error) {
	args := m.Called(client, url)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockedExternalObject) SaveToCache(string, string, int) {
}

func (m *MockedExternalObject) Upsert(document bson.D, filter bson.D, today string) (*mongo.UpdateResult, error) {
	args := m.Called(document, filter, today)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *MockedExternalObject) GetIdForToday(today string) string {
	args := m.Called(today)
	return args.String(0)
}

type MockedExternalDBObject struct {
	mock.Mock
}

func (m *MockedExternalDBObject) FindLatestDoc() (*mongo.Cursor, error) {
	args := m.Called()
	return args.Get(0).(*mongo.Cursor), args.Error(1)
}

func TestMongoQuery_EarlyExitIfExternalApiFails(t *testing.T) {

	//Mocks
	//create an instance of our test object
	testObj := new(MockedExternalObject)

	//On-Return pattern (When-Then in Mockito)
	testObj.On("MakeExternalHTTPRequest", mock.AnythingOfType("*http.Client"),
		mock.AnythingOfType("string")).Return(&http.Response{Status: "500"},
		errors.New("External Call Failed"))

	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/data/1.0/1.0", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = LocResults(c, testObj)
	assert := assert.New(t)
	assert.Equal(rec.Code, 500, "Server did not throw error")
	assert.Equal(strings.Trim(rec.Body.String(), "\r\n "), "{\"message\":\"We are having issues retrieving data,"+
		" please check later\"}", "Body does not match")
}

func TestMongoQuery_EarlyExitIfPresentInCache(t *testing.T) {

	//Mocks
	//create an instance of our test object
	testObj := new(MockedExternalObject)

	//On-Return pattern (When-Then in Mockito)
	testObj.On("IsPresentInCache", mock.AnythingOfType("string"), mock.AnythingOfType("func(string) (string, error)")).Return(
		userResponse, nil)
	marshalled, _ := json.Marshal(geoResponse)
	testObj.On("MakeExternalHTTPRequest", mock.AnythingOfType("*http.Client"),
		mock.AnythingOfType("string")).Return(&http.Response{Status: "200",
		Body: ioutil.NopCloser(bytes.NewBufferString(string(marshalled)))}, nil)

	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/data/1.0/1.0", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	res := LocResults(c, testObj)
	assert := assert.New(t)
	assert.Equal(rec.Code, 200, "Server threw error")
	assert.Nil(res)
	marshalled, _ = json.Marshal(userResponse)
	assert.Equal(strings.Trim(rec.Body.String(), "\r\n "), string(marshalled))
}

func TestMongoQuery_ActualQueryIfNotInCache(t *testing.T) {

	testObj := new(MockedExternalObject)

	marshalled, _ := json.Marshal(geoResponse)
	testObj.On("MakeExternalHTTPRequest", mock.AnythingOfType("*http.Client"),
		mock.AnythingOfType("string")).Return(&http.Response{Status: "200",
		Body: ioutil.NopCloser(bytes.NewBufferString(string(marshalled)))}, nil)
	testObj.On("IsPresentInCache", mock.AnythingOfType("string"), mock.AnythingOfType("func(string) (string, error)")).Return(
		models.UserResponse{}, redis.Nil)
	testObj.On("GetLatestDoc", mock.AnythingOfType("db.ODBExternal")).Return(mongoResponse, nil)

	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/data/1.0/1.0", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	res := LocResults(c, testObj)
	assert := assert.New(t)
	assert.Equal(rec.Code, 200, "Server threw error")
	assert.Nil(res)
	marshalled, _ = json.Marshal(userResponse)
	assert.Equal(strings.Trim(rec.Body.String(), "\r\n "), string(marshalled))
}

func TestMongoQuery_ActualQueryIfNotInCache_OusideIndia(t *testing.T) {

	testObj := new(MockedExternalObject)

	geoResponseMod := geoResponse
	geoResponseMod.Address.State = "Jharkhand"
	marshalled, _ := json.Marshal(geoResponseMod)
	testObj.On("MakeExternalHTTPRequest", mock.AnythingOfType("*http.Client"),
		mock.AnythingOfType("string")).Return(&http.Response{Status: "200",
		Body: ioutil.NopCloser(bytes.NewBufferString(string(marshalled)))}, nil)
	testObj.On("IsPresentInCache", mock.AnythingOfType("string"), mock.AnythingOfType("func(string) (string, error)")).Return(
		models.UserResponse{}, redis.Nil)
	testObj.On("GetLatestDoc", mock.AnythingOfType("db.ODBExternal")).Return(mongoResponse, nil)

	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/data/1.0/1.0", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	res := LocResults(c, testObj)
	t.Log(res)
	t.Log(rec)
	assert := assert.New(t)
	assert.Equal(rec.Code, 404, "Server did not throw StateNotFound error")
	assert.Nil(res)
	stateNotFound := models.StateNotFound{
		State:   "Jharkhand",
		Message: "Not Found",
	}
	marshalled, _ = json.Marshal(stateNotFound)
	assert.Equal(strings.Trim(rec.Body.String(), "\r\n "), string(marshalled))
}

func errorFunc(key string) (string, error) {
	return "", errors.New("Test error")
}

func successFunc(key string) (string, error) {
	return "{\"state\":\"West Bengal\",\"totalCasesByState\":1561014,\"totalCasesInIndia\":33448115," +
		"\"lastUpdated\":\"2021-09-19T04:04:19.620Z\"}", nil
}

func TestCacheInteraction(t *testing.T) {
	assert := assert.New(t)
	val, err := OExternal{}.IsPresentInCache("Key", successFunc)
	assert.Nil(err, "Success Function does not Return Nil Error")
	assert.Equal(val.State, "West Bengal", "State does not match")
	val, err = OExternal{}.IsPresentInCache("Key", errorFunc)
	assert.Equal(val, models.UserResponse{}, "Does not return empty object back")
	assert.Equal(err.Error(), "Test error", "Does not propagate error up in hierarchy")
}

func TestMongoUpsert_EarlyExitIfExternalApiFails(t *testing.T) {

	//Mocks
	//create an instance of our test object
	testObj := new(MockedExternalObject)

	//On-Return pattern (When-Then in Mockito)
	testObj.On("MakeExternalHTTPRequest", mock.AnythingOfType("*http.Client"),
		mock.AnythingOfType("string")).Return(&http.Response{Status: "500"},
		errors.New("External Call Failed"))

	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = Fetchcall(c, testObj)
	assert := assert.New(t)
	assert.Equal(rec.Code, 500, "Server did not throw error")
	assert.Equal(strings.Trim(rec.Body.String(), "\r\n "), "{\"message\":\"We are having issues retrieving data,"+
		" please check later\"}", "Body does not match")
}

func TestMongoUpsert_ActualUpsert_Upsert(t *testing.T) {

	//Mocks
	//create an instance of our test object
	testObj := new(MockedExternalObject)

	//On-Return pattern (When-Then in Mockito)
	marshalled, _ := json.Marshal(response)
	testObj.On("MakeExternalHTTPRequest", mock.AnythingOfType("*http.Client"),
		mock.AnythingOfType("string")).Return(&http.Response{Status: "200",
		Body: ioutil.NopCloser(bytes.NewBufferString(string(marshalled)))}, nil)
	testObj.On("Upsert", mock.AnythingOfType("primitive.D"),
		mock.AnythingOfType("primitive.D"), mock.AnythingOfType("string")).Return(
		&mongo.UpdateResult{
			MatchedCount:  1,
			ModifiedCount: 1,
			UpsertedCount: 0,
			UpsertedID:    nil,
		}, nil)
	testObj.On("GetIdForToday", mock.AnythingOfType("string")).Return("6146b7082aba7e84729c2bd7")

	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	res := Fetchcall(c, testObj)
	assert := assert.New(t)
	assert.Equal(rec.Code, 200, "Server threw error")
	assert.Nil(res)
	expectedResponse := models.ApiResponse{
		LatestRecord: "6146b7082aba7e84729c2bd7",
		Inserted:     false, //Since we are returning upserted count 0
		Modified:     true,
	}
	marshalled, _ = json.Marshal(expectedResponse)
	assert.Equal(strings.Trim(rec.Body.String(), "\r\n "), string(marshalled))
}

func TestMongoUpsert_ActualUpsert_Insert(t *testing.T) {

	//Mocks
	//create an instance of our test object
	testObj := new(MockedExternalObject)

	//On-Return pattern (When-Then in Mockito)
	marshalled, _ := json.Marshal(response)
	testObj.On("MakeExternalHTTPRequest", mock.AnythingOfType("*http.Client"),
		mock.AnythingOfType("string")).Return(&http.Response{Status: "200",
		Body: ioutil.NopCloser(bytes.NewBufferString(string(marshalled)))}, nil)
	testObj.On("Upsert", mock.AnythingOfType("primitive.D"),
		mock.AnythingOfType("primitive.D"), mock.AnythingOfType("string")).Return(
		&mongo.UpdateResult{
			MatchedCount:  0,
			ModifiedCount: 0,
			UpsertedCount: 1,
			UpsertedID:    "6146b7082aba7e84729c2bd7",
		}, nil)
	testObj.On("GetIdForToday", mock.AnythingOfType("string")).Return("6146b7082aba7e84729c2bd7")

	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	res := Fetchcall(c, testObj)
	assert := assert.New(t)
	assert.Equal(rec.Code, 200, "Server threw error")
	assert.Nil(res)
	expectedResponse := models.ApiResponse{
		LatestRecord: "6146b7082aba7e84729c2bd7",
		Inserted:     true, //Since we are returning upserted count 1
		Modified:     false,
	}
	marshalled, _ = json.Marshal(expectedResponse)
	assert.Equal(strings.Trim(rec.Body.String(), "\r\n "), string(marshalled))
}

func TestMongoUpsert_ActualUpsert_Insert500(t *testing.T) {

	//Mocks
	//create an instance of our test object
	testObj := new(MockedExternalObject)

	//On-Return pattern (When-Then in Mockito)
	marshalled, _ := json.Marshal(response)
	testObj.On("MakeExternalHTTPRequest", mock.AnythingOfType("*http.Client"),
		mock.AnythingOfType("string")).Return(&http.Response{Status: "200",
		Body: ioutil.NopCloser(bytes.NewBufferString(string(marshalled)))}, nil)
	testObj.On("Upsert", mock.AnythingOfType("primitive.D"),
		mock.AnythingOfType("primitive.D"), mock.AnythingOfType("string")).Return(
		&mongo.UpdateResult{
			MatchedCount:  0,
			ModifiedCount: 0,
			UpsertedCount: 1,
			UpsertedID:    "6146b7082aba7e84729c2bd7",
		}, nil)
	testObj.On("GetIdForToday", mock.AnythingOfType("string")).Return("-")

	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = Fetchcall(c, testObj)
	assert := assert.New(t)
	assert.Equal(rec.Code, 500, "Server did not throw error")
	assert.Equal(strings.Trim(rec.Body.String(), "\r\n "), "{\"message\":\"Unable to find Id of last upserted record\"}")
}

func TestGetLatestDoc_MongoError(t *testing.T) {
	testObj := new(MockedExternalDBObject)
	assert := assert.New(t)
	testObj.On("FindLatestDoc").Return(&mongo.Cursor{}, errors.New("Unable to fetch doc"))
	val, err := OExternal{}.GetLatestDoc(testObj)
	assert.Equal(err.Error(), "Unable to fetch doc", "Error Messages Differ")
	assert.Equal(val, models.MongoResponse{}, "Empty struct not returned on error")
}
