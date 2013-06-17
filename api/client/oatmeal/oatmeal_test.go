package oatmeal

import (
	"io"
	"os"
	"fmt"
	"crypto/sha1"
	"encoding/hex"

	. "macrobooru/api/client"
	. "macrobooru/models"
	"testing"
)

const (
	testUsername  = "test"
	testPassword  = "purpleMountainRoad"
	beta2Endpoint = "http://beta2.macrobooru.ironclad.mobi/v2/api"
)

func uploadStatics(t *testing.T, client *Client) map[string]GUID {
	mapping := make(map[string]GUID)

	for i := 0; i <= 8; i += 1 {
		file := fmt.Sprintf("oatmeal-%d.jpg", i)
		if i == 0 {
			file = "oatmeal-cover.jpg"
		}

		filePath := "img/" + file

		fh, er := os.Open(filePath)
		if er != nil {
			t.Fatal(er)
		}

		hash := sha1.New()

		if _, er := io.Copy(hash, fh); er != nil {
			t.Fatal(er)
		}

		if _, er := fh.Seek(0, os.SEEK_SET); er != nil {
			t.Fatal(er)
		}

		static := Static{
			Pid: NewGUID(),
			Mime: "image/jpeg",
			SHA1Hash: hex.EncodeToString(hash.Sum(nil)),
		}

		fmt.Printf("%#v\n", static)

		if er := NewUploadModification(static, fh).Execute(client) ; er != nil {
			t.Fatal(er)
		}

		mapping[file] = static.Pid
	}

	return mapping
}

func TestBuildOatmeal(t *testing.T) {
	client, er := NewClient(beta2Endpoint)
	if er != nil {
		t.Fatal(er)
	}

	if er := client.Authenticate(testUsername, testPassword); er != nil {
		t.Fatal(er)
	}

	userOut := []User{}
	query := NewQuery()

	query.Add("User", &userOut).Where(map[string]interface{}{"username =": testUsername})

	if er := query.Execute(client); er != nil {
		t.Fatal(er)
	}

	if len(userOut) != 1 {
		t.Fatalf("Cannot look up PID for user %s", testUsername)
	}

	user := userOut[0]
	staticMap := uploadStatics(t, client)
	objects := []interface{}{}

	book := Book{
		Pid: NewGUID(),
		Title: "Making Blueberry Oatmeal Bake",
		UserRef: user.Pid,
		UrlSlug: "making-blueberry-oatmeal-bake",
	}
	objects = append(objects, book)

	page_1 := Page{
		Pid: NewGUID(),
		Title: "Introduction",
		Ordinal: 1,
		BookRef: book.Pid,
	}
	objects = append(objects, page_1)

	partial_1_1 := Partial{
		Pid: NewGUID(),
		Media: staticMap["oatmeal-cover.jpg"],
		PageRef: page_1.Pid,
		BookRef: book.Pid,
		Ordinal: 1,
	}
	objects = append(objects, partial_1_1)

	partial_1_2 := Partial{
		Pid: NewGUID(),
		PageRef: page_1.Pid,
		BookRef: book.Pid,
		Ordinal: 2,
		Text: `
**Yield:** Serves 6 generously, or 12 as part of a larger brunch spread

Oatmeal, also known as white oats is ground oat groats (i.e. oat-meal, cf. cornmeal, peasemeal, etc.), or a porridge made from oats (also called oatmeal cereal or stirabout, in Ireland). Oatmeal can also be ground oat, steel-cut oats, crushed oats, rolled oats, or porridge.

There has been increasing interest in oatmeal in recent years because of its health benefits. Daily consumption of a bowl of oatmeal can lower blood cholesterol, because of its soluble fibre content. After it was reported that oats can help lower cholesterol, an "oat bran craze" swept the U.S. in the late 1980s, peaking in 1989. The food craze was short-lived and faded by the early 1990s. The popularity of oatmeal and other oat products increased again after the January 1997 decision by the Food and Drug Administration that food with a lot of oat bran or rolled oats can carry a label claiming it may reduce the risk of heart disease when combined with a low-fat diet. This is because of the beta-glucan in the oats. Rolled oats have long been a staple of many athletes' diets, especially weight trainers, because of its high content of complex carbohydrates and water-soluble fibre that encourages slow digestion and stabilizes blood-glucose levels. Oatmeal porridge also contains more B vitamins and calories than other kinds of porridges.

*Although I love this huckleberry version, feel free to substitute your favorite in-season berries, or any other fruit for that matter. Another version I love is made with plump, amaretto-soaked golden raisins in place of the berries and sliced almonds in place of the walnuts.*
		`,
	}
	objects = append(objects, partial_1_2)

	page_2 := Page{
		Pid: NewGUID(),
		Title: "Making the mash",
		Ordinal: 2,
		BookRef: book.Pid,
	}
	objects = append(objects, page_2)

	partial_2_1 := Partial{
		Pid: NewGUID(),
		PageRef: page_2.Pid,
		BookRef: book.Pid,
		Ordinal: 1,
		Media: staticMap["oatmeal-1.jpg"],
	}
	objects = append(objects, partial_2_1)

	partial_2_2 := Partial{
		Pid: NewGUID(),
		PageRef: page_2.Pid,
		BookRef: book.Pid,
		Ordinal: 2,
		Text: `
== Step 1

Preheat your oven to 375 degrees. Mash the bananas with a fork, starting with just a couple and adding more until you have 1.5 cups of banana mash. Try not to leave any large chunks. (last time it took me 4 medium bananas, this time only 3).
		`,
	}
	objects = append(objects, partial_2_2)

	partial_2_3 := Partial{
		Pid: NewGUID(),
		PageRef: page_2.Pid,
		BookRef: book.Pid,
		Ordinal: 3,
		Media: staticMap["oatmeal-2.jpg"],
	}
	objects = append(objects, partial_2_3)

	partial_2_4 := Partial{
		Pid: NewGUID(),
		PageRef: page_2.Pid,
		BookRef: book.Pid,
		Ordinal: 4,
		Text: `
== Step 2

Bananas are a staple starch for many tropical populations. Depending upon cultivar and ripeness, the flesh can vary in taste from starchy to sweet, and texture from firm to mushy. Both the skin and inner part can be eaten raw or cooked. The banana's flavor is due, amongst other chemicals, to isoamyl acetate which is one of the main constituents of banana oil.

Combine the banana mash in the large bowl with the eggs, sugar, vanilla, salt, and baking powder. Whisk to combine.
		`,
	}
	objects = append(objects, partial_2_4)

	partial_2_5 := Partial{
		Pid: NewGUID(),
		PageRef: page_2.Pid,
		BookRef: book.Pid,
		Ordinal: 5,
		Media: staticMap["oatmeal-3.jpg"],
	}
	objects = append(objects, partial_2_5)

	partial_2_6 := Partial{
		Pid: NewGUID(),
		PageRef: page_2.Pid,
		BookRef: book.Pid,
		Ordinal: 6,
		Text: `
== Step 3

Next, add the milk and whisk again. It's a lot easier to make a smooth, cohesive mixture if you whisk the other ingredients prior to adding the milk. That's why it's done in two steps.
		`,
	}
	objects = append(objects, partial_2_6)

	page_3 := Page{
		Pid: NewGUID(),
		Title: "Preparing the dish",
		Ordinal: 3,
		BookRef: book.Pid,
	}
	objects = append(objects, page_3)

	partial_3_1 := Partial{
		Pid: NewGUID(),
		PageRef: page_3.Pid,
		BookRef: book.Pid,
		Ordinal: 1,
		Media: staticMap["oatmeal-4.jpg"],
	}
	objects = append(objects, partial_3_1)

	partial_3_2 := Partial{
		Pid: NewGUID(),
		PageRef: page_3.Pid,
		BookRef: book.Pid,
		Ordinal: 2,
		Text: `
== Step 4

Stir in the dry oats. It's easier to stir them in with a large spoon than a whisk, but I didn't feel like dirtying another utensil.
		`,
	}
	objects = append(objects, partial_3_2)

	partial_3_3 := Partial{
		Pid: NewGUID(),
		PageRef: page_3.Pid,
		BookRef: book.Pid,
		Ordinal: 3,
		Media: staticMap["oatmeal-5.jpg"],
	}
	objects = append(objects, partial_3_3)

	partial_3_4 := Partial{
		Pid: NewGUID(),
		PageRef: page_3.Pid,
		BookRef: book.Pid,
		Ordinal: 4,
		Text: `
== Step 5

I love having frozen blueberries on hand (keeping them frozen and stirring them in last helps prevent the entire mix from turning purple). They're a great addition to many dishes, are less expensive than fresh, and keep for quite a while in the freezer. I used half a bag for this recipe (and used the rest for my smoothie packs).

Blueberries are perennial flowering plants with indigo-colored berries in the section Cyanococcus within the genus Vaccinium (a genus that also includes cranberries and bilberries). Species in the section Cyanococcus are the most common fruits sold as "blueberries" and are native to North America (commercially cultivated highbush blueberries were not introduced into Europe until the 1930s).

They are usually erect, but sometimes prostrate shrubs varying in size from 10 centimeters (3.9 in) to 4 meters (13 ft) tall. In commercial blueberry production, smaller species are known as "lowbush blueberries" (synonymous with "wild"), and the larger species are known as "highbush blueberries".

The leaves can be either deciduous or evergreen, ovate to lanceolate, and 1-8 cm (0.39-3.1 in) long and 0.5-3.5 cm (0.20-1.4 in) broad. The flowers are bell-shaped, white, pale pink or red, sometimes tinged greenish. The fruit is a berry 5-16 millimeters (0.20-0.63 in) in diameter with a flared crown at the end; they are pale greenish at first, then reddish-purple, and finally dark blue when ripe. They have a sweet taste when mature, with variable acidity. Blueberry bushes typically bear fruit in the middle of the growing season: fruiting times are affected by local conditions such as altitude and latitude, so the height of the crop can vary from May to August depending upon these conditions.
		`,
	}
	objects = append(objects, partial_3_4)

	partial_3_5 := Partial{
		Pid: NewGUID(),
		PageRef: page_3.Pid,
		BookRef: book.Pid,
		Ordinal: 5,
		Media: staticMap["oatmeal-6.jpg"],
	}
	objects = append(objects, partial_3_5)

	partial_3_6 := Partial{
		Pid: NewGUID(),
		PageRef: page_3.Pid,
		BookRef: book.Pid,
		Ordinal: 6,
		Text: `
== Step 6

Keep the blueberries frozen until you add them so they don't bleed purple color throughout the oatmeal. Adding them at the very end helps this as well. Pour the blueberry oat mixture into an 8x8 inch casserole dish that has been sprayed with non-stick spray.
		`,
	}
	objects = append(objects, partial_3_6)

	page_4 := Page{
		Pid: NewGUID(),
		Title: "Baking",
		Ordinal: 4,
		BookRef: book.Pid,
	}
	objects = append(objects, page_4)

	partial_4_1 := Partial{
		Pid: NewGUID(),
		PageRef: page_4.Pid,
		BookRef: book.Pid,
		Ordinal: 1,
		Media: staticMap["oatmeal-7.jpg"],
	}
	objects = append(objects, partial_4_1)

	partial_4_2 := Partial{
		Pid: NewGUID(),
		PageRef: page_4.Pid,
		BookRef: book.Pid,
		Ordinal: 2,
		Text: `
== Step 7

Bake for about 45 minutes at 375 degrees or until the top is golden brown and no longer wet to the touch in the center. I baked this oatmeal at 25 degrees higher than my other baked oatmeals to compensate for the little blueberry ice cubes floating around in it. It worked out well.
		`,
	}
	objects = append(objects, partial_4_2)

	partial_4_3 := Partial{
		Pid: NewGUID(),
		PageRef: page_4.Pid,
		BookRef: book.Pid,
		Ordinal: 3,
		Media: staticMap["oatmeal-8.jpg"],
	}
	objects = append(objects, partial_4_3)

	partial_4_4 := Partial{
		Pid: NewGUID(),
		PageRef: page_4.Pid,
		BookRef: book.Pid,
		Ordinal: 4,
		Text: `
== Step 8

Serve warm or refrigerate until ready to eat. These oats can be quickly reheated in the microwave each morning for a quick, filling breakfast. I like to eat mine with a little bit of cold milk poured over top. This makes for a very fast, very hearty breakfast every morning of the week!
		`,
	}
	objects = append(objects, partial_4_4)

	if er := NewModification().AddObjects(objects...).Execute(client); er != nil {
		t.Fatal(er)
	}
}
