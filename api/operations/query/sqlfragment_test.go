package query

import (
	"fmt"
	"testing"
)

func makeVerifier(t *testing.T, sql, expected *string, args *[]interface{}) func(int, string) {
	return func(expectedLen int, which string) {
		if *sql != *expected {
			t.Fatalf("%s: Incorrect SQL, got\n%s\nexpected\n%s", which, *sql, *expected)
		}

		if len(*args) != expectedLen {
			t.Fatalf("%s: Incorrect args, got %#v\n%s\n", which, *args, *sql)
		}
	}
}

func TestSimple(t *testing.T) {
	var sql, expected string
	var args []interface{}

	verify := makeVerifier(t, &sql, &expected, &args)

	frag := SqlFragment{
		Table: "table",
		Alias: "alias",
	}

	sql, args = frag.toSQL()
	expected = fmt.Sprintf("SELECT alias.* FROM table AS alias LIMIT %d OFFSET 0", MaxLimit)
	verify(0, "vanilla")

	frag.Limit = 10
	sql, args = frag.toSQL()
	expected = "SELECT alias.* FROM table AS alias LIMIT 10 OFFSET 0"
	verify(0, "limit")

	frag.Offset = 10
	sql, args = frag.toSQL()
	expected = "SELECT alias.* FROM table AS alias LIMIT 10 OFFSET 10"
	verify(0, "offset")

	frag.Where = append(frag.Where, WhereClause{
		Field: "foo",
		Value: "value",
	})
	sql, args = frag.toSQL()
	expected = "SELECT alias.* FROM table AS alias WHERE foo = $1 LIMIT 10 OFFSET 10"
	verify(1, "single where")

	if arg, ok := args[0].(string); !ok {
		t.Fatalf("arg not transferred properly -- is not a string")
	} else if arg != "value" {
		t.Fatalf("arg value not transferred properly -- is %#v", arg)
	}
}

func TestOrder(t *testing.T) {
	var sql, expected string
	var args []interface{}

	verify := makeVerifier(t, &sql, &expected, &args)

	frag := SqlFragment{
		Table:  "table",
		Alias:  "alias",
		Order:  "bar",
		Limit:  10,
		Offset: 12,
	}

	sql, args = frag.toSQL()
	expected = "SELECT alias.* FROM table AS alias ORDER BY bar LIMIT 10 OFFSET 12"
	verify(0, "vanilla")
}

func TestMultipleWhere(t *testing.T) {
	frag := SqlFragment{
		Table:  "table",
		Alias:  "alias",
		Limit:  10,
		Offset: 12,
		Where: []WhereClause{
			WhereClause{
				Field: "foo",
				Value: "value",
			},
			WhereClause{
				Field: "bar",
				Value: "baz",
			},
		},
	}

	sql, args := frag.toSQL()
	expected1 := "SELECT alias.* FROM table AS alias WHERE foo = $1 AND bar = $2 LIMIT 10 OFFSET 12"
	expected2 := "SELECT alias.* FROM table AS alias WHERE bar = $1 AND foo = $2 LIMIT 10 OFFSET 12"

	if len(args) != 2 {
		t.Fatalf("where2: Incorrect arg length, got %#v", args)
	}

	if sql == expected1 {
		if arg, ok := args[0].(string); !ok {
			t.Fatalf("arg not transferred properly -- is not a string")
		} else if arg != "value" {
			t.Fatalf("arg value not transferred properly -- is %#v", arg)
		}

		if arg, ok := args[1].(string); !ok {
			t.Fatalf("arg not transferred properly -- is not a string")
		} else if arg != "baz" {
			t.Fatalf("arg value not transferred properly -- is %#v", arg)
		}

	} else if sql == expected2 {
		if arg, ok := args[1].(string); !ok {
			t.Fatalf("arg not transferred properly -- is not a string")
		} else if arg != "value" {
			t.Fatalf("arg value not transferred properly -- is %#v", arg)
		}

		if arg, ok := args[0].(string); !ok {
			t.Fatalf("arg not transferred properly -- is not a string")
		} else if arg != "baz" {
			t.Fatalf("arg value not transferred properly -- is %#v", arg)
		}

	} else {
		t.Fatalf("where2: Incorrect SQL, got `%s`", sql)
	}
}

func TestCannedJoin(t *testing.T) {
	var sql, expected string
	var args []interface{}

	verify := makeVerifier(t, &sql, &expected, &args)

	frag := SqlFragment{
		Table: "target",
		Alias: "alias1",
		Join: &SqlFragment{
			Table: "subquery",
			Alias: "alias2",
		},
		On: map[string]string{
			"targetId": "subqueryId",
		},
	}

	sql, args = frag.toSQL()
	expected = fmt.Sprintf("SELECT alias1.* FROM target AS alias1 INNER JOIN subquery alias2 ON alias1.targetId = alias2.subqueryId LIMIT %d OFFSET 0", MaxLimit)
	verify(0, "vanilla")

	frag.Join.Limit = 10
	sql, args = frag.toSQL()
	expected = fmt.Sprintf("SELECT alias1.* FROM target AS alias1 INNER JOIN subquery alias2 ON alias1.targetId = alias2.subqueryId LIMIT %d OFFSET 0", MaxLimit)
	verify(0, "should ignore join limit")

	frag.Join.Offset = 10
	sql, args = frag.toSQL()
	expected = fmt.Sprintf("SELECT alias1.* FROM target AS alias1 INNER JOIN subquery alias2 ON alias1.targetId = alias2.subqueryId LIMIT %d OFFSET 0", MaxLimit)
	verify(0, "should ignore join offset")

	frag.Limit = 10
	sql, args = frag.toSQL()
	expected = "SELECT alias1.* FROM target AS alias1 INNER JOIN subquery alias2 ON alias1.targetId = alias2.subqueryId LIMIT 10 OFFSET 0"
	verify(0, "should use base limit")

	frag.Offset = 12
	sql, args = frag.toSQL()
	expected = "SELECT alias1.* FROM target AS alias1 INNER JOIN subquery alias2 ON alias1.targetId = alias2.subqueryId LIMIT 10 OFFSET 12"
	verify(0, "should use base offset")

	frag.Join.Join = &SqlFragment{
		Table: "moresub",
		Alias: "alias3",
	}

	frag.Join.On = map[string]string{
		"subqueryPants": "subsubId",
	}

	sql, args = frag.toSQL()
	expected = "SELECT alias1.* FROM target AS alias1 INNER JOIN subquery alias2 ON alias1.targetId = alias2.subqueryId INNER JOIN moresub alias3 ON alias2.subqueryPants = alias3.subsubId LIMIT 10 OFFSET 12"
	verify(0, "should use base offset")
}

func TestJoin(t *testing.T) {
	var sql, expected string
	var args []interface{}

	verify := makeVerifier(t, &sql, &expected, &args)

	frag1 := &SqlFragment{
		Table:  "target",
		Alias:  "alias1",
		Limit:  10,
		Offset: 12,
	}

	frag2 := &SqlFragment{
		Table: "subquery",
		Alias: "alias2",
	}

	frag1.JoinOn(frag2, "target_id", "subquery_id")

	sql, args = frag1.toSQL()
	expected = "SELECT alias1.* FROM target AS alias1 INNER JOIN subquery alias2 ON alias1.target_id = alias2.subquery_id LIMIT 10 OFFSET 12"
	verify(0, "vanilla")
}

func TestDoubleJoin(t *testing.T) {
	var sql, expected string
	var args []interface{}

	verify := makeVerifier(t, &sql, &expected, &args)

	frag1 := &SqlFragment{
		Table:  "target",
		Alias:  "alias1",
		Limit:  10,
		Offset: 12,
	}

	frag2 := &SqlFragment{
		Table: "subquery1",
		Alias: "alias2",
	}

	frag3 := &SqlFragment{
		Table: "subquery2",
		Alias: "alias3",
	}

	frag2.JoinOn(frag3, "frag2_id", "frag3_id")
	frag1.JoinOn(frag2, "frag1_id", "frag2_id")

	sql, args = frag1.toSQL()
	expected = "SELECT alias1.* FROM target AS alias1 INNER JOIN subquery1 alias2 ON alias1.frag1_id = alias2.frag2_id INNER JOIN subquery2 alias3 ON alias2.frag2_id = alias3.frag3_id LIMIT 10 OFFSET 12"
	verify(0, "vanilla")

	frag1.Where = []WhereClause{
		WhereClause{
			Field: "field",
			Value: "value",
		},
	}

	sql, args = frag1.toSQL()
	expected = "SELECT alias1.* FROM target AS alias1 INNER JOIN subquery1 alias2 ON alias1.frag1_id = alias2.frag2_id INNER JOIN subquery2 alias3 ON alias2.frag2_id = alias3.frag3_id WHERE field = $1 LIMIT 10 OFFSET 12"
	verify(1, "where")
}
