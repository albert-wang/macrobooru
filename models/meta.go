package models

import (
	"reflect"
	"strings"
)

type ModelMeta interface {
	Name() string
	TableName() string
	PrimaryKey() string
	RelationExists(string) bool
	GUID() GUID

	Type() reflect.Type
	SliceType() reflect.Type

	RelationFieldNames(string) (other string, self string, meta ModelMeta)
	BridgeRelationMeta(string) (other string, self string, join ModelMeta, meta ModelMeta)
}

var (
	TagMeta            = &tagMeta{}
	CommentMeta        = &commentMeta{}
	TagBridgeMeta      = &tagBridgeMeta{}
	ImageMeta          = &imageMeta{}
	UploadMetadataMeta = &uploadMetadataMeta{}
	RatingMeta         = &ratingMeta{}
	StaticMeta         = &staticMeta{}
)

var modelNameMap = map[string]ModelMeta{
	"Tag":            TagMeta,
	"Comment":        CommentMeta,
	"TagBridge":      TagBridgeMeta,
	"Image":          ImageMeta,
	"UploadMetadata": UploadMetadataMeta,
	"Rating":         RatingMeta,
	"Static":         StaticMeta,
}

func ModelByName(name string) ModelMeta {
	if model, ok := modelNameMap[name]; ok {
		return model
	}

	return nil
}

func JsonFieldInfo(model ModelMeta, fieldName string) (string, reflect.Type, bool) {
	modelType := model.Type()

	for i := 0; i < modelType.NumField(); i += 1 {
		field := modelType.Field(i)

		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		crudTag := strings.Split(field.Tag.Get("crud"), ",")[0]

		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		if jsonTag != fieldName {
			continue
		}

		if crudTag == "" || crudTag == "-" {
			continue
		}

		return crudTag, field.Type, true
	}

	return "", nil, false
}
