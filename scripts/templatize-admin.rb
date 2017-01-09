#!/usr/bin/env ruby

meta = /<meta name="(gateway-ui\/config\/environment)" content="([^"]*)" \/>/

path = ARGV[0]
file = File.read(path)

file.gsub!(meta) do |match|
<<-HTML
    <meta name="#{$1}" content="{{interpolate #{$2.dump}}}" />
HTML
end

file.gsub!("ga('create', 'GOOGLE_ANALYTICS_TRACKING_ID', 'auto');", "ga('create', '{{analytics}}', 'auto');")

File.write("#{path}.template", file)
