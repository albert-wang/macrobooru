package models

import (
	"encoding/json"
	"reflect"

	"time"
)

type Image struct {
	ID int64 `json:"-" crud:"orm_id"`

	Pid            GUID      `json:"pid" crud:"pid"`
	Filehash       string    `json:"filehash" crud:"filehash"`
	Mime           string    `json:"mime" crud:"mime"`
	UploadedDate   time.Time `json:"uploadedDate" crud:"uploadedDate,unix"`
	RatingsAverage float32   `json:"ratingsAverage" crud:"ratingsAverage"`
}

type imageMeta struct{}

func (*imageMeta) GUID() GUID {
	return newModelGUID(4)
}

func (*imageMeta) Name() string {
	return "Image"
}

func (*imageMeta) TableName() string {
	return "Image"
}

func (*imageMeta) PrimaryKey() string {
	return "pid"
}

func (*imageMeta) Type() reflect.Type {
	return reflect.TypeOf(Image{})
}

func (*imageMeta) SliceType() reflect.Type {
	return reflect.TypeOf([]Image{})
}

func (modelMeta *imageMeta) RelationExists(rel string) bool {
	_, _, meta := modelMeta.RelationFieldNames(rel)
	return meta != nil
}

func (*imageMeta) RelationFieldNames(rel string) (string, string, ModelMeta) {
	if rel == "comments" {
		return "comments", "parent_id", CommentMeta
	}
	if rel == "ratings" {
		return "ratings", "image_id", RatingMeta
	}

	return "", "", nil
}

func (*imageMeta) BridgeRelationMeta(rel string) (string, string, ModelMeta, ModelMeta) {
	return "", "", nil, nil
}

type wireImage struct {
	ID int64 `json:"-"`

	Pid            GUID    `json:"pid"`
	Filehash       string  `json:"filehash"`
	Mime           string  `json:"mime"`
	UploadedDate   string  `json:"uploadedDate"`
	RatingsAverage float32 `json:"ratingsAverage"`
}

func (model *Image) UnmarshalJSON(data []byte) (er error) {
	var wire wireImage

	if er = json.Unmarshal(data, &wire); er != nil {
		return er
	}

	model.Pid = wire.Pid
	model.Filehash = wire.Filehash
	model.Mime = wire.Mime
	if model.UploadedDate, er = parseTimestamp(wire.UploadedDate); er != nil {
		return er
	}
	model.RatingsAverage = wire.RatingsAverage

	return nil
}

func (model *Image) MarshalJSON() ([]byte, error) {
	wire := wireImage{
		Pid:            model.Pid,
		Filehash:       model.Filehash,
		Mime:           model.Mime,
		UploadedDate:   unparseTimestamp(model.UploadedDate),
		RatingsAverage: model.RatingsAverage,
	}

	return json.Marshal(wire)
}
