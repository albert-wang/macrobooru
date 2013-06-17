package query

import (
	"fmt"
	"macrobooru/api"
	"macrobooru/models"

	"github.com/lye/crud"

	"database/sql"
	"encoding/json"
	"reflect"
)

type QueryPayload map[string]QueryRequest

func (*QueryPayload) Name() string {
	return "query"
}

func (qp *QueryPayload) requests() map[string]QueryRequest {
	tmp := (*map[string]QueryRequest)(qp)
	return *tmp
}

func (*QueryPayload) Parse(req *api.RequestWrapper) (api.Operation, error) {
	var payload QueryPayload

	if er := json.Unmarshal(req.RawData, &payload); er != nil {
		return nil, api.ErrorInvalidInputFormat(er.Error())
	}

	return &payload, nil
}

func (*QueryPayload) ParseResponse(wrapper api.ResponseWrapper) (interface{}, error) {
	res := QueryResponse{}

	if er := json.Unmarshal(wrapper.Data, &res); er != nil {
		return nil, api.ErrorGeneric(er)
	}

	return res, nil
}

func (qp *QueryPayload) Process(u *models.User, db *sql.DB) (interface{}, error) {
	responseMap := map[string]QueryResponsePart{}

	for name, req := range qp.requests() {
		if req.Transient {
			continue
		}

		sqlFrag, apiEr := req.Decompose(*qp)
		if apiEr != nil {
			return nil, apiEr
		}

		sql, params := sqlFrag.toCountSQL()
		rows, er := db.Query(sql, params...)
		if er != nil {
			return nil, api.ErrorGeneric(er)
		}
		defer rows.Close()

		var totalCount int64

		if rows.Next() {
			if er := rows.Scan(&totalCount); er != nil {
				return nil, api.ErrorGeneric(er)
			}
		}

		zeroVal := reflect.Zero(sqlFrag.Target.Type()).Interface()
		slice := []interface{}{}

		if totalCount > 0 {
			sql, params = sqlFrag.toSQL()
			fmt.Printf("C: %d\nQ: %s\nA: %#v\n", totalCount, sql, params)

			rows, er := db.Query(sql, params...)
			if er != nil {
				return nil, api.ErrorGeneric(er)
			}
			defer rows.Close()

			for rows.Next() {
				if er := crud.Scan(rows, &zeroVal); er != nil {
					return nil, api.ErrorGeneric(er)
				}

				slice = append(slice, zeroVal)
			}
		}

		responseMap[name] = QueryResponsePart{
			Total:     totalCount,
			Slice:     slice,
			ModelName: sqlFrag.Target.Type().Name(),
		}
	}

	return QueryResponse(responseMap), nil
}

func (qp *QueryPayload) GetNamedRequest(name string) (QueryRequest, bool) {
	if qp == nil {
		return QueryRequest{}, false
	}

	ret, ok := (*(*map[string]QueryRequest)(qp))[name]
	return ret, ok
}
