require 'rubygems'
require 'bundler'

Bundler.require

post '/secret' do
  "#{request.body.read == "password"}"
end

get '/topsecret' do
  "Super secret information\n"
end
