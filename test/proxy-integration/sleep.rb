require 'sinatra'

get '/timeout' do
  sleep(61)
  "finished!"
end