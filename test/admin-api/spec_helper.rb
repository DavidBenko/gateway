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

def clear_users!(id)
  get "/accounts/#{id}/users"
  json_body[:users].each do |user|
    delete "/accounts/#{id}/users/#{user[:id]}"
  end
end

def login(email, pw)
  post "/sessions", {email: email, password: pw}
  expect_status(200)
  cookie = response.cookies.first
  Airborne.configuration.headers = {cookies: { cookie[0] => cookie[1].gsub("%3D","=")}}
end

def logout!
  Airborne.configuration.headers = nil
end

def fixtures
  {
    accounts: {
      lulz: { name: "LulzCorp" },
      foo:  { name: "Foo Corp" },
      bar:  { name: "Bar Corp" },
    },
    users: {
      geff:  { name: "Geff",  email: "g@ffery.com", password: "password", password_confirmation: "password" },
      brain: { name: "Brain", email: "br@in.com",   password: "password", password_confirmation: "password" },
    }
  }
end

class Hash
  def without(*keys)
    cpy = self.dup
    keys.each { |key| cpy.delete(key) }
    cpy
  end
end

shared_examples "invalid json" do
  it { expect_status(400) }
  it { expect_json("error", "unexpected end of JSON input") }
end
