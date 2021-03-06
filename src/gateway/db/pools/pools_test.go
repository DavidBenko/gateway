package pools_test

import (
	"testing"

	"gateway/db"
	"gateway/db/pools"
	dbtest "gateway/db/testing"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

func Test(t *testing.T) { gc.TestingT(t) }

type PoolsSuite struct {
	specs map[string]db.Specifier
}

var _ = gc.Suite(&PoolsSuite{})

func (s *PoolsSuite) SetUpTest(c *gc.C) {
	s.specs = dbtest.Specs()
}

func (s *PoolsSuite) TestConnect(c *gc.C) {
	for i, t := range []struct {
		should    string
		given     db.Specifier
		expectErr string
	}{{
		should: "connect given a Specifier",
		given:  s.specs["simple-ok"],
	}, {
		should:    "pass back an error from NewDB",
		given:     s.specs["simple-newdb-err"],
		expectErr: `NewDB error`,
	}} {
		c.Logf("test %d: should %s", i, t.should)
		p := dbtest.MakePool()

		// Assert NewDB worked correctly (db is our expected result)
		db, err := t.given.NewDB()
		if t.expectErr != "" {
			c.Assert(err, gc.ErrorMatches, t.expectErr)
		} else {
			c.Assert(err, gc.IsNil)
		}

		// Now Connect
		d, err := pools.Connect(p, t.given)
		if t.expectErr != "" {
			c.Assert(err, gc.ErrorMatches, t.expectErr)
			continue
		}
		c.Assert(err, gc.IsNil)

		// Check that the DB returned by Connect matches the one
		// generated by the spec
		c.Check(d.Spec(), jc.DeepEquals, db.Spec())

		// Check that the DB we got has been added to the pool
		got, ok := p.DBs[t.given.UniqueServer()]
		c.Check(ok, gc.Equals, true)

		// Check that the DB we added to the pool is the one we expect
		c.Check(got.Spec(), jc.DeepEquals, db.Spec())
	}
}

func (s *PoolsSuite) TestFlushEntry(c *gc.C) {
	for i, t := range []struct {
		should string
		given  []db.Specifier
		flush  []db.Specifier
		expect []db.Specifier
	}{{
		should: "connect some Specifiers and flush them all",
		given: []db.Specifier{
			s.specs["simple-ok"],
			s.specs["simple-ok"],
			s.specs["simple-ok"],
		},
		flush: []db.Specifier{
			s.specs["simple-ok"],
		},
		expect: []db.Specifier{},
	}, {
		should: "connect some Specifiers and flush a few",
		given: []db.Specifier{
			s.specs["simple-ok"],
			s.specs["simple-ok-1"],
			s.specs["simple-ok-2"],
			s.specs["simple-ok-3"],
		},
		flush: []db.Specifier{
			s.specs["simple-ok"],
			s.specs["simple-ok-2"],
		},
		expect: []db.Specifier{
			s.specs["simple-ok-1"],
			s.specs["simple-ok-3"],
		},
	}} {
		c.Logf("test %d: should %s", i, t.should)
		p := dbtest.MakePool()

		d, err := dbtest.Connect(p, t.given)
		c.Assert(err, gc.IsNil)

		// Check that the DB(s) have been added to the pool
		for i, spec := range t.given {
			got, ok := p.DBs[spec.UniqueServer()]
			c.Check(ok, gc.Equals, true)
			// Check that what was added to the pool is correct
			c.Check(got.Spec(), jc.DeepEquals, d[i].Spec())
		}

		// Flush it
		for _, s := range t.flush {
			pools.FlushEntry(p, s)
		}

		// Check that the pool now contains the expected set
		c.Check(len(t.expect), gc.Equals, len(p.DBs))
		for _, spec := range t.expect {
			// We know we have the correct one if it's present,
			// because that is tested in TestConnect.
			_, ok := p.DBs[spec.UniqueServer()]
			c.Check(ok, gc.Equals, true)
		}
	}
}

func (s *PoolsSuite) TestIterator(c *gc.C) {
	for i, t := range []struct {
		should string
		given  []db.Specifier
		expect []db.Specifier
	}{{
		should: "connect some Specifiers and flush them all",
		given: []db.Specifier{
			s.specs["simple-ok"],
			s.specs["simple-ok"],
			s.specs["simple-ok"],
		},
		expect: []db.Specifier{
			s.specs["simple-ok"],
		},
	}, {
		should: "connect some Specifiers and flush a few",
		given: []db.Specifier{
			s.specs["simple-ok"],
			s.specs["simple-ok-1"],
			s.specs["simple-ok-2"],
			s.specs["simple-ok-3"],
		},
		expect: []db.Specifier{
			s.specs["simple-ok"],
			s.specs["simple-ok-1"],
			s.specs["simple-ok-2"],
			s.specs["simple-ok-3"],
		},
	}} {
		c.Logf("test %d: should %s", i, t.should)
		p := dbtest.MakePool()

		d, err := dbtest.Connect(p, t.given)
		c.Assert(err, gc.IsNil)

		// Check that the DB(s) have been added to the pool
		for i, spec := range t.given {
			got, ok := p.DBs[spec.UniqueServer()]
			c.Check(ok, gc.Equals, true)
			// Check that what was added to the pool is correct
			c.Check(got.Spec(), jc.DeepEquals, d[i].Spec())
		}

		c.Check(len(t.expect), gc.Equals, len(p.DBs))

		// Check that the Iterator method returns the set of all
		// entries
		expectSeenCount := make(map[string]int)
		for _, spec := range t.expect {
			expectSeenCount[spec.UniqueServer()] += 1
		}

		seenCount := make(map[string]int)
		for spec := range p.Iterator() {
			seenCount[spec.UniqueServer()] += 1
		}

		c.Assert(len(seenCount), gc.Equals, len(expectSeenCount))

		for k, count := range expectSeenCount {
			if _, ok := seenCount[k]; !ok {
				c.Fail()
			}
			c.Check(count, gc.Equals, seenCount[k])
		}
	}
}
