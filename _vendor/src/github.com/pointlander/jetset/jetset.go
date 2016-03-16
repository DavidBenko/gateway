// Copyright 2015 The jetset Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jetset

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/pointlander/compress"
)

type Set struct {
	data []byte
	length int
}

func (set Set) String() string {
	buffer, in, output := bytes.NewBuffer(set.data), make(chan []byte, 1), make([]byte, set.length)
	in <- output
	close(in)
	compress.BijectiveBurrowsWheelerDecoder(in).MoveToFrontRunLengthDecoder().AdaptiveDecoder().Decode(buffer)

	codes, code, reader, space := "[", uint64(0), bytes.NewReader(output), ""
	next := func() error {
		return binary.Read(reader, binary.BigEndian, &code)
	}
	for err := next(); err == nil; err = next() {
		codes += space + fmt.Sprintf("%v", code)
		space = " "
	}
	return codes + "]"
}

type decompressor struct {
	buffer *bytes.Reader
	adding bool
	current, count, offset uint64
}

func newDecompressor(set Set) *decompressor {
	if len(set.data) == 0 {
		return &decompressor{buffer: bytes.NewReader(nil)}
	}
	buffer, in, output := bytes.NewBuffer(set.data), make(chan []byte, 1), make([]byte, set.length)
	in <- output
	close(in)
	compress.BijectiveBurrowsWheelerDecoder(in).MoveToFrontRunLengthDecoder().AdaptiveDecoder().Decode(buffer)
	return &decompressor{buffer: bytes.NewReader(output)}
}

func (d *decompressor) decompress() (uint64, bool) {
	if d.adding {
		if d.count > 0 {
			d.count--
			d.current += d.offset
			return d.current, true
		} else {
			var code uint64
			err := binary.Read(d.buffer, binary.BigEndian, &code)
			if err == nil {
				if code == 0 {
					err := binary.Read(d.buffer, binary.BigEndian, &d.count)
					if err != nil {
						panic(err)
					}

					err = binary.Read(d.buffer, binary.BigEndian, &d.offset)
					if err != nil {
						panic(err)
					}

					d.count--
					d.current += d.offset
					return d.current, true
				} else {
					d.current += code
					return d.current, true
				}
			} else {
				return 0, false
			}
		}
	} else {
		err := binary.Read(d.buffer, binary.BigEndian, &d.current)
		if err == nil {
			d.adding = true
			return d.current, true
		} else {
			return 0, false
		}
	}
}

type compressor struct {
	buffer *bytes.Buffer
	writing bool
	current, count, offset uint64
}

func newCompressor() *compressor {
	return &compressor{buffer: &bytes.Buffer{}}
}

func (c *compressor) compress(e uint64) {
	if c.writing {
		offset := e - c.current
		if c.offset == offset {
			c.count++
		} else if c.count > 0 {
			binary.Write(c.buffer, binary.BigEndian, uint64(0))
			binary.Write(c.buffer, binary.BigEndian, c.count + 1)
			binary.Write(c.buffer, binary.BigEndian, c.offset)
			c.count, c.offset = 0, offset
		} else {
			binary.Write(c.buffer, binary.BigEndian, c.offset)
			c.offset = offset
		}
	} else {
		c.writing, c.offset = true, e
	}

	c.current = e
}

func (c *compressor) close() Set {
	if c.writing {
		if c.count > 0 {
			binary.Write(c.buffer, binary.BigEndian, uint64(0))
			binary.Write(c.buffer, binary.BigEndian, c.count + 1)
			binary.Write(c.buffer, binary.BigEndian, c.offset)
		} else {
			binary.Write(c.buffer, binary.BigEndian, c.offset)
		}
		buffer, in := &bytes.Buffer{}, make(chan []byte, 1)
		in <- c.buffer.Bytes()
		close(in)
		compress.BijectiveBurrowsWheelerCoder(in).MoveToFrontRunLengthCoder().AdaptiveCoder().Code(buffer)
		return Set{data: buffer.Bytes(), length: c.buffer.Len()}
	}
	return Set{}
}

func (s Set) Copy() Set {
	cp := make([]byte, len(s.data))
	copy(cp, s.data)
	return Set{data: cp, length: s.length}
}

func (s Set) Add(e uint64) Set {
	comp, decomp, found := newCompressor(), newDecompressor(s), false
	d, status := decomp.decompress()
	if !status {
		comp.compress(e)
		return comp.close()
	}
	for status {
 		if !found && e == d {
			found = true
		} else if !found && e < d {
			comp.compress(e)
			found = true
		}
		comp.compress(d)
		d, status = decomp.decompress()
	}
	if !found {
		comp.compress(e)
	}
	return comp.close()
}

func (a Set) AddRange(begin, end uint64) Set {
	comp, decomp_a, d_b := newCompressor(), newDecompressor(a), begin
	d_a, status_a := decomp_a.decompress()
	for status_a && d_b <= end {
		if d_a == d_b {
			comp.compress(d_a)
			d_a, status_a = decomp_a.decompress()
			d_b++
		} else if d_a < d_b {
			comp.compress(d_a)
			d_a, status_a = decomp_a.decompress()
		} else {
			comp.compress(d_b)
			d_b++
		}
	}
	for status_a {
		comp.compress(d_a)
		d_a, status_a = decomp_a.decompress()
	}
	for d_b <= end {
		comp.compress(d_b)
		d_b++
	}
	return comp.close()
}

func (s Set) Has(e uint64) bool {
	decomp := newDecompressor(s)
	for item, status := decomp.decompress(); status; item, status = decomp.decompress() {
		if item == e {
			return true
		}
	}
	return false
}

func (s Set) Complement(max uint64) Set {
	comp, decomp, i := newCompressor(), newDecompressor(s), uint64(0)
	d, status := decomp.decompress()
	for i <= max && status {
		if i < d {
			comp.compress(i)
		} else if i == d {
			d, status = decomp.decompress()
		}
		i++
	}
	for i <= max {
		comp.compress(i)
		i++
	}
	return comp.close()
}

func (a Set) Union(b Set) Set {
	comp, decomp_a, decomp_b := newCompressor(), newDecompressor(a), newDecompressor(b)
	d_a, status_a := decomp_a.decompress()
	d_b, status_b := decomp_b.decompress()
	for status_a && status_b {
		if d_a == d_b {
			comp.compress(d_a)
			d_a, status_a = decomp_a.decompress()
			d_b, status_b = decomp_b.decompress()
		} else if d_a < d_b {
			comp.compress(d_a)
			d_a, status_a = decomp_a.decompress()
		} else {
			comp.compress(d_b)
			d_b, status_b = decomp_b.decompress()
		}
	}
	for status_a {
		comp.compress(d_a)
		d_a, status_a = decomp_a.decompress()
	}
	for status_b {
		comp.compress(d_b)
		d_b, status_b = decomp_b.decompress()
	}
	return comp.close()
}

func (a Set) Intersection(b Set) Set {
	comp, decomp_a, decomp_b := newCompressor(), newDecompressor(a), newDecompressor(b)
	d_a, status_a := decomp_a.decompress()
	d_b, status_b := decomp_b.decompress()
	for status_a && status_b {
		if d_a == d_b {
			comp.compress(d_a)
			d_a, status_a = decomp_a.decompress()
			d_b, status_b = decomp_b.decompress()
		} else if d_a < d_b {
			d_a, status_a = decomp_a.decompress()
		} else {
			d_b, status_b = decomp_b.decompress()
		}
	}
	return comp.close()
}

func (a Set) Intersects(b Set) bool {
	decomp_a, decomp_b := newDecompressor(a), newDecompressor(b)
	d_a, status_a := decomp_a.decompress()
	d_b, status_b := decomp_b.decompress()
	for status_a && status_b {
		if d_a == d_b {
			return true
		} else if d_a < d_b {
			d_a, status_a = decomp_a.decompress()
		} else {
			d_b, status_b = decomp_b.decompress()
		}
	}
	return false
}

func (s Set) Len() int {
	length, decomp := 0, newDecompressor(s)
	for _, status := decomp.decompress(); status; _, status = decomp.decompress() {
		length++
	}
	return length
}
