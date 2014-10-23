require 'rubygems'
require 'bundler'

Bundler.require

get '/' do
  custom_headers = request.env.keys.select{|k| k =~ /^HTTP_X/}
  [
    200, 
    { "X-Foo" => "Foo", "X-Bar" => ["Bar", "Baz"] }, 
    request.env.select{|k,v| custom_headers.include?(k)}.inspect+"\n"
  ]
end
