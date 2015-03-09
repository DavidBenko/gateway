#!/usr/bin/env ruby

require 'rubygems'
require 'erb'

version, _, commit = *`git describe --long`.strip.split('-')

filename = "./version_gen.go"
output = File.open(filename, "w")

template = <<-ERB

package version

/*****************************************************
 *****************************************************
 ***                                               ***
 *** This is generated code. Do not edit directly. ***
 ***                                               ***
 *****************************************************
 *****************************************************/

func Name() string {
  return "#{version}"
}

func Commit() string {
  return "#{commit}"
}

ERB

output.write ERB.new(template).result
output.close

`goimports -w ./#{filename}`
