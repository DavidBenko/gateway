require 'airborne'
require_relative './monkey'

Airborne.configure do |config|
  config.base_url = "localhost:5000"
end
