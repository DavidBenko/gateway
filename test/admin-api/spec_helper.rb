require 'airborne'

Airborne.configure do |config|
  config.base_url = "localhost:5000/admin"
end

def clear_db!
  get "/accounts"
  json_body[:accounts].each do |account|
    delete "/accounts/#{account[:id]}"
  end
end

def login(email, pw)
  post "/sessions", {email: email, password: pw}
  cookie = response.cookies.first
  Airborne.configuration.headers = {cookies: { cookie[0] => cookie[1].gsub("%3D","=")}}
end

def logout!
  Airborne.configuration.headers = nil
end

def fixtures
  {
    accounts: {
      lulz: { account: { name: "LulzCorp" } },
      foo:  { account: { name: "Foo Corp" } },
      bar:  { account: { name: "Bar Corp" } },
    }
  }
end