/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"runtime"
	"testing"
)

func TestLooksLikePath(t *testing.T) {
	type pathTest struct {
		v    string
		want bool
	}
	tests := []pathTest{
		{"foo.com", false},
		{"127.0.0.1:234", false},
		{"foo", false},

		{"/foo", true},
		{"./foo", true},
		{"../foo", true},
	}
	if runtime.GOOS == "windows" {
		tests = append(tests,
			pathTest{`\foo`, true},
			pathTest{`.\foo`, true},
			pathTest{`..\foo`, true},
			pathTest{`C:/dir`, true},
			pathTest{`C:\dir`, true},
			pathTest{`//server/share/dir`, true},
			pathTest{`\\server\share\dir`, true},
		)
	}
	for _, tt := range tests {
		got := looksLikePath(tt.v)
		if got != tt.want {
			t.Errorf("looksLikePath(%q) = %v; want %v", tt.v, got, tt.want)
		}
	}
}
