require 'rubygems'
require 'bundler'

Bundler.require

get '/' do
  "Hello, world!\n"
end

post '/' do
  "Hello, #{request.body.read}!\n"
end
