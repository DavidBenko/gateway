require 'rubygems'
require 'bundler'
require 'json'

Bundler.require

get '/' do
  "Hello, world!\n"
end

post '/' do
  "Hello, #{request.body.read}!\n"
end

post '/echo' do
  request.body.read
end

get '/foo' do
  {foo: "foo!"}.to_json
end

get '/bar' do
  {bar: "bar!"}.to_json
end

get '/headers' do
  custom_headers = request.env.keys.select{|k| k =~ /^HTTP_X/}
  [
    200, 
    { "X-Foo" => "Foo", "X-Bar" => ["Bar", "Baz"] }, 
    request.env.select{|k,v| custom_headers.include?(k)}.inspect+"\n"
  ]
end

post '/secret' do
  "#{request.body.read == "password"}"
end

get '/topsecret' do
  "Super secret information\n"
end