#!/usr/bin/env ruby

require 'rubygems'
require 'uri'

meta = /<meta name="([^"]*)" content="([^"]*)" \/>/

path = ARGV[0]
file = File.read(path)

file.gsub!(meta) { |match| %[<meta name="#{$1}" content="{{replacePath #{$2.dump}}}" />] }

File.write("#{path}.template", file)