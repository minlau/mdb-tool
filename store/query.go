package store

import (
	"github.com/segmentio/encoding/json"
)

type GroupQueryResult struct {
	GroupName string      `json:"groupName"`
	Data      *QueryData  `json:"data"`
	Error     *QueryError `json:"error"`
}

type QueryData struct {
	Columns []Column                 `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
}

// Column is used to store original column name(Name) and custom name(FieldName) for json response.
// This is required, because sql row can contain multiple columns with same name
type Column struct {
	Name      string `json:"name"`
	FieldName string `json:"fieldName"`
}

type QueryError struct {
	Message string `json:"message"`
	Err     error  `json:"err"`
}

func NewQueryError(err error) *QueryError {
	if err == nil {
		return nil
	}
	if errJson, _ := json.Marshal(err); string(errJson) == "{}" {
		return &QueryError{Message: err.Error(), Err: nil}
	}
	return &QueryError{Message: err.Error(), Err: err}
}
