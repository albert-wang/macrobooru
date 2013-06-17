package client

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"reflect"

	"macrobooru/api/operations/modify"
	"macrobooru/models"
)

type Modification struct {
	deferredError error
	guidMap       map[string]bool
	payload       modify.ModifyPayload
}

func NewModification() *Modification {
	return &Modification{
		guidMap: make(map[string]bool),
	}
}

func (mod *Modification) modifyObject(object interface{}, shouldDelete bool) error {
	val := reflect.ValueOf(object)

	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	modelMeta := models.ModelByName(val.Type().Name())
	if modelMeta == nil {
		return fmt.Errorf("Non-model type '%s' in modification", val.Type().Name())
	}

	fieldValues := map[string]interface{}{}

	var guid models.GUID

	tmpBytes, er := json.Marshal(object)
	if er != nil {
		return er
	}

	if er := json.Unmarshal(tmpBytes, &fieldValues); er != nil {
		return er
	}

	if guidIface, ok := fieldValues[modelMeta.PrimaryKey()]; ok {
		if guidStr, ok := guidIface.(string); ok {
			if guid, er = models.GUIDFromString(guidStr); er != nil {
				return er
			}
		} else {
			return fmt.Errorf("Model %s pkey field not serialized to string (e.g. may not be a GUID?)", val.Type().Name())
		}

	} else {
		return fmt.Errorf("Model %s does not appear to have a primary key", val.Type().Name())
	}

	if !guid.IsValid() {
		/* XXX: Do this automatically? We need some way to return this */
		return fmt.Errorf("Model %s does not have a valid GUID set. This is not yet done automatically.", val.Type().Name())
	}

	if _, ok := mod.guidMap[guid.String()]; ok {
		/* XXX: Maybe overwrite the old mod? */
		return fmt.Errorf("Cannot send modification/deletion of same object (%s) twice in one request", guid.String())
	}

	mod.guidMap[guid.String()] = true

	req := modify.ModifyRequest{
		ModelName:   modelMeta.Name(),
		Model:       modelMeta,
		GUID:        guid,
		FieldValues: fieldValues,
		Delete:      shouldDelete,
	}

	mod.payload = append(mod.payload, req)

	return nil
}

func (mod *Modification) AddObjects(objects ...interface{}) *Modification {
	for i := 0; i < len(objects) && mod.deferredError == nil; i += 1 {
		mod.deferredError = mod.modifyObject(objects[i], false)
	}

	return mod
}

func (mod *Modification) DeleteObjects(objects ...interface{}) *Modification {
	for i := 0; i < len(objects) && mod.deferredError == nil; i += 1 {
		mod.deferredError = mod.modifyObject(objects[i], true)
	}

	return mod
}

func (mod *Modification) Execute(client *Client) error {
	if mod.deferredError != nil {
		return mod.deferredError
	}

	resWrapper, er := client.Execute(&mod.payload, nil)
	if er != nil {
		return er
	}

	rawResponse, er := mod.payload.ParseResponse(*resWrapper)
	if er != nil {
		return er
	}

	/* This returns specialized error values to tell us exactly which
	 * things are causing problems, if any. Not sure it would get to this
	 * point, since the client would presumably return an error first */
	_ = rawResponse.(modify.ModifyResponse)

	return nil
}

type UploadModification struct {
	static models.Static
	data   multipart.File
}

func NewUploadModification(static models.Static, data multipart.File) *UploadModification {
	return &UploadModification{
		static: static,
		data:   data,
	}
}

func (mod *UploadModification) buildPayload() (*modify.ModifyPayload, error) {
	modelMeta := models.StaticMeta

	if !mod.static.Pid.IsValid() {
		/* XXX: Generate one? We need some way of returning the value */
		return nil, fmt.Errorf("Must supply a valid GUID for the static object")
	}

	if mod.static.SHA1Hash == "" {
		hash := sha1.New()
		if _, er := io.Copy(hash, mod.data); er != nil {
			return nil, er
		}

		if _, er := mod.data.Seek(0, os.SEEK_SET); er != nil {
			return nil, er
		}

		mod.static.SHA1Hash = fmt.Sprintf("%x", hash.Sum(nil))
	}

	payload := modify.ModifyPayload{
		modify.ModifyRequest{
			ModelName: modelMeta.Name(),
			Model:     modelMeta,
			GUID:      mod.static.Pid,
			FieldValues: map[string]interface{}{
				"pid":      mod.static.Pid,
				"mime":     mod.static.Mime,
				"SHA1Hash": mod.static.SHA1Hash,
				"path":     "attached-file",
			},
		},
	}

	return &payload, nil
}

func (mod *UploadModification) Execute(client *Client) error {
	payload, er := mod.buildPayload()
	if er != nil {
		return er
	}

	resWrapper, er := client.Execute(payload, map[string]multipart.File{
		"attached-file": mod.data,
	})

	if er != nil {
		return er
	}

	rawResponse, er := payload.ParseResponse(*resWrapper)
	if er != nil {
		return er
	}

	/* Again, we currently don't unmarshal any of the meaningful data from
	 * error responses, and there's no meaningful data in successful responses.
	 * Just going through the motions here. */
	_ = rawResponse.(modify.ModifyResponse)

	return nil
}
