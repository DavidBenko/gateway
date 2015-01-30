require_relative './spec_helper'

shared_examples "invalid credentials" do
  it { expect_status(400) }

  it "should not let you into private areas" do
    get "/apis"
    expect_status(401)
  end
end

describe "sessions" do
  before(:all) do
    clear_db!

    post "/accounts", fixtures[:accounts][:foo]
    expect_status(200)
    post "/accounts/#{json_body[:account][:id]}/users", fixtures[:users][:geff]
    expect_status(200)
    @geff = fixtures[:users][:geff][:user]
  end

  describe "create" do
    context "valid credentials" do
      before(:all) do
        logout!
        post "/sessions", { email: @geff[:email], password: @geff[:password] }
        set_auth_cookie
      end

      it { expect_status(200) }
      it { expect(body).to be_empty }

      it "should let you into private areas" do
        get "/apis"
        expect_status(200)
      end
    end

    context "invalid email" do
      before(:all) do
        logout!
        post "/sessions", { email: @geff[:email]+"x", password: @geff[:password] }
        set_auth_cookie
      end

      it_behaves_like "invalid credentials"
      it { expect_json("error", "No user with that email") }
    end

    context "invalid password" do
      before(:all) do
        logout!
        post "/sessions", { email: @geff[:email], password: @geff[:password]+"x" }
        set_auth_cookie
      end

      it_behaves_like "invalid credentials"
      it { expect_json("error", "Invalid password") }
    end

    context "invalid while logged in" do
      before(:all) do
        login @geff[:email], @geff[:password]
        post "/sessions", { email: @geff[:email]+"x", password: @geff[:password] }
        set_auth_cookie
      end

      it_behaves_like "invalid credentials"
    end
  end

  describe "destroy" do
    before(:all) do
      login @geff[:email], @geff[:password]
      delete "/sessions"
      set_auth_cookie
    end

    it { expect_status(200) }
    it { expect(body).to be_empty }

    it "should not let you into private areas" do
      get "/apis"
      expect_status(401)
    end
  end

  def set_auth_cookie
    unless response && response.cookies && response.cookies.first && response.cookies.first.size >= 2
      Airborne.configuration.headers = {}
      return
    end
    
    cookie = response.cookies.first
    Airborne.configuration.headers = {cookies: { cookie[0] => cookie[1].gsub("%3D","=")}}
  end
end
