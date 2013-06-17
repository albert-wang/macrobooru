package query

import (
	"fmt"
	"strings"

	"macrobooru/api"
	"macrobooru/models"
)

type QueryRequest struct {
	RequestName   string                 `json:"-"`
	ModelName     string                 `json:"model,omitempty"`
	WhereClauses  map[string]interface{} `json:"where,omitempty"`
	Relation      string                 `json:"relation,omitempty"`
	Subgraph      string                 `json:"subgraph,omitempty"`
	Transient     bool                   `json:"transient,omitempty"`
	Order         string                 `json:"order,omitempty"`
	Limit         int64                  `json:"limit,omitempty"`
	Offset        int64                  `json:"offset,omitempty"`
	SearchClauses map[string]string      `json:"search,omitempty"`
}

func (qr *QueryRequest) Decompose(payload QueryPayload) (*SqlFragment, error) {
	seenSubgraphs := []string{}
	return qr.decompose(payload, &seenSubgraphs)
}

func desugarWhere(clauses map[string]interface{}, model models.ModelMeta) []WhereClause {
	whereClauses := []WhereClause{}

	if val, ok := clauses["#primary"]; ok {
		delete(clauses, "#primary")
		clauses[model.PrimaryKey()] = val
	}

	for field, value := range clauses {
		field = strings.TrimSpace(field)
		fieldParts := strings.Split(field, " ")

		operand := whereEqual
		if len(fieldParts) > 1 {
			operandPart := fieldParts[len(fieldParts)-1]

			switch operandPart {
			case ">":
				operand = whereGreater

			case ">=":
				operand = whereGreaterEqual
			case "<":
				operand = whereLess

			case "<=":
				operand = whereLessEqual

			case "=":
				fallthrough
			case "==":
				operand = whereEqual

			case "!=":
				operand = whereNotEqual

			case "NULL":
				operand = whereNull

			case "NOTNULL":
				operand = whereNotNull
			}
		}

		whereClauses = append(whereClauses, WhereClause{
			Field:   fieldParts[0],
			Value:   value,
			Operand: operand,
		})
	}

	return whereClauses
}

func (qr *QueryRequest) decompose(payload QueryPayload, seenSubgraphs *[]string) (*SqlFragment, error) {
	if qr.Relation != "" {
		if qr.Subgraph != "" {
			return qr.decomposeRelationSubgraph(payload, seenSubgraphs)
		} else {
			return qr.decomposeRelation(payload, seenSubgraphs)
		}

	} else if qr.ModelName != "" {
		return qr.decomposeDirect(payload, seenSubgraphs)
	}

	return nil, api.ErrorInvalidInputFormat("must supply either model or subgraph fields")
}

func nextUnnamedAlias(seenSubgraphs *[]string) string {
	*seenSubgraphs = append(*seenSubgraphs, "")
	return fmt.Sprintf("alias%d", len(*seenSubgraphs))
}

func (qr *QueryRequest) decomposeDirect(payload QueryPayload, seenSubgraphs *[]string) (*SqlFragment, error) {
	model := models.ModelByName(qr.ModelName)
	if model == nil {
		return nil, api.ErrorInvalidInputFormat("model does not exist")
	}

	frag := &SqlFragment{
		Table:  model.TableName(),
		Alias:  nextUnnamedAlias(seenSubgraphs),
		Where:  desugarWhere(qr.WhereClauses, model),
		Limit:  qr.Limit,
		Offset: qr.Offset,
		Target: model,
		/* XXX: ORDER MUST BE CONFIRMED VIA MODEL METADATA */
	}

	return frag, nil
}

func (qr *QueryRequest) decomposeRelation(payload QueryPayload, seenSubgraphs *[]string) (*SqlFragment, error) {
	relModel := models.ModelByName(qr.ModelName)
	if relModel == nil {
		return nil, api.ErrorInvalidInputFormat(fmt.Sprintf("related model (%s) does not exist", qr.ModelName))
	}

	joinField, targetField, model := relModel.RelationFieldNames(qr.Relation)

	if model == nil {
		return nil, api.ErrorInvalidInputFormat(fmt.Sprintf("relation (%s) does not exist", qr.Relation))
	}

	joinFragment, er := qr.decomposeDirect(payload, seenSubgraphs)
	if er != nil {
		return nil, er
	}

	frag := &SqlFragment{
		Table:  model.TableName(),
		Alias:  nextUnnamedAlias(seenSubgraphs),
		Limit:  qr.Limit,
		Offset: qr.Offset,
		Target: model,
		/* XXX: ORDER */
	}

	frag.JoinOn(joinFragment, targetField, joinField)

	return frag, nil
}

func (qr *QueryRequest) decomposeRelationSubgraph(payload QueryPayload, seenSubgraphs *[]string) (*SqlFragment, error) {
	subgraphReq, ok := payload.GetNamedRequest(qr.Subgraph)
	if !ok {
		return nil, api.ErrorInvalidInputFormat(fmt.Sprintf("referenced subgraph (%s) does not exist", qr.Subgraph))
	}

	subgraphJoin, er := subgraphReq.decompose(payload, seenSubgraphs)
	if er != nil {
		return nil, er
	}

	subgraphModel := models.ModelByName(subgraphReq.ModelName)

	/* Cycle check */
	for _, seen := range *seenSubgraphs {
		if seen == qr.Subgraph {
			return nil, api.ErrorInvalidInputFormat("cyclic subgraph reference")
		}
	}
	*seenSubgraphs = append(*seenSubgraphs, qr.Subgraph)
	alias := fmt.Sprintf("alias%d", len(*seenSubgraphs))

	if joinField, foreignField, joinModel, model := subgraphModel.BridgeRelationMeta(qr.Relation); model != nil {
		bridgeFrag := &SqlFragment{
			Table:  joinModel.TableName(),
			Alias:  nextUnnamedAlias(seenSubgraphs),
			Target: joinModel,
		}

		bridgeFrag.JoinOn(subgraphJoin, foreignField, subgraphModel.PrimaryKey())

		frag := &SqlFragment{
			Table:  model.TableName(),
			Alias:  alias,
			Where:  desugarWhere(qr.WhereClauses, model),
			Limit:  qr.Limit,
			Offset: qr.Offset,
			Target: model,
		}

		frag.JoinOn(bridgeFrag, model.PrimaryKey(), joinField)
		return frag, nil

	} else if joinField, targetField, model := subgraphModel.RelationFieldNames(qr.Relation); model != nil {
		frag := &SqlFragment{
			Table:  model.TableName(),
			Alias:  alias,
			Where:  desugarWhere(qr.WhereClauses, model),
			Limit:  qr.Limit,
			Offset: qr.Offset,
			Target: model,
		}

		frag.JoinOn(subgraphJoin, targetField, joinField)
		return frag, nil
	}

	return nil, api.ErrorInvalidInputFormat("Don't understand how to process request")
}
