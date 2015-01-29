require_relative './spec_helper'

describe "accounts" do
  describe "index" do
    context "empty" do
      before(:all) do
        get '/accounts'
      end

      it "should return 200" do
        expect(response.code).to eq(200)
      end

      it "should return an array in 'accounts'" do
        expect_json_types({accounts: :array})
      end
    end

    context "with data" do
      before(:all) do
        clear_db!
      end

      it "should return all accounts" do
        expect_count_to_equal(0)
        post '/accounts', fixtures[:accounts][:foo]
        expect_count_to_equal(1)
        post '/accounts', fixtures[:accounts][:bar]
        expect_count_to_equal(2)
      end
    end

    def expect_count_to_equal(num)
      get '/accounts'
      expect(json_body[:accounts].size).to equal(num)
    end
  end

  describe "create" do
    context "with valid data" do
      before(:all) do
        clear_db!
        post '/accounts', fixtures[:accounts][:lulz]
      end

      it "should return 200" do
        expect(response.code).to eq(200)
      end

      it "should return the new ID" do
        expect(json_body[:account][:id]).to_not be_nil
      end

      it "should return the name" do
        expect(json_body[:account][:name]).to eq("LulzCorp")
      end
    end

    context "with invalid json" do
      before(:all) do
        post '/accounts', '{"account":{"name":"LulzCo'
      end

      it "should return 400" do
        expect(response.code).to eq(400)
      end

      it "should return an error" do
        expect(json_body[:error]).to eq("unexpected end of JSON input")
      end
    end

    context "without a name" do
      before(:all) do
        post '/accounts',  {account: { name: ""}}
      end

      it "should return 400" do
        expect(response.code).to eq(400)
      end

      it "should return an error" do
        expect(json_body[:errors]).to eq({name: ["must not be blank"]})
      end
    end

    context "with a duplicate name" do
      before(:all) do
        clear_db!
        post '/accounts', fixtures[:accounts][:lulz]
        expect(response.code).to eq(200)
        post '/accounts',  fixtures[:accounts][:lulz]
      end

      it "should return 400" do
        expect(response.code).to eq(400)
      end

      it "should return an error" do
        expect(json_body[:errors]).to eq({name: ["is already taken"]})
      end
    end
  end

  describe "show" do
    before(:all) do
      clear_db!
      post '/accounts', fixtures[:accounts][:lulz]
      expect(response.code).to eq(200)
      @id = json_body[:account][:id]
    end

    context "existing" do
      before(:all) do
        get "/accounts/#{@id}"
      end

      it "should return 200" do
        expect(response.code).to eq(200)
      end

      it "should return the id" do
        expect(json_body[:account][:id]).to eq(@id)
      end

      it "should return the name" do
        expect(json_body[:account][:name]).to eq("LulzCorp")
      end
    end


    context "non-existing" do
      before(:all) do
        get "/accounts/#{@id+1}"
      end

      it "should return 404" do
        expect(response.code).to eq(404)
      end
    end
  end

  describe "update" do
    def setup_account
      clear_db!
      post '/accounts', fixtures[:accounts][:lulz]
      expect(response.code).to eq(200)
      @id = json_body[:account][:id]
    end

    context "with valid data" do
      before(:all) do
        setup_account
        put "/accounts/#{@id}", fixtures[:accounts][:foo]
      end

      it "should return 200" do
        expect(response.code).to eq(200)
      end

      it "should return the same ID" do
        expect(json_body[:account][:id]).to eq(@id)
      end

      it "should return the new name" do
        expect(json_body[:account][:name]).to eq("Foo Corp")
      end
    end

    context "with invalid json" do
      before(:all) do
        setup_account
        put "/accounts/#{@id}", '{"account":{"name":"LulzCo'
      end

      it "should return 400" do
        expect(response.code).to eq(400)
      end

      it "should return an error" do
        expect(json_body[:error]).to eq("unexpected end of JSON input")
      end
    end

    context "without a name" do
      before(:all) do
        setup_account
        put "/accounts/#{@id}", {account: { name: ""}}
      end

      it "should return 400" do
        expect(response.code).to eq(400)
      end

      it "should return an error" do
        expect(json_body[:errors]).to eq({name: ["must not be blank"]})
      end
    end

    context "with a duplicate name" do
      before(:all) do
        setup_account
        post '/accounts', {account: { name: "BooBoo Butt"}}
        expect(response.code).to eq(200)
        put "/accounts/#{@id}", {account: { name: "BooBoo Butt"}}
      end

      it "should return 400" do
        expect(response.code).to eq(400)
      end

      it "should return an error" do
        expect(json_body[:errors]).to eq({name: ["is already taken"]})
      end
    end

    context "non-existing" do
      before(:all) do
        setup_account
        put "/accounts/#{@id+1}", {account: { name: "BooBoo Butt"}}
      end

      it "should return 404" do
        expect(response.code).to eq(404)
      end
    end
  end

  describe "delete" do
    before(:all) do
      clear_db!
      post '/accounts', fixtures[:accounts][:lulz]
      expect(response.code).to eq(200)
      @id = json_body[:account][:id]
    end

    context "existing" do
      before(:all) do
        delete "/accounts/#{@id}"
      end

      it "should return 200" do
        expect(response.code).to eq(200)
      end

      it "should return nothing" do
        expect(body).to eq("")
      end

      it "should remove the item" do
        get "/accounts/#{@id}"
        expect(response.code).to eq(404)
      end
    end

    context "non-existing" do
      before(:all) do
        delete "/accounts/#{@id+1}"
      end

      it "should return 404" do
        expect(response.code).to eq(404)
      end
    end
  end

  # it 'should validate types' do
  #   puts response.headers.inspect
  #   expect_json_types({accounts: :array})
  # end
  #
  # it "should do stuff" do
  #   post '/accounts', {:account => {:name => 'John Doe'}}
  #   account_id = json_body[:account][:id]
  #   puts account_id
  #   post "/accounts/#{account_id}/users", {:user => {:name => "Tester", :email => "test@foo.com", :password => "foobar", :password_confirmation => "foobar"}}
  #   user_id = json_body[:user][:id]
  #   puts json_body.inspect
  #   login "test@foo.com", "foobar"
  #   # puts cookies.inspect
  #   # cooks = { "__ap_gateway" => cookies.first[1].gsub("%3D","=")}
  #   # puts cooks.inspect
  #   get "/users/#{user_id}"#, {cookies: cooks}
  #   puts body.inspect
  #   # puts json_body.inspect
  # end

  # it 'should validate values' do
  #   get 'http://example.com/api/v1/simple_get' #json api that returns { "name" : "John Doe" }
  #   expect_json({:name => "John Doe"})
  # end
end
