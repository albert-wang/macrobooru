package models

import (
	"encoding/json"
	"reflect"
)

type TagBridge struct {
	ID int64 `json:"-" crud:"orm_id"`

	Pid      GUID `json:"pid" crud:"pid"`
	Image_id GUID `json:"image_id" crud:"image_id"`
	Tag_id   GUID `json:"tag_id" crud:"tag_id"`
}

type tagBridgeMeta struct{}

func (*tagBridgeMeta) GUID() GUID {
	return newModelGUID(3)
}

func (*tagBridgeMeta) Name() string {
	return "TagBridge"
}

func (*tagBridgeMeta) TableName() string {
	return "TagBridge"
}

func (*tagBridgeMeta) PrimaryKey() string {
	return "pid"
}

func (*tagBridgeMeta) Type() reflect.Type {
	return reflect.TypeOf(TagBridge{})
}

func (*tagBridgeMeta) SliceType() reflect.Type {
	return reflect.TypeOf([]TagBridge{})
}

func (modelMeta *tagBridgeMeta) RelationExists(rel string) bool {
	_, _, meta := modelMeta.RelationFieldNames(rel)
	return meta != nil
}

func (*tagBridgeMeta) RelationFieldNames(rel string) (string, string, ModelMeta) {

	return "", "", nil
}

func (*tagBridgeMeta) BridgeRelationMeta(rel string) (string, string, ModelMeta, ModelMeta) {

	return "", "", nil, nil
}

type wireTagBridge struct {
	ID int64 `json:"-"`

	Pid      GUID `json:"pid"`
	Image_id GUID `json:"image_id"`
	Tag_id   GUID `json:"tag_id"`
}

func (model *TagBridge) UnmarshalJSON(data []byte) (er error) {
	var wire wireTagBridge

	if er = json.Unmarshal(data, &wire); er != nil {
		return er
	}

	model.Pid = wire.Pid
	model.Image_id = wire.Image_id
	model.Tag_id = wire.Tag_id

	return nil
}

func (model *TagBridge) MarshalJSON() ([]byte, error) {
	wire := wireTagBridge{
		Pid:      model.Pid,
		Image_id: model.Image_id,
		Tag_id:   model.Tag_id,
	}

	return json.Marshal(wire)
}
