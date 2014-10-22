require 'rubygems'
require 'bundler'
require 'json'

Bundler.require

get '/foo' do
  {foo: "foo!"}.to_json
end

get '/bar' do
  {bar: "bar!"}.to_json
end
