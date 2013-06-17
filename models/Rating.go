package models

import (
	"encoding/json"
	"reflect"
)

type Rating struct {
	ID int64 `json:"-" crud:"orm_id"`

	Pid        GUID   `json:"pid" crud:"pid"`
	Image_id   GUID   `json:"image_id" crud:"image_id"`
	Rating     int32  `json:"rating" crud:"rating"`
	RaterEmail string `json:"raterEmail" crud:"raterEmail"`
}

type ratingMeta struct{}

func (*ratingMeta) GUID() GUID {
	return newModelGUID(6)
}

func (*ratingMeta) Name() string {
	return "Rating"
}

func (*ratingMeta) TableName() string {
	return "Rating"
}

func (*ratingMeta) PrimaryKey() string {
	return "pid"
}

func (*ratingMeta) Type() reflect.Type {
	return reflect.TypeOf(Rating{})
}

func (*ratingMeta) SliceType() reflect.Type {
	return reflect.TypeOf([]Rating{})
}

func (modelMeta *ratingMeta) RelationExists(rel string) bool {
	_, _, meta := modelMeta.RelationFieldNames(rel)
	return meta != nil
}

func (*ratingMeta) RelationFieldNames(rel string) (string, string, ModelMeta) {

	return "", "", nil
}

func (*ratingMeta) BridgeRelationMeta(rel string) (string, string, ModelMeta, ModelMeta) {

	return "", "", nil, nil
}

type wireRating struct {
	ID int64 `json:"-"`

	Pid        GUID   `json:"pid"`
	Image_id   GUID   `json:"image_id"`
	Rating     int32  `json:"rating"`
	RaterEmail string `json:"raterEmail"`
}

func (model *Rating) UnmarshalJSON(data []byte) (er error) {
	var wire wireRating

	if er = json.Unmarshal(data, &wire); er != nil {
		return er
	}

	model.Pid = wire.Pid
	model.Image_id = wire.Image_id
	model.Rating = wire.Rating
	model.RaterEmail = wire.RaterEmail

	return nil
}

func (model *Rating) MarshalJSON() ([]byte, error) {
	wire := wireRating{
		Pid:        model.Pid,
		Image_id:   model.Image_id,
		Rating:     model.Rating,
		RaterEmail: model.RaterEmail,
	}

	return json.Marshal(wire)
}
