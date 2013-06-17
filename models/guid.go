package models

import (
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type GUID struct {
	bytes []byte
}

var _ sql.Scanner = &GUID{}
var _ json.Marshaler = &GUID{}
var _ json.Unmarshaler = &GUID{}

const (
	invalidGUIDString = "00000000FFFFFFFF00000000FFFFFFFF"
	invalidGUIDPart   = 0xffffffff
)

func NewGUID() (g GUID) {
	g.bytes = make([]byte, 16)

	if _, er := rand.Read(g.bytes); er != nil {
		/* XXX: ... */
		panic(er)
	}

	return
}

func newModelGUID(modelId int) GUID {
	var result GUID

	result.bytes = make([]byte, 16)
	if _, er := rand.Read(result.bytes); er != nil {
		panic(er)
	}

	t := time.Now().Unix()
	result.bytes[0] = byte(modelId & 0xFF)
	binary.BigEndian.PutUint32(result.bytes[12:16], uint32(t&0xFFFFFFFF))
	return result
}

func GUIDFromString(s string) (g GUID, er error) {
	if s == invalidGUIDString {
		return
	}

	g.bytes, er = hex.DecodeString(s)
	return
}

func GUIDFromPair(high, low uint64) (g GUID) {
	if high == invalidGUIDPart && low == invalidGUIDPart {
		return
	}

	g.bytes = make([]byte, 16)

	binary.BigEndian.PutUint64(g.bytes[0:8], high)
	binary.BigEndian.PutUint64(g.bytes[8:16], low)

	return
}

func (g GUID) String() string {
	if g.bytes == nil {
		return invalidGUIDString
	}

	return strings.ToUpper(hex.EncodeToString(g.bytes))
}

func (g GUID) Low() uint64 {
	if g.bytes == nil {
		return invalidGUIDPart
	}

	return binary.BigEndian.Uint64(g.bytes[8:16])
}

func (g GUID) High() uint64 {
	if g.bytes == nil {
		return invalidGUIDPart
	}

	return binary.BigEndian.Uint64(g.bytes[0:8])
}

func (g GUID) Equal(g1 GUID) bool {
	return g.High() == g1.High() && g.Low() == g1.Low()
}

func (g GUID) IsValid() bool {
	return g.bytes != nil
}

func (g GUID) IsReserved() bool {
	return !g.IsValid()
}

func (g *GUID) Scan(src interface{}) error {
	switch val := src.(type) {
	case []byte:
		tmp, er := GUIDFromString(string(val))
		if er != nil {
			return er
		}

		*g = tmp

	case string:
		tmp, er := GUIDFromString(val)
		if er != nil {
			return er
		}

		*g = tmp

	default:
		return fmt.Errorf("Unable to coerce %T %#v to GUID", src, src)
	}

	return nil
}

func (g GUID) MarshalJSON() ([]byte, error) {
	return []byte("\"" + g.String() + "\""), nil
}

func (g *GUID) UnmarshalJSON(bs []byte) error {
	var strVal string

	if er := json.Unmarshal(bs, &strVal); er != nil {
		return er
	}

	if tmp, er := GUIDFromString(strVal); er != nil {
		return er

	} else {
		*g = tmp
	}

	return nil
}
