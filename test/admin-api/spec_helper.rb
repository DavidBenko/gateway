require 'airborne'

Airborne.configure do |config|
  config.base_url = "http://localhost:5000/admin"
end

def clear_db!
  get "/accounts"
  json_body[:accounts].each do |account|
    delete "/accounts/#{account[:id]}"
  end
end

def clear_users!(id, options={})
  cept = options[:cept]
  if cept.kind_of? Fixnum
    cept = [cept]
  end

  get "/accounts/#{id}/users"
  json_body[:users].each do |user|
    next if cept && cept.include?(user[:id])
    delete "/accounts/#{id}/users/#{user[:id]}"
  end
end

def clear_apis!
  get "/apis"
  json_body[:apis].each do |api|
    delete "/apis/#{api[:id]}"
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
      geff:  { name: "Geff",  email: "g@ffery.com", password: "password", password_confirmation: "password", admin: true, confirmed: true },
      brain: { name: "Brain", email: "br@in.com",   password: "password", password_confirmation: "password", confirmed: true },
      poter: { name: "Poter", email: "p@ter.com",   password: "password", password_confirmation: "password", confirmed: true },
    },
    apis: {
      widgets: {
        name: "Widgets",
        description: "Lots of widgets here",
        cors_allow_origin: "*",
        cors_allow_headers: "content-type, accept",
        cors_allow_credentials: true,
        cors_request_headers: "*",
        cors_max_age: 600
      },
      gadgets: {
        name: "Gadgets",
        description: "No widgets",
        cors_allow_origin: "*",
        cors_allow_headers: "content-type, accept",
        cors_allow_credentials: true,
        cors_request_headers: "*",
        cors_max_age: 600
      },
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
