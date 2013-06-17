package query

import (
	"bytes"
	"fmt"
	"strings"

	"macrobooru/models"
)

const MaxLimit = 200

const (
	whereEqual    = 0
	whereNotEqual = iota
	whereGreater
	whereGreaterEqual
	whereLess
	whereLessEqual
	whereNull
	whereNotNull
)

type WhereClause struct {
	Field   string
	Value   interface{}
	Operand int
}

type SqlFragment struct {
	Table string
	Alias string

	Where []WhereClause

	Limit  int64
	Offset int64
	Order  string

	Join   *SqlFragment
	On     map[string]string
	Target models.ModelMeta
}

func (frag *SqlFragment) whereSQL() (string, []interface{}) {
	bits, args := frag.rawWhereSQL("")

	if len(bits) == 0 {
		return "", nil
	}

	return " WHERE " + strings.Join(bits, " AND "), args
}

func (frag *SqlFragment) rawWhereSQL(prefix string) ([]string, []interface{}) {
	bits := []string{}
	args := []interface{}{}

	for _, whereClause := range frag.Where {
		/* Special case to handle `#primary : [pid, pid]`; it falls over to other
		 * fields, but that's an acceptable generalization */
		if slice, ok := whereClause.Value.([]interface{}); ok {
			placeholders := []string{}

			for _, val := range slice {
				placeholders = append(placeholders, "?")
				args = append(args, val)
			}

			placeholder := strings.Join(placeholders, ", ")

			if len(placeholder) > 0 {
				bits = append(bits, fmt.Sprintf("%s%s IN (%s)", prefix, whereClause.Field, placeholder))

			} else {
				bits = append(bits, "0 = 1")
			}

			continue
		}

		switch whereClause.Operand {
		case whereGreater:
			bits = append(bits, fmt.Sprintf("%s%s > ?", prefix, whereClause.Field))
			args = append(args, whereClause.Value)

		case whereGreaterEqual:
			bits = append(bits, fmt.Sprintf("%s%s >= ?", prefix, whereClause.Field))
			args = append(args, whereClause.Value)

		case whereLess:
			bits = append(bits, fmt.Sprintf("%s%s < ?", prefix, whereClause.Field))
			args = append(args, whereClause.Value)

		case whereLessEqual:
			bits = append(bits, fmt.Sprintf("%s%s <= ?", prefix, whereClause.Field))
			args = append(args, whereClause.Value)

		case whereEqual:
			bits = append(bits, fmt.Sprintf("%s%s = ?", prefix, whereClause.Field))
			args = append(args, whereClause.Value)

		case whereNotEqual:
			bits = append(bits, fmt.Sprintf("%s%s != ?", prefix, whereClause.Field))
			args = append(args, whereClause.Value)

		case whereNull:
			bits = append(bits, fmt.Sprintf("%s%s IS NULL", prefix, whereClause.Field))

		case whereNotNull:
			bits = append(bits, fmt.Sprintf("%s%s IS NOT NULL", prefix, whereClause.Field))
		}
	}

	return bits, args
}

func (frag *SqlFragment) toSQL() (string, []interface{}) {
	whereClauses, whereArgs := frag.whereSQL()
	joinClauses, joinArgs := frag.joinSQL()

	if frag.Limit < 1 || frag.Limit > MaxLimit {
		frag.Limit = MaxLimit
	}
	limitClause := fmt.Sprintf(" LIMIT %d", frag.Limit)

	offsetClause := fmt.Sprintf(" OFFSET %d", frag.Offset)

	orderClause := ""
	if frag.Order != "" {
		orderClause = " ORDER BY " + frag.Order
	}

	args := make([]interface{}, len(whereArgs)+len(joinArgs))
	copy(args[0:], whereArgs)
	copy(args[len(whereArgs):], joinArgs)

	sql := fmt.Sprintf(`SELECT %s.* FROM %s AS %s%s%s%s%s%s`, frag.Alias, frag.Table, frag.Alias, joinClauses, whereClauses, orderClause, limitClause, offsetClause)

	return frag.anonToOrderedPlaceholders(sql), args
}

func (frag *SqlFragment) toCountSQL() (string, []interface{}) {
	whereClauses, whereArgs := frag.whereSQL()
	joinClauses, joinArgs := frag.joinSQL()

	args := make([]interface{}, len(whereArgs)+len(joinArgs))
	copy(args[0:], whereArgs)
	copy(args[len(whereArgs):], joinArgs)

	sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s AS %s%s%s`, frag.Table, frag.Alias, joinClauses, whereClauses)
	return frag.anonToOrderedPlaceholders(sql), args
}

func (frag *SqlFragment) joinSQL() (string, []interface{}) {
	if frag.Join == nil {
		return "", nil
	}

	onClauses := []string{}
	for k, v := range frag.On {
		onClauses = append(onClauses, fmt.Sprintf("%s.%s = %s.%s", frag.Alias, k, frag.Join.Alias, v))
	}

	additionalClauses, args := frag.Join.rawWhereSQL(frag.Join.Alias + ".")
	for _, clause := range additionalClauses {
		onClauses = append(onClauses, clause)
	}

	onClause := ""
	if len(onClauses) > 0 {
		onClause = "ON " + strings.Join(onClauses, " AND ")
	}

	join := fmt.Sprintf(" INNER JOIN %s %s %s", frag.Join.Table, frag.Join.Alias, onClause)

	if frag.Join.Join != nil {
		newJoin, _ := frag.Join.joinSQL()
		join += newJoin
	}

	return join, args
}

func (frag *SqlFragment) JoinOn(join *SqlFragment, myField, joinField string) {
	frag.Join = join

	if frag.On == nil {
		frag.On = make(map[string]string)
	}

	frag.On[myField] = joinField
}

func (frag *SqlFragment) anonToOrderedPlaceholders(sql string) string {
	/* Fix argument placeholders to actually work */
	var buf bytes.Buffer
	seenPlaceholders := 1

	for _, b := range []byte(sql) {
		if b == '?' {
			placeholder := fmt.Sprintf("$%d", seenPlaceholders)
			buf.WriteString(placeholder)
			seenPlaceholders += 1

		} else {
			buf.WriteByte(b)
		}
	}

	return buf.String()
}
