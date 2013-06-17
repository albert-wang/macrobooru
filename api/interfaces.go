package api

import (
	"database/sql"
	"macrobooru/models"
)

type Operation interface {
	// Name of this operation, as specified in the API.
	Name() string

	// Parses the wrapper's .RawData into an an Operation instance.
	Parse(*RequestWrapper) (Operation, error)

	// Runs the operation with the specified (potentially nil) user and database.
	Process(*models.User, *sql.DB) (interface{}, error)

	// Parses the wrapper's .Data into an operation-specific format.
	ParseResponse(ResponseWrapper) (interface{}, error)
}
