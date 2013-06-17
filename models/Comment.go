package models

import (
	"encoding/json"
	"reflect"

	"time"
)

type Comment struct {
	ID int64 `json:"-" crud:"orm_id"`

	Pid         GUID      `json:"pid" crud:"pid"`
	Parent_id   GUID      `json:"parent_id" crud:"parent_id"`
	DateCreated time.Time `json:"dateCreated" crud:"dateCreated,unix"`
	Contents    string    `json:"contents" crud:"contents"`
}

type commentMeta struct{}

func (*commentMeta) GUID() GUID {
	return newModelGUID(2)
}

func (*commentMeta) Name() string {
	return "Comment"
}

func (*commentMeta) TableName() string {
	return "Comment"
}

func (*commentMeta) PrimaryKey() string {
	return "pid"
}

func (*commentMeta) Type() reflect.Type {
	return reflect.TypeOf(Comment{})
}

func (*commentMeta) SliceType() reflect.Type {
	return reflect.TypeOf([]Comment{})
}

func (modelMeta *commentMeta) RelationExists(rel string) bool {
	_, _, meta := modelMeta.RelationFieldNames(rel)
	return meta != nil
}

func (*commentMeta) RelationFieldNames(rel string) (string, string, ModelMeta) {

	return "", "", nil
}

func (*commentMeta) BridgeRelationMeta(rel string) (string, string, ModelMeta, ModelMeta) {

	return "", "", nil, nil
}

type wireComment struct {
	ID int64 `json:"-"`

	Pid         GUID   `json:"pid"`
	Parent_id   GUID   `json:"parent_id"`
	DateCreated string `json:"dateCreated"`
	Contents    string `json:"contents"`
}

func (model *Comment) UnmarshalJSON(data []byte) (er error) {
	var wire wireComment

	if er = json.Unmarshal(data, &wire); er != nil {
		return er
	}

	model.Pid = wire.Pid
	model.Parent_id = wire.Parent_id
	if model.DateCreated, er = parseTimestamp(wire.DateCreated); er != nil {
		return er
	}
	model.Contents = wire.Contents

	return nil
}

func (model *Comment) MarshalJSON() ([]byte, error) {
	wire := wireComment{
		Pid:         model.Pid,
		Parent_id:   model.Parent_id,
		DateCreated: unparseTimestamp(model.DateCreated),
		Contents:    model.Contents,
	}

	return json.Marshal(wire)
}
