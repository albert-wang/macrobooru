package modify

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"macrobooru/api"
	"macrobooru/models"
)

type ModifyRequest struct {
	json.Unmarshaler

	ModelName   string
	Model       models.ModelMeta
	GUID        models.GUID
	Delete      bool
	FieldValues map[string]interface{}

	sqlFields []string
	sqlArgs   []interface{}
}

func (req *ModifyRequest) UnmarshalJSON(bs []byte) (er error) {
	/* These are surprisingly hard to unmarshal, since the field values are interleaved
	 * with the metadata. They can be distinguished because the metadata keys are prefixed
	 * with a '#' */
	rawValues := map[string]interface{}{}

	if er = json.Unmarshal(bs, &rawValues); er != nil {
		return er
	}

	var ok bool

	/* We need to pull out the metadata first (rather than inline with the rest of the data)
	 * because we need the ModelMeta to figure out which value is the primary key (in case
	 * they don't send `#primary` */

	if modelName, ok := rawValues["#model"]; !ok {
		return api.ErrorMissingInputField("#model")

	} else if req.ModelName, ok = modelName.(string); !ok {
		return api.ErrorInvalidInputField("#model")
	}

	req.Model = models.ModelByName(req.ModelName)
	if req.Model == nil {
		return api.ErrorInvalidInputField("#model")
	}

	var guidIface interface{}

	if guidIface, ok = rawValues["#primary"]; !ok {
		if guidIface, ok = rawValues[req.Model.PrimaryKey()]; !ok {
			return api.ErrorMissingInputField("#primary")
		}
	}

	if guidStr, ok := guidIface.(string); !ok {
		return api.ErrorInvalidInputField("#primary")

	} else if req.GUID, er = models.GUIDFromString(guidStr); er != nil {
		return api.ErrorGeneric(er)
	}

	_, req.Delete = rawValues["#delete"]

	req.FieldValues = make(map[string]interface{})

	for k, v := range rawValues {
		if !strings.HasPrefix(k, "#") {
			req.FieldValues[k] = v
		}
	}

	return nil
}

func (req *ModifyRequest) MarshalJSON() (bs []byte, er error) {
	values := map[string]interface{}{
		"#model":   req.ModelName,
		"#primary": req.GUID.String(),
	}

	if req.Delete {
		values["#delete"] = true

	} else {
		/* XXX: Sending relation fields as part of a delete request causes the server
		 * to erroneously check those (and potentially return an error if they no longer
		 * exist). So only send field data if we're *not* deleting. */
		for k, v := range req.FieldValues {
			values[k] = v
		}
	}

	return json.Marshal(values)
}

func (req *ModifyRequest) verify(db *sql.DB) error {
	guidType := reflect.TypeOf(models.GUID{})
	timeType := reflect.TypeOf(time.Time{})

	for fieldName, fieldValue := range req.FieldValues {
		sqlField, fieldType, ok := models.JsonFieldInfo(req.Model, fieldName)
		if !ok {
			return api.ErrorInvalidInputField(fieldName)
		}

		/* Ugh, we have to do type-specific coercions here to make everything work because 
		 * JSON is stupid and our spec is pretty much undefined past JSON */

		if fieldType == guidType {
			/* Just make sure it's a valid GUID. We serialize to a string anyway. Treat the
			 * empty string and nil as invalid GUIDs (why... what?) */
			if fieldValue == nil {
				fieldValue = models.GUID{}.String()

			} else if guidStr, ok := fieldValue.(string); ok {
				if guidStr == "" {
					fieldValue = models.GUID{}.String()

				} else if _, er := models.GUIDFromString(guidStr); er != nil {
					return api.ErrorGeneric(er)
				}
			}

		} else if fieldType.Kind() == reflect.Bool {
			/* Bools are passed as integer values (?!?) which are actually float64's (because
			 * all numerics in JSON are float64's by-spec, so meh */
			switch val := fieldValue.(type) {
			case float64:
				fieldValue = val == 0
			}

		} else if fieldType == timeType {
			/* JSON doesn't define a date interchange format, so ... let's just ... guess.
			 * Firefox/Chrome use RFC3339 with nanoseconds; no idea what anyone else does
			 * (incl. nativ itself, lol). It *seems* like nativ marshals timestamps to
			 * numerics via UNIX timestamps, so let's do that. */

			if fieldValue == nil {
				/* Just ignore this field; this is not a nullable time */
				continue

			} else {
				switch val := fieldValue.(type) {
				case float64:
					fieldValue = int64(val)

				case string:
					if tmp, er := time.Parse(time.RFC3339Nano, val); er == nil {
						fieldValue = tmp.Unix()

					} else if tmp, er := time.Parse(time.RFC3339, val); er == nil {
						fieldValue = tmp.Unix()

					} else if ival, er := strconv.ParseInt(val, 10, 64); er == nil {
						fieldValue = time.Unix(ival, 0)
					}

				default:
					continue
				}
			}
		}

		/* Bypass the assignable check for GUIDs, because we marshal them to strings (which are,
		 * duh, unassignable to a models.GUID) for SQL reasons */
		if fieldType != guidType && fieldType != timeType {
			if !reflect.TypeOf(fieldValue).AssignableTo(fieldType) {
				return api.ErrorInvalidInputField(fieldName)
			}
		}

		req.sqlArgs = append(req.sqlArgs, fieldValue)
		req.sqlFields = append(req.sqlFields, sqlField)
	}

	return nil
}

func (req *ModifyRequest) run(db *sql.DB) error {
	if len(req.sqlFields) == 0 {
		return nil
	}

	/* Attempt to get the book first, ugh, why the hell don't we use `ON CONFLICT REPLACE`?! */
	q := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = $1", req.Model.TableName(), req.Model.PrimaryKey())
	rows, er := db.Query(q, req.GUID.String())
	if er != nil {
		return api.ErrorGeneric(er)
	}

	if !rows.Next() {
		rows.Close()
		return api.ErrorGeneric(fmt.Errorf("No rows returned for a COUNT(*)"))
	}

	var count int64
	er = rows.Scan(&count)
	rows.Close()

	if er != nil {
		return api.ErrorGeneric(er)
	}

	fmt.Printf("%#v\n%#v\n", req.sqlFields, req.sqlArgs)

	if count != 0 {
		fields := []string{}

		for idx, fieldName := range req.sqlFields {
			fields = append(fields, fmt.Sprintf("%s = $%d", fieldName, idx+1))
		}

		fieldList := strings.Join(fields, ", ")
		req.sqlArgs = append(req.sqlArgs, req.GUID.String())

		q := fmt.Sprintf("UPDATE %s SET %s WHERE %s = $%d", req.Model.TableName(), fieldList, req.Model.PrimaryKey(), len(req.sqlArgs))

		if _, er := db.Exec(q, req.sqlArgs...); er != nil {
			return api.ErrorGeneric(er)
		}

	} else {
		fields := []string{}
		placeholders := []string{}

		for idx, fieldName := range req.sqlFields {
			fields = append(fields, fieldName)
			placeholders = append(placeholders, fmt.Sprintf("$%d", idx+1))
		}

		fieldList := strings.Join(fields, ", ")
		phList := strings.Join(placeholders, ", ")

		q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", req.Model.TableName(), fieldList, phList)

		if _, er := db.Exec(q, req.sqlArgs...); er != nil {
			return api.ErrorGeneric(er)
		}
	}

	return nil
}
