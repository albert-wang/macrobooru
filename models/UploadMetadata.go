package models

import (
	"encoding/json"
	"reflect"
)

type UploadMetadata struct {
	ID int64 `json:"-" crud:"orm_id"`

	Pid               GUID   `json:"pid" crud:"pid"`
	ImageGUID         GUID   `json:"imageGUID" crud:"imageGUID"`
	UploadedBy        string `json:"uploadedBy" crud:"uploadedBy"`
	OriginalExtension string `json:"originalExtension" crud:"originalExtension"`
}

type uploadMetadataMeta struct{}

func (*uploadMetadataMeta) GUID() GUID {
	return newModelGUID(5)
}

func (*uploadMetadataMeta) Name() string {
	return "UploadMetadata"
}

func (*uploadMetadataMeta) TableName() string {
	return "UploadMetadata"
}

func (*uploadMetadataMeta) PrimaryKey() string {
	return "pid"
}

func (*uploadMetadataMeta) Type() reflect.Type {
	return reflect.TypeOf(UploadMetadata{})
}

func (*uploadMetadataMeta) SliceType() reflect.Type {
	return reflect.TypeOf([]UploadMetadata{})
}

func (modelMeta *uploadMetadataMeta) RelationExists(rel string) bool {
	_, _, meta := modelMeta.RelationFieldNames(rel)
	return meta != nil
}

func (*uploadMetadataMeta) RelationFieldNames(rel string) (string, string, ModelMeta) {

	return "", "", nil
}

func (*uploadMetadataMeta) BridgeRelationMeta(rel string) (string, string, ModelMeta, ModelMeta) {

	return "", "", nil, nil
}

type wireUploadMetadata struct {
	ID int64 `json:"-"`

	Pid               GUID   `json:"pid"`
	ImageGUID         GUID   `json:"imageGUID"`
	UploadedBy        string `json:"uploadedBy"`
	OriginalExtension string `json:"originalExtension"`
}

func (model *UploadMetadata) UnmarshalJSON(data []byte) (er error) {
	var wire wireUploadMetadata

	if er = json.Unmarshal(data, &wire); er != nil {
		return er
	}

	model.Pid = wire.Pid
	model.ImageGUID = wire.ImageGUID
	model.UploadedBy = wire.UploadedBy
	model.OriginalExtension = wire.OriginalExtension

	return nil
}

func (model *UploadMetadata) MarshalJSON() ([]byte, error) {
	wire := wireUploadMetadata{
		Pid:               model.Pid,
		ImageGUID:         model.ImageGUID,
		UploadedBy:        model.UploadedBy,
		OriginalExtension: model.OriginalExtension,
	}

	return json.Marshal(wire)
}
