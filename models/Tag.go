package models

import (
	"encoding/json"
	"reflect"
)

type Tag struct {
	ID int64 `json:"-" crud:"orm_id"`

	Pid  GUID   `json:"pid" crud:"pid"`
	Name string `json:"name" crud:"name"`
}

type tagMeta struct{}

func (*tagMeta) GUID() GUID {
	return newModelGUID(1)
}

func (*tagMeta) Name() string {
	return "Tag"
}

func (*tagMeta) TableName() string {
	return "Tag"
}

func (*tagMeta) PrimaryKey() string {
	return "pid"
}

func (*tagMeta) Type() reflect.Type {
	return reflect.TypeOf(Tag{})
}

func (*tagMeta) SliceType() reflect.Type {
	return reflect.TypeOf([]Tag{})
}

func (modelMeta *tagMeta) RelationExists(rel string) bool {
	_, _, meta := modelMeta.RelationFieldNames(rel)
	return meta != nil
}

func (*tagMeta) RelationFieldNames(rel string) (string, string, ModelMeta) {

	return "", "", nil
}

func (*tagMeta) BridgeRelationMeta(rel string) (string, string, ModelMeta, ModelMeta) {
	return "", "", nil, nil
}

type wireTag struct {
	ID int64 `json:"-"`

	Pid  GUID   `json:"pid"`
	Name string `json:"name"`
}

func (model *Tag) UnmarshalJSON(data []byte) (er error) {
	var wire wireTag

	if er = json.Unmarshal(data, &wire); er != nil {
		return er
	}

	model.Pid = wire.Pid
	model.Name = wire.Name

	return nil
}

func (model *Tag) MarshalJSON() ([]byte, error) {
	wire := wireTag{
		Pid:  model.Pid,
		Name: model.Name,
	}

	return json.Marshal(wire)
}
