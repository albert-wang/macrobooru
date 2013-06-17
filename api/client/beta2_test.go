package client

import (
	"macrobooru/models"
	"testing"
)

const (
	testUsername  = "test"
	testPassword  = "purpleMountainRoad"
	beta2Endpoint = "http://beta2.macrobooru.ironclad.mobi/v2/api"
)

func getAuthenticatedClient(t *testing.T) *Client {
	client, er := NewClient(beta2Endpoint)
	if er != nil {
		t.Fatal(er)
	}

	if er := client.Authenticate(testUsername, testPassword); er != nil {
		t.Fatal(er)
	}

	return client
}

func TestBeta2Authenticate(t *testing.T) {
	getAuthenticatedClient(t)
}

func getTestUser(t *testing.T, client *Client) models.User {
	userOut := []models.User{}

	query := NewQuery()

	userQ := query.Add("User", &userOut)
	userQ.Where(map[string]interface{}{
		"username =": testUsername,
	})

	if er := query.Execute(client); er != nil {
		t.Fatal(er)
	}

	if len(userOut) != 1 {
		t.Fatalf("len(users) != 1, got\n%#v", userOut)
	}

	return userOut[0]
}

func TestBeta2QueryUser(t *testing.T) {
	client := getAuthenticatedClient(t)

	user := getTestUser(t, client)

	if user.Email != "test@macrobooru.com" {
		t.Errorf("user.Email mismatch")
	}

	if user.Username != "test" {
		t.Errorf("user.Username mismatch")
	}
}

func TestBeta2RtBookNoMedia(t *testing.T) {
	client := getAuthenticatedClient(t)

	user := getTestUser(t, client)

	book := models.Book{
		Pid:     models.NewGUID(),
		Title:   "auto-gen-test-book",
		UserRef: user.Pid,
		UrlSlug: "auto-gen-test-book",
	}

	page1 := models.Page{
		Pid:     models.NewGUID(),
		Title:   "page-1-title",
		Ordinal: 0,
		BookRef: book.Pid,
	}

	page2 := models.Page{
		Pid:     models.NewGUID(),
		Title:   "page-2-title",
		Ordinal: 1,
		BookRef: book.Pid,
	}

	partial1_1 := models.Partial{
		Pid:     models.NewGUID(),
		Text:    "partial-1-1-text",
		Ordinal: 0,
		PageRef: page1.Pid,
	}

	partial1_2 := models.Partial{
		Pid:     models.NewGUID(),
		Text:    "partial-1-2-text",
		Ordinal: 1,
		PageRef: page1.Pid,
	}

	partial2_1 := models.Partial{
		Pid:     models.NewGUID(),
		Text:    "partial-2-1-text",
		Ordinal: 0,
		PageRef: page2.Pid,
	}

	partial2_2 := models.Partial{
		Pid:     models.NewGUID(),
		Text:    "partial-2-2-text",
		Ordinal: 1,
		PageRef: page2.Pid,
	}

	mod := NewModification().AddObjects(book, page1, page2, partial1_1, partial1_2, partial2_1, partial2_2)

	if er := mod.Execute(client); er != nil {
		t.Fatal(er)
	}

	/* XXX: Query + check contents */
	query := NewQuery()

	var bookSlice []models.Book
	var pageSlice []models.Page
	var partialSlice []models.Partial

	query.Add("book", &bookSlice).
		Where(map[string]interface{}{
		"pid =": book.Pid,
	})

	query.Subgraph("pages", "book", "pages", &pageSlice)
	query.Subgraph("partials", "pages", "partials", &partialSlice)

	if er := query.Execute(client); er != nil {
		t.Fatal(er)
	}

	if len(bookSlice) != 1 {
		t.Errorf("len(bookSlice) != 1 (is %d)", len(bookSlice))

	} else {
		fetchedBook := bookSlice[0]

		if !book.Pid.Equal(fetchedBook.Pid) {
			t.Errorf("mismatch: book.Pid")
		}

		if book.Title != fetchedBook.Title {
			t.Errorf("mismatch: book.Title")
		}

		if !book.UserRef.Equal(fetchedBook.UserRef) {
			t.Errorf("mismatch: book.UserRef")
		}

		if book.UrlSlug != fetchedBook.UrlSlug {
			t.Errorf("mismatch: book.UrlSlug")
		}
	}

	if len(pageSlice) != 2 {
		t.Errorf("len(pageSlice) != 2 (is %d)", len(pageSlice))

	} else {
		for i := range pageSlice {
			fetchedPage := pageSlice[i]
			var oldPage models.Page

			if fetchedPage.Pid.Equal(page1.Pid) {
				oldPage = page1

			} else if fetchedPage.Pid.Equal(page2.Pid) {
				oldPage = page2

			} else {
				t.Errorf("unknown Pid in returned page slice")
				break
			}

			if oldPage.Title != fetchedPage.Title {
				t.Errorf("mismatch: page.Title")
			}

			if oldPage.Ordinal != fetchedPage.Ordinal {
				t.Errorf("mismatch: page.Ordinal")
			}

			if !oldPage.BookRef.Equal(fetchedPage.BookRef) {
				t.Errorf("mismatch: page.BookRef")
			}
		}
	}

	if len(partialSlice) != 4 {
		t.Errorf("len(partialSlice) != 4 (is %d)", len(partialSlice))

	} else {
		for i := range partialSlice {
			fetchedPartial := partialSlice[i]
			var oldPartial models.Partial

			if fetchedPartial.Pid.Equal(partial1_1.Pid) {
				oldPartial = partial1_1

			} else if fetchedPartial.Pid.Equal(partial1_2.Pid) {
				oldPartial = partial1_2

			} else if fetchedPartial.Pid.Equal(partial2_1.Pid) {
				oldPartial = partial2_1

			} else if fetchedPartial.Pid.Equal(partial2_2.Pid) {
				oldPartial = partial2_2

			} else {
				t.Errorf("unknown partial Pid returned")
				break
			}

			if oldPartial.Text != fetchedPartial.Text {
				t.Errorf("mismatch: partial.Text")
			}

			if oldPartial.Ordinal != fetchedPartial.Ordinal {
				t.Errorf("mismatch: partial.Ordinal")
			}

			if !oldPartial.PageRef.Equal(fetchedPartial.PageRef) {
				t.Errorf("mismatch: partial.PageRef")
			}
		}
	}

	mod = NewModification().DeleteObjects(book, page1, page2, partial1_1, partial1_2, partial2_1, partial2_2)

	if er := mod.Execute(client); er != nil {
		t.Fatal(er)
	}
}
