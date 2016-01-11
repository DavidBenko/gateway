#!/usr/bin/env ruby

meta = /<meta name="(gateway\/config\/environment)" content="([^"]*)" \/>/

path = ARGV[0]
file = File.read(path)

file.gsub!(meta) do |match|
<<-HTML
    <meta name="#{$1}" content="{{interpolate #{$2.dump}}}" />
HTML
end

File.write("#{path}.template", file)
