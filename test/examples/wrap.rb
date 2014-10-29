#!/usr/bin/env ruby

require 'json'

hash = {}
wrapper = nil
filename = ARGV.shift
ARGV.each do |arg|
  k, v = arg.split("=")
  v = v.to_i if k == "id"
  if k == "wrapper"
    wrapper = v
  else
    hash[k] = v
  end
end
hash["script"] = File.read(filename)
hash = {wrapper => hash} if wrapper
puts hash.to_json