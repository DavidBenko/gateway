package search

import (
	"reflect"
	"testing"
)

func TestTermFacetResultsMerge(t *testing.T) {

	fr1 := &FacetResult{
		Field:   "type",
		Total:   100,
		Missing: 25,
		Other:   25,
		Terms: []*TermFacet{
			&TermFacet{
				Term:  "blog",
				Count: 25,
			},
			&TermFacet{
				Term:  "comment",
				Count: 24,
			},
			&TermFacet{
				Term:  "feedback",
				Count: 1,
			},
		},
	}
	fr1Only := &FacetResult{
		Field:   "category",
		Total:   97,
		Missing: 22,
		Other:   15,
		Terms: []*TermFacet{
			&TermFacet{
				Term:  "clothing",
				Count: 35,
			},
			&TermFacet{
				Term:  "electronics",
				Count: 25,
			},
		},
	}
	frs1 := FacetResults{
		"types":      fr1,
		"categories": fr1Only,
	}

	fr2 := &FacetResult{
		Field:   "type",
		Total:   100,
		Missing: 25,
		Other:   25,
		Terms: []*TermFacet{
			&TermFacet{
				Term:  "blog",
				Count: 25,
			},
			&TermFacet{
				Term:  "comment",
				Count: 22,
			},
			&TermFacet{
				Term:  "flag",
				Count: 3,
			},
		},
	}
	frs2 := FacetResults{
		"types": fr2,
	}

	expectedFr := &FacetResult{
		Field:   "type",
		Total:   200,
		Missing: 50,
		Other:   51,
		Terms: []*TermFacet{
			&TermFacet{
				Term:  "blog",
				Count: 50,
			},
			&TermFacet{
				Term:  "comment",
				Count: 46,
			},
			&TermFacet{
				Term:  "flag",
				Count: 3,
			},
		},
	}
	expectedFrs := FacetResults{
		"types":      expectedFr,
		"categories": fr1Only,
	}

	frs1.Merge(frs2)
	frs1.Fixup("types", 3)
	if !reflect.DeepEqual(frs1, expectedFrs) {
		t.Errorf("expected %v, got %v", expectedFrs, frs1)
	}
}

func TestNumericFacetResultsMerge(t *testing.T) {

	lowmed := 3.0
	medhi := 6.0
	hihigher := 9.0

	fr1 := &FacetResult{
		Field:   "rating",
		Total:   100,
		Missing: 25,
		Other:   25,
		NumericRanges: []*NumericRangeFacet{
			&NumericRangeFacet{
				Name:  "low",
				Max:   &lowmed,
				Count: 25,
			},
			&NumericRangeFacet{
				Name:  "med",
				Count: 24,
				Max:   &lowmed,
				Min:   &medhi,
			},
			&NumericRangeFacet{
				Name:  "hi",
				Count: 1,
				Min:   &medhi,
				Max:   &hihigher,
			},
		},
	}
	frs1 := FacetResults{
		"ratings": fr1,
	}

	fr2 := &FacetResult{
		Field:   "rating",
		Total:   100,
		Missing: 25,
		Other:   25,
		NumericRanges: []*NumericRangeFacet{
			&NumericRangeFacet{
				Name:  "low",
				Max:   &lowmed,
				Count: 25,
			},
			&NumericRangeFacet{
				Name:  "med",
				Max:   &lowmed,
				Min:   &medhi,
				Count: 22,
			},
			&NumericRangeFacet{
				Name:  "highest",
				Min:   &hihigher,
				Count: 3,
			},
		},
	}
	frs2 := FacetResults{
		"ratings": fr2,
	}

	expectedFr := &FacetResult{
		Field:   "rating",
		Total:   200,
		Missing: 50,
		Other:   51,
		NumericRanges: []*NumericRangeFacet{
			&NumericRangeFacet{
				Name:  "low",
				Count: 50,
				Max:   &lowmed,
			},
			&NumericRangeFacet{
				Name:  "med",
				Max:   &lowmed,
				Min:   &medhi,
				Count: 46,
			},
			&NumericRangeFacet{
				Name:  "highest",
				Min:   &hihigher,
				Count: 3,
			},
		},
	}
	expectedFrs := FacetResults{
		"ratings": expectedFr,
	}

	frs1.Merge(frs2)
	frs1.Fixup("ratings", 3)
	if !reflect.DeepEqual(frs1, expectedFrs) {
		t.Errorf("expected %#v, got %#v", expectedFrs, frs1)
	}
}

func TestDateFacetResultsMerge(t *testing.T) {

	lowmed := "2010-01-01"
	medhi := "2011-01-01"
	hihigher := "2012-01-01"

	fr1 := &FacetResult{
		Field:   "birthday",
		Total:   100,
		Missing: 25,
		Other:   25,
		DateRanges: []*DateRangeFacet{
			&DateRangeFacet{
				Name:  "low",
				End:   &lowmed,
				Count: 25,
			},
			&DateRangeFacet{
				Name:  "med",
				Count: 24,
				Start: &lowmed,
				End:   &medhi,
			},
			&DateRangeFacet{
				Name:  "hi",
				Count: 1,
				Start: &medhi,
				End:   &hihigher,
			},
		},
	}
	frs1 := FacetResults{
		"birthdays": fr1,
	}

	fr2 := &FacetResult{
		Field:   "birthday",
		Total:   100,
		Missing: 25,
		Other:   25,
		DateRanges: []*DateRangeFacet{
			&DateRangeFacet{
				Name:  "low",
				End:   &lowmed,
				Count: 25,
			},
			&DateRangeFacet{
				Name:  "med",
				Start: &lowmed,
				End:   &medhi,
				Count: 22,
			},
			&DateRangeFacet{
				Name:  "highest",
				Start: &hihigher,
				Count: 3,
			},
		},
	}
	frs2 := FacetResults{
		"birthdays": fr2,
	}

	expectedFr := &FacetResult{
		Field:   "birthday",
		Total:   200,
		Missing: 50,
		Other:   51,
		DateRanges: []*DateRangeFacet{
			&DateRangeFacet{
				Name:  "low",
				Count: 50,
				End:   &lowmed,
			},
			&DateRangeFacet{
				Name:  "med",
				Start: &lowmed,
				End:   &medhi,
				Count: 46,
			},
			&DateRangeFacet{
				Name:  "highest",
				Start: &hihigher,
				Count: 3,
			},
		},
	}
	expectedFrs := FacetResults{
		"birthdays": expectedFr,
	}

	frs1.Merge(frs2)
	frs1.Fixup("birthdays", 3)
	if !reflect.DeepEqual(frs1, expectedFrs) {
		t.Errorf("expected %#v, got %#v", expectedFrs, frs1)
	}
}
