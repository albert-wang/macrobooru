package client

import (
	"fmt"
	. "macrobooru/api/operations/query"
	"macrobooru/models"
	"reflect"
)

type Query struct {
	requests      map[string]*QueryDetails
	deferredError error
}

func NewQuery() *Query {
	return &Query{
		requests: map[string]*QueryDetails{},
	}
}

func (query *Query) modelMetaForOut(out interface{}) (models.ModelMeta, error) {
	if out == nil {
		return nil, fmt.Errorf("query: must pass a model name or slice of models to each query")
	}

	ty := reflect.TypeOf(out)
	val := reflect.ValueOf(out)

	if ty.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("query: must pass pointer-to-slice of models (got non-ptr)")
	}

	ty = ty.Elem()
	val = val.Elem()

	if ty.Kind() != reflect.Slice {
		return nil, fmt.Errorf("query: must pass pointer-to-slice of models (got non-slice)")
	}

	ty = ty.Elem()
	modelName := ty.Name()
	modelMeta := models.ModelByName(modelName)

	if modelMeta == nil {
		return nil, fmt.Errorf("query: no such model (%s)", modelName)
	}

	return modelMeta, nil
}

func (query *Query) Add(name string, out interface{}) *QueryDetails {
	details := QueryDetails{}

	var er error
	var modelMeta models.ModelMeta
	var transient bool

	if name == "" {
		name = fmt.Sprintf("___auto_%d", len(query.requests))
	}

	if modelMeta, transient = out.(models.ModelMeta); !transient {
		modelMeta, er = query.modelMetaForOut(out)

		if er != nil {
			query.deferredError = er
			return &details
		}
	}

	req := QueryRequest{
		RequestName:   name,
		ModelName:     modelMeta.Name(),
		WhereClauses:  map[string]interface{}{},
		SearchClauses: map[string]string{},
		Transient:     transient,
	}

	details = QueryDetails{
		QueryRequest: req,
		modelMeta:    modelMeta,
		outSlicePtr:  out,
	}

	query.requests[name] = &details
	return &details
}

func (query *Query) Subgraph(name, joinSubgraph, subgraphRelation string, out interface{}) *QueryDetails {
	details := QueryDetails{}

	if name == "" {
		name = fmt.Sprintf("___auto_%d", len(query.requests))
	}

	var modelMeta models.ModelMeta
	var er error
	var transient bool

	if modelMeta, transient = out.(models.ModelMeta); !transient {
		modelMeta, er = query.modelMetaForOut(out)

		if er != nil {
			query.deferredError = er
			return &details
		}
	}

	/* XXX: Use modelMeta to verify this subgraph exists */

	req := QueryRequest{
		RequestName:   name,
		WhereClauses:  map[string]interface{}{},
		SearchClauses: map[string]string{},
		Subgraph:      joinSubgraph,
		Relation:      subgraphRelation,
		Transient:     transient,
	}

	details = QueryDetails{
		QueryRequest: req,
		modelMeta:    modelMeta,
		outSlicePtr:  out,
	}

	query.requests[name] = &details
	return &details
}

func (query *Query) buildPayload() QueryPayload {
	payloadMap := map[string]QueryRequest{}

	for name, details := range query.requests {
		payloadMap[name] = details.QueryRequest
	}

	return QueryPayload(payloadMap)
}

func (query *Query) Execute(client *Client) error {
	if query.deferredError != nil {
		return query.deferredError
	}

	payload := query.buildPayload()

	resWrapper, er := client.Execute(&payload, nil)
	if er != nil {
		query.deferredError = er
		return er
	}

	/* XXX: Marshal outmap */
	rawResponse, er := payload.ParseResponse(*resWrapper)
	if er != nil {
		query.deferredError = er
		return er
	}

	resp := rawResponse.(QueryResponse)

	for name, part := range resp {
		req, ok := query.requests[name]
		if !ok {
			continue
		}

		if req.Transient {
			continue
		}

		outSlicePtr := reflect.ValueOf(req.outSlicePtr)
		outSliceVal := reflect.Zero(outSlicePtr.Elem().Type())

		for _, model := range part.Slice {
			modelVal := reflect.ValueOf(model)
			outSliceVal = reflect.Append(outSliceVal, modelVal)
		}

		if req.outTotalPtr != nil {
			*req.outTotalPtr = part.Total
		}

		outSlicePtr.Elem().Set(outSliceVal)
	}

	return nil
}

func (query *Query) Error() error {
	return query.deferredError
}

type QueryDetails struct {
	QueryRequest

	modelMeta   models.ModelMeta
	outSlicePtr interface{}
	outTotalPtr *int64
}

func (details *QueryDetails) Where(fields map[string]interface{}) *QueryDetails {
	if details.WhereClauses != nil {
		for k, v := range fields {
			details.WhereClauses[k] = v
		}
	}

	return details
}

func (details *QueryDetails) TextSearch(fields map[string]string) *QueryDetails {
	if details.SearchClauses != nil {
		for k, v := range fields {
			details.SearchClauses[k] = v
		}
	}

	return details
}

func (details *QueryDetails) Paginate(order string, page, perPage int64) *QueryDetails {
	details.Order = order
	details.Limit = perPage
	details.Offset = page * perPage
	return details
}

func (details *QueryDetails) Total(out *int64) *QueryDetails {
	details.outTotalPtr = out
	return details
}
