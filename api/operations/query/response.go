package query

import (
	"encoding/json"
	"fmt"
	"macrobooru/models"
	"reflect"
)

type QueryResponse map[string]QueryResponsePart

type QueryResponsePart struct {
	Total     int64         `json:"total"`
	Slice     []interface{} `json:"slice"`
	ModelName string        `json:"model,omitempty"`
}

type rawQueryResponsePart struct {
	Total     int64             `json:"total"`
	Slice     []json.RawMessage `json:"slice"`
	ModelName string            `json:"model"`
}

func (part *QueryResponsePart) UnmarshalJSON(bs []byte) error {
	rawPart := rawQueryResponsePart{}

	if er := json.Unmarshal(bs, &rawPart); er != nil {
		return er
	}

	modelMeta := models.ModelByName(rawPart.ModelName)
	if modelMeta == nil {
		return fmt.Errorf("No such model (%s)", rawPart.ModelName)
	}

	slice := []interface{}{}

	for _, msg := range rawPart.Slice {
		zeroVal := reflect.New(modelMeta.Type())

		if er := json.Unmarshal(msg, zeroVal.Interface()); er != nil {
			return er
		}

		slice = append(slice, zeroVal.Elem().Interface())
	}

	part.Total = rawPart.Total
	part.Slice = slice
	part.ModelName = rawPart.ModelName
	return nil
}
