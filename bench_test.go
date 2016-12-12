package cobalt_test

import (
	"net/http/httptest"
	"testing"

	"github.com/ardanlabs/cobalt"
)

var data = `[
	{
		"input_index": 0,
		"candidate_index": 0,
		"addressee": "Apple Inc",
		"delivery_line_1": "1 Infinite Loop",
		"delivery_line_2": "PO Box 42",
		"last_line": "Cupertino CA 95014-2083",
		"delivery_point_barcode": "950142083017",
		"components": {
			"primary_number": "1",
			"street_name": "Infinite",
			"street_suffix": "Loop",
			"city_name": "Cupertino",
			"state_abbreviation": "CA",
			"zipcode": "95014",
			"plus4_code": "2083",
			"delivery_point": "01",
			"delivery_point_check_digit": "7"
		},
		"metadata": {
			"record_type": "S",
			"county_fips": "06085",
			"county_name": "Santa Clara",
			"carrier_route": "C067",
			"congressional_district": "15",
			"rdi": "Commercial",
			"latitude": 37.33118,
			"longitude": -122.03062,
			"precision": "Zip9"
		},
		"analysis": {
			"dpv_match_code": "Y",
			"dpv_footnotes": "AABB",
			"dpv_cmra": "N",
			"dpv_vacant": "N",
			"active": "Y"
		}
	}
	]`

func BenchmarkContextRequest(b *testing.B) {
	path := "/Hello/:name/World"
	c := cobalt.New(&JSONEncoder{})

	mw := func(h cobalt.Handler) cobalt.Handler {
		return func(c *cobalt.Context) {
			c.SetData("DATA", data)
			h(c)
		}
	}

	h := func(ctx *cobalt.Context) {
		v := ctx.GetData("DATA").(string)
		ctx.Response.Write([]byte(v))
	}

	c.Get(path, h, mw)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		r := NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		b.StartTimer()
		c.ServeHTTP(w, r)
	}
	b.ReportAllocs()
}
