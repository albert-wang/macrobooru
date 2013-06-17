package models

import (
	"encoding/json"
	"reflect"

	"time"
)

type PointOfInterest struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

func (poi *PointOfInterest) UnmarshalJSON(bs []byte) error {
	naive := map[string]interface{}{
		"x": 0,
		"y": 0,
	}

	if er := json.Unmarshal(bs, &naive); er == nil {
		poi.X = naive["x"].(int64)
		poi.Y = naive["y"].(int64)
	}

	return nil
}

type Static struct {
	ID int64 `json:"-" crud:"orm_id"`

	Pid                  GUID            `json:"pid" crud:"pid"`
	LastUpdatedTimestamp time.Time       `json:"lastUpdatedTimestamp" crud:"lastUpdatedTimestamp,unix"`
	Mime                 string          `json:"mime" crud:"mime"`
	Path                 string          `json:"path" crud:"path"`
	SHA1Hash             string          `json:"SHA1Hash" crud:"SHA1Hash"`
	Thumb                PointOfInterest `json:"thumb" crud:"thumb"`
}

type staticMeta struct{}

func (*staticMeta) GUID() GUID {
	return newModelGUID(7)
}

func (*staticMeta) Name() string {
	return "Static"
}

func (*staticMeta) TableName() string {
	return "Static"
}

func (*staticMeta) PrimaryKey() string {
	return "pid"
}

func (*staticMeta) Type() reflect.Type {
	return reflect.TypeOf(Static{})
}

func (*staticMeta) SliceType() reflect.Type {
	return reflect.TypeOf([]Static{})
}

func (modelMeta *staticMeta) RelationExists(rel string) bool {
	_, _, meta := modelMeta.RelationFieldNames(rel)
	return meta != nil
}

func (*staticMeta) RelationFieldNames(rel string) (string, string, ModelMeta) {

	return "", "", nil
}

func (*staticMeta) BridgeRelationMeta(rel string) (string, string, ModelMeta, ModelMeta) {

	return "", "", nil, nil
}

type wireStatic struct {
	ID int64 `json:"-"`

	Pid                  GUID            `json:"pid"`
	LastUpdatedTimestamp *string         `json:"lastUpdatedTimestamp"`
	Mime                 string          `json:"mime"`
	Path                 string          `json:"path"`
	SHA1Hash             string          `json:"SHA1Hash"`
	Thumb                PointOfInterest `json:"thumb" crud:"thumb"`
}

func (model *Static) UnmarshalJSON(data []byte) (er error) {
	var wire wireStatic

	if er = json.Unmarshal(data, &wire); er != nil {
		return er
	}

	model.Pid = wire.Pid
	if wire.LastUpdatedTimestamp != nil {
		if model.LastUpdatedTimestamp, er = parseTimestamp(*wire.LastUpdatedTimestamp); er != nil {
			return er
		}
	}
	model.Mime = wire.Mime
	model.Path = wire.Path
	model.SHA1Hash = wire.SHA1Hash
	model.Thumb = wire.Thumb

	return nil
}

func (model *Static) MarshalJSON() ([]byte, error) {
	wire := wireStatic{
		Pid:                  model.Pid,
		LastUpdatedTimestamp: unparseTimestampOptional(model.LastUpdatedTimestamp),
		Mime:                 model.Mime,
		Path:                 model.Path,
		SHA1Hash:             model.SHA1Hash,
		Thumb:                model.Thumb,
	}

	return json.Marshal(wire)
}
