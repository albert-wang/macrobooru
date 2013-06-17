package models

import (
	"encoding/json"
	"reflect"

	"time"
)

type User struct {
	ID int64 `json:"-" crud:"orm_id"`

	LastUpdatedTimestamp time.Time `json:"lastUpdatedTimestamp" crud:"lastUpdatedTimestamp,unix"`
	Pid                  GUID      `json:"pid" crud:"pid"`
	Email                string    `json:"email" crud:"email"`
	Username             string    `json:"username" crud:"username"`
	DisplayName          string    `json:"displayName" crud:"displayName"`
	Location             string    `json:"location" crud:"location"`
	Image                GUID      `json:"image" crud:"image"`
	Passhash             string    `json:"passhash" crud:"passhash"`
	IsSuspended          int32     `json:"isSuspended" crud:"isSuspended"`
	IsAdmin              int32     `json:"isAdmin" crud:"isAdmin"`
	TwitterID            *string   `json:"twitterID" crud:"twitterID"`
	GoogleID             *string   `json:"googleID" crud:"googleID"`
	FacebookID           *string   `json:"facebookID" crud:"facebookID"`
}

type userMeta struct{}

func (*userMeta) GUID() GUID {
	return newModelGUID(6)
}

func (*userMeta) Name() string {
	return "User"
}

func (*userMeta) TableName() string {
	return "User"
}

func (*userMeta) PrimaryKey() string {
	return "pid"
}

func (*userMeta) Type() reflect.Type {
	return reflect.TypeOf(User{})
}

func (*userMeta) SliceType() reflect.Type {
	return reflect.TypeOf([]User{})
}

func (modelMeta *userMeta) RelationExists(rel string) bool {
	_, _, meta := modelMeta.RelationFieldNames(rel)
	return meta != nil
}

func (*userMeta) RelationFieldNames(rel string) (string, string, ModelMeta) {
	return "", "", nil
}

func (*userMeta) BridgeRelationMeta(rel string) (string, string, ModelMeta, ModelMeta) {

	return "", "", nil, nil
}

type wireUser struct {
	ID int64 `json:"-"`

	LastUpdatedTimestamp *string `json:"lastUpdatedTimestamp"`
	Pid                  GUID    `json:"pid"`
	Email                string  `json:"email"`
	Username             string  `json:"username"`
	DisplayName          string  `json:"displayName"`
	Location             string  `json:"location"`
	Image                GUID    `json:"image"`
	Passhash             string  `json:"passhash"`
	IsSuspended          int32   `json:"isSuspended"`
	IsAdmin              int32   `json:"isAdmin"`
	TwitterID            *string `json:"twitterID"`
	GoogleID             *string `json:"googleID"`
	FacebookID           *string `json:"facebookID"`
}

func (model *User) UnmarshalJSON(data []byte) (er error) {
	var wire wireUser

	if er = json.Unmarshal(data, &wire); er != nil {
		return er
	}

	if wire.LastUpdatedTimestamp != nil {
		if model.LastUpdatedTimestamp, er = parseTimestamp(*wire.LastUpdatedTimestamp); er != nil {
			return er
		}
	}
	model.Pid = wire.Pid
	model.Email = wire.Email
	model.Username = wire.Username
	model.DisplayName = wire.DisplayName
	model.Location = wire.Location
	model.Image = wire.Image
	model.Passhash = wire.Passhash
	model.IsSuspended = wire.IsSuspended
	model.IsAdmin = wire.IsAdmin
	model.TwitterID = wire.TwitterID
	model.GoogleID = wire.GoogleID
	model.FacebookID = wire.FacebookID

	return nil
}

func (model *User) MarshalJSON() ([]byte, error) {
	wire := wireUser{
		LastUpdatedTimestamp: unparseTimestampOptional(model.LastUpdatedTimestamp),
		Pid:                  model.Pid,
		Email:                model.Email,
		Username:             model.Username,
		DisplayName:          model.DisplayName,
		Location:             model.Location,
		Image:                model.Image,
		Passhash:             model.Passhash,
		IsSuspended:          model.IsSuspended,
		IsAdmin:              model.IsAdmin,
		TwitterID:            model.TwitterID,
		GoogleID:             model.GoogleID,
		FacebookID:           model.FacebookID,
	}

	return json.Marshal(wire)
}
