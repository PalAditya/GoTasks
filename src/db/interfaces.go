package db

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type IDBExternal interface {
	FindLatestDoc() (*mongo.Cursor, error)
}

type ODBExternal struct {
}
