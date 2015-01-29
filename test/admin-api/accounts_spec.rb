require_relative './spec_helper'

shared_examples "missing account" do
  it { expect(response.code).to eq(404) }
  it { expect(json_body[:error]).to eq("No account matches") }
end

describe "accounts" do
  describe "index" do
    context "empty" do
      before(:all) do
        get '/accounts'
      end

      it { expect(response.code).to eq(200) }
      it { expect_json_types({accounts: :array}) }
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

      it { expect(response.code).to eq(200) }
      it { expect(json_body[:account][:id]).to_not be_nil }
      it { expect(json_body[:account][:name]).to eq("LulzCorp") }
    end

    context "with invalid json" do
      before(:all) do
        post '/accounts', '{"account":{"name":"LulzCo'
      end

      it { expect(response.code).to eq(400) }
      it { expect(json_body[:error]).to eq("unexpected end of JSON input") }
    end

    context "without a name" do
      before(:all) do
        post '/accounts',  {account: { name: ""}}
      end

      it { expect(response.code).to eq(400) }
      it { expect(json_body[:errors]).to eq({name: ["must not be blank"]}) }
    end

    context "with a duplicate name" do
      before(:all) do
        clear_db!
        post '/accounts', fixtures[:accounts][:lulz]
        expect(response.code).to eq(200)
        post '/accounts',  fixtures[:accounts][:lulz]
      end

      it { expect(response.code).to eq(400) }
      it { expect(json_body[:errors]).to eq({name: ["is already taken"]}) }
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

      it { expect(response.code).to eq(200) }
      it { expect(json_body[:account][:id]).to eq(@id) }
      it { expect(json_body[:account][:name]).to eq("LulzCorp") }
    end

    context "non-existing" do
      before(:all) do
        get "/accounts/#{@id+1}"
      end

      it_behaves_like "missing account"
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

      it { expect(response.code).to eq(200) }
      it { expect(json_body[:account][:id]).to eq(@id) }
      it { expect(json_body[:account][:name]).to eq("Foo Corp") }
    end

    context "with invalid json" do
      before(:all) do
        setup_account
        put "/accounts/#{@id}", '{"account":{"name":"LulzCo'
      end

      it { expect(response.code).to eq(400) }
      it { expect(json_body[:error]).to eq("unexpected end of JSON input") }
    end

    context "without a name" do
      before(:all) do
        setup_account
        put "/accounts/#{@id}", {account: { name: ""}}
      end

      it { expect(response.code).to eq(400) }
      it { expect(json_body[:errors]).to eq({name: ["must not be blank"]}) }
    end

    context "with a duplicate name" do
      before(:all) do
        setup_account
        post '/accounts', {account: { name: "BooBoo Butt"}}
        expect(response.code).to eq(200)
        put "/accounts/#{@id}", {account: { name: "BooBoo Butt"}}
      end

      it { expect(response.code).to eq(400) }
      it { expect(json_body[:errors]).to eq({name: ["is already taken"]}) }
    end

    context "non-existing" do
      before(:all) do
        setup_account
        put "/accounts/#{@id+1}", {account: { name: "BooBoo Butt"}}
      end

      it_behaves_like "missing account"
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

      it { expect(response.code).to eq(200) }
      it { expect(body).to be_empty }

      it "should remove the item" do
        get "/accounts/#{@id}"
        expect(response.code).to eq(404)
      end
    end

    context "non-existing" do
      before(:all) do
        delete "/accounts/#{@id+1}"
      end
      
      it_behaves_like "missing account"
    end
  end
end
