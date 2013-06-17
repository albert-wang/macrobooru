package query

import (
	"encoding/json"
	"fmt"
	"testing"
)

func testRequestSql(t *testing.T, jsonBytes []byte, subgraphName string, expectedSql string) []interface{} {
	var payload QueryPayload
	if er := json.Unmarshal(jsonBytes, &payload); er != nil {
		t.Fatal(er)
	}

	request, ok := payload.GetNamedRequest(subgraphName)
	if !ok {
		t.Fatalf("unable to fetch named request")
	}

	sqlFrag, er := request.Decompose(payload)
	if er != nil {
		t.Fatal(er)
	}

	sql, args := sqlFrag.toSQL()
	if sql != expectedSql {
		t.Logf("%#v\n", sqlFrag)
		t.Fatalf("Incorrect SQL, got\n%s\nexpected\n%s", sql, expectedSql)
	}

	return args
}

func TestRequestSimple(t *testing.T) {
	jsonBytes := []byte(`
		{ "module" :
			{ "model" : "Book"
			, "where" :
				{ "pid" : "asd"
				}
			}
		}
	`)

	expected := fmt.Sprintf("SELECT alias1.* FROM book AS alias1 WHERE pid = $1 LIMIT %d OFFSET 0", MaxLimit)

	testRequestSql(t, jsonBytes, "module", expected)
}

func TestRequestRelation(t *testing.T) {
	jsonBytes := []byte(`
		{ "primary" :
			{ "model" : "Book"
			, "where" :
				{ "pid" : "asd"
				}
			, "relation" : "pages"
			}
		}
	`)

	expected := fmt.Sprintf("SELECT alias2.* FROM page AS alias2 INNER JOIN book alias1 ON alias2.bookRef = alias1.pid AND alias1.pid = $1 LIMIT %d OFFSET 0", MaxLimit)

	testRequestSql(t, jsonBytes, "primary", expected)
}

func TestRequestSubgraph(t *testing.T) {
	jsonBytes := []byte(`
		{ "primary" :
			{ "model" : "Book"
			, "where" :
				{ "pid" : "asd"
				}
			}
		, "pages" :
			{ "subgraph" : "primary"
			, "relation" : "pages"
			}
		}
	`)

	expected := fmt.Sprintf("SELECT alias2.* FROM page AS alias2 INNER JOIN book alias1 ON alias2.bookRef = alias1.pid AND alias1.pid = $1 LIMIT %d OFFSET 0", MaxLimit)

	testRequestSql(t, jsonBytes, "pages", expected)
}

func TestRequestSubgraph2(t *testing.T) {
	jsonBytes := []byte(`
		{ "one" :
			{ "model" : "Tag"
			, "where" : 
				{ "name" : "cat"
				, "#primary" : "asd"
				}
			}
		, "two" :
			{ "subgraph" : "one"
			, "relation" : "BookTags"
			}
		}
	`)

	expected1 := fmt.Sprintf("SELECT alias2.* FROM book AS alias2 INNER JOIN tagbridge alias3 ON alias2.pid = alias3.book_id INNER JOIN tag alias1 ON alias3.tag_id = alias1.pid AND alias1.name = $1 AND alias1.pid = $2 LIMIT %d OFFSET 0", MaxLimit)
	expected2 := fmt.Sprintf("SELECT alias2.* FROM book AS alias2 INNER JOIN tagbridge alias3 ON alias2.pid = alias3.book_id INNER JOIN tag alias1 ON alias3.tag_id = alias1.pid AND alias1.pid = $1 AND alias1.name = $2 LIMIT %d OFFSET 0", MaxLimit)

	var payload QueryPayload
	if er := json.Unmarshal(jsonBytes, &payload); er != nil {
		t.Fatal(er)
	}

	request, ok := payload.GetNamedRequest("two")
	if !ok {
		t.Fatalf("unable to fetch named request")
	}

	sqlFrag, er := request.Decompose(payload)
	if er != nil {
		t.Fatal(er)
	}

	sql, _ := sqlFrag.toSQL()
	if sql != expected1 && sql != expected2 {
		t.Logf("%#v\n", sqlFrag)
		t.Fatalf("Incorrect SQL, got\n%s\nexpected\n%s", sql, expected1)
	}
}

func TestOld1(t *testing.T) {
	jsonBytes := []byte(`
		{
			"book" : {
				"model" : "Book"
			}
		}
	`)

	expected := fmt.Sprintf("SELECT alias1.* FROM book AS alias1 LIMIT %d OFFSET 0", MaxLimit)

	testRequestSql(t, jsonBytes, "book", expected)
}

func TestOld2(t *testing.T) {
	jsonBytes := []byte(`
		{
			"__unnamed__" : {
				"model" : "Book"
			}, 
			"Ratings" : { 
				"subgraph" : "__unnamed__", 
				"relation" : "ratings"
			}
		}
	`)

	expected := fmt.Sprintf("SELECT alias2.* FROM rating AS alias2 INNER JOIN book alias1 ON alias2.book_id = alias1.pid LIMIT %d OFFSET 0", MaxLimit)

	testRequestSql(t, jsonBytes, "Ratings", expected)
}
