#!/usr/bin/env ruby

require 'json'

hash = {}
filename = ARGV.shift
ARGV.each do |arg|
  k, v = arg.split("=")
  v = v.to_i if k == "id"
  hash[k] = v
end

hash["script"] = File.read(filename)
puts hash.to_json