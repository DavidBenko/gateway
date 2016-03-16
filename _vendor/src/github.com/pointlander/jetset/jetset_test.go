// Copyright 2015 The jetset Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jetset

import (
	"testing"
)

func TestCompress(t *testing.T) {
	items, comp := []uint64{1, 3, 4, 5, 7, 8, 9, 10, 11, 12, 14, 15, 16, 17, 20, 22, 24, 26, 28}, newCompressor()
	for _, item := range items {
		comp.compress(item)
	}
	set := comp.close()
	decomp := newDecompressor(set)
	for _, item := range items {
		d, status := decomp.decompress()
		if !status || d != item {
			t.Errorf("compression/decompression error")
			break
		}
	}
}

func makeTestSetA() Set {
	set := Set{}
	set = set.Add(10)
	set = set.Add(11)
	set = set.Add(11)
	set = set.Add(12)
	set = set.Add(14)
	set = set.Add(15)
	set = set.Add(16)
	set = set.Add(17)
	set = set.Add(20)
	set = set.Add(1)
	set = set.Add(1)
	set = set.Add(3)
	set = set.Add(4)
	set = set.Add(5)
	set = set.Add(7)
	set = set.Add(8)
	set = set.Add(9)
	set = set.Add(22)
	set = set.Add(21)
	return set
}

func makeTestSetB() Set {
	set := Set{}
	set = set.Add(10)
	set = set.Add(12)
	set = set.Add(15)
	set = set.Add(17)
	set = set.Add(1)
	set = set.Add(1)
	set = set.Add(4)
	set = set.Add(7)
	set = set.Add(9)
	set = set.Add(21)
	set = set.Add(23)
	set = set.Add(24)
	set = set.Add(28)
	return set
}

func TestAdd(t *testing.T) {
	items, set := []uint64{1, 3, 4, 5, 7, 8, 9, 10, 11, 12, 14, 15, 16, 17, 20, 21, 22}, makeTestSetA()
	decomp := newDecompressor(set)
	for _, item := range items {
		d, status := decomp.decompress()
		if !status || d != item {
			t.Errorf("add error")
			break
		}
	}
}

func TestAddRange(t *testing.T) {
	items, set := []uint64{1, 3, 4, 5, 7, 8, 9, 10, 11, 12, 14, 15, 16, 17, 20, 21, 22, 23, 24, 25, 26}, makeTestSetA().AddRange(21, 26)
	decomp := newDecompressor(set)
	for _, item := range items {
		d, status := decomp.decompress()
		if !status || d != item {
			t.Errorf("add range error")
			break
		}
	}
}

func TestHas(t *testing.T) {
	set := makeTestSetA()
	if !set.Has(11) {
		t.Errorf("set should include 11")
	}
	if set.Has(13) {
		t.Errorf("set shouldn't have 13")
	}
	if !set.Has(16) {
		t.Errorf("set should have 16")
	}
}

func TestComplement(t *testing.T) {
	set := makeTestSetA()
	set = set.Complement(0x110000)
	if set.Has(11) {
		t.Errorf("set shouldn't have 11")
	}
	if !set.Has(13) {
		t.Errorf("set should have 13")
	}
	if set.Has(16) {
		t.Errorf("set shouldn't have 16")
	}
}

func TestUnion(t *testing.T) {
	a, b := makeTestSetA(), makeTestSetB()
	set := a.Union(b)
	if !set.Has(10) {
		t.Errorf("set should have 10")
	}
	if !set.Has(28) {
		t.Errorf("set should have 28")
	}
	if !set.Has(9) {
		t.Errorf("set should have 9")
	}
}

func TestIntersection(t *testing.T) {
	a, b := makeTestSetA(), makeTestSetB()
	set := a.Intersection(b)
	if !set.Has(10) {
		t.Errorf("set should have 10")
	}
	if set.Has(28) {
		t.Errorf("set shouldn't have 28")
	}
	if !set.Has(9) {
		t.Errorf("set should have 9")
	}
	if set.Has(11) {
		t.Errorf("set should have 11")
	}
}

func TestIntersects(t *testing.T) {
	a, b := makeTestSetA(), makeTestSetB()
	if !a.Intersects(b) {
		t.Errorf("a should intersect b")
	}
}

func TestLen(t *testing.T) {
	a, b := makeTestSetA(), makeTestSetB()
	if a.Len() != 17 {
		t.Errorf("length of set should be 17")
	}
	if b.Len() != 12 {
		t.Errorf("length of set should be 12")
	}
}
