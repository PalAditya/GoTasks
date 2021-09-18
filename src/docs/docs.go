// Package classification covid.
//
// Documentation of our APIs
//
//     Schemes: http
//     BasePath: /
//     Version: 1.0.0
//     Host: localhost:1323
//
//
//     Produces:
//     - application/json
//
//
// swagger:meta
package docs

import (
	"InShorts/src/models"
)

// swagger:route GET /api covid saveToDB
// Makes an API call to https://api.rootnet.in/covid19-in/stats/latest, marshalls the data and persists in DB.
// responses:
//   200: SuccessResponse
//   500: ErrorResponse

// Wraps the default Mongo Update Structure to show the id of the document and whether it was an insert or not. This seems like an Admin API so exposing the Id is fine
// swagger:response SuccessResponse
type ApiSucccessResponseWrapper struct {
	// in:body
	Body models.ApiResponse
}

// Returns a generic error message since Mongo Update Errors should count as Internal Server Error
// swagger:response ErrorResponse
type ApiErrorResponseWrapper struct {
	// in:body
	Body models.ErrorMessage
}

// swagger:route GET /data/{lat}/{long} covid stateResults
// Takes in latitude and longitude as parameters and returns data for that state, if within India
// responses:
//   200: SuccessResponseForFetch
//   404: ErrorResponseForFetch

//swagger:parameters stateResults
type LatRequest struct {
	//in: path
	Latitude string `json:"lat"`
}

//swagger:parameters stateResults
type LongRequest struct {
	//in: path
	Longitude string `json:"long"`
}

// Returns total Covid Cases for the state, in India, and the last time this data was fetched. It includes all cured cases and not just 'active' ones
// swagger:response SuccessResponseForFetch
type ApiSucccessResponseForFetchWrapper struct {
	// in:body
	Body models.UserResponse
}

// Returns the state and the fact that it was not found within database
// swagger:response ErrorResponseForFetch
type ApiErrorResponseForFetchWrapper struct {
	// in:body
	Body models.StateNotFound
}
