require_relative "./spec_helper"

shared_examples "a missing account" do
  it { expect_status(404) }
  it { expect_json("error", "No account matches") }
end

shared_examples "a valid account" do
  it { expect_status(200) }
  it { expect_json_types("account", {id: :int, name: :string}) }
end

describe "accounts" do
  describe "index" do
    context "empty" do
      before(:all) do
        clear_db!
        get "/accounts"
      end

      it { expect_status(200) }
      it { expect_json_types({accounts: :array}) }
      it { expect_json("accounts", []) }
    end

    context "with data" do
      before(:all) do
        clear_db!
      end

      it "should return all accounts" do
        expect_count_to_equal(0)
        post "/accounts", account: fixtures[:accounts][:foo]
        expect_count_to_equal(1)
        post "/accounts", account: fixtures[:accounts][:bar]
        expect_count_to_equal(2)
      end
    end

    def expect_count_to_equal(num)
      get "/accounts"
      expect_json_sizes("accounts", num)
    end
  end

  describe "create" do
    context "with valid data" do
      before(:all) do
        clear_db!
        post "/accounts", account: fixtures[:accounts][:lulz]
      end
      it_behaves_like "a valid account"
      it { expect_json("account.name", "LulzCorp") }
    end

    context "with invalid json" do
      before(:all) do
        post "/accounts", '{"account":{"name":"LulzCo"'
      end

      it_behaves_like "invalid json"
    end

    context "without a name" do
      before(:all) do
        post "/accounts",  {account: { name: ""}}
      end

      it { expect_status(400) }
      it { expect_json("errors", {name: ["must not be blank"]}) }
    end

    context "with a duplicate name" do
      before(:all) do
        clear_db!
        post "/accounts", account: fixtures[:accounts][:lulz]
        expect_status(200)
        post "/accounts",  account: fixtures[:accounts][:lulz]
      end

      it { expect_status(200) }
      it { expect_json("account.name", "LulzCorp") }
    end
  end

  describe "show" do
    before(:all) do
      clear_db!
      post "/accounts", account: fixtures[:accounts][:lulz]
      expect_status(200)
      @id = json_body[:account][:id]
    end

    context "existing" do
      before(:all) do
        get "/accounts/#{@id}"
      end

      it_behaves_like "a valid account"
      it { expect_json("account", {id: @id, name: "LulzCorp"}) }
    end

    context "non-existing" do
      before(:all) do
        get "/accounts/#{@id+1}"
      end

      it_behaves_like "a missing account"
    end
  end

  describe "update" do
    def setup_account
      clear_db!
      post "/accounts", account: fixtures[:accounts][:lulz]
      expect_status(200)
      @id = json_body[:account][:id]
    end

    context "with valid data" do
      before(:all) do
        setup_account
        put "/accounts/#{@id}", account: fixtures[:accounts][:foo]
      end

      it_behaves_like "a valid account"
      it { expect_json("account", {id: @id, name: "Foo Corp"}) }
    end

    context "with invalid json" do
      before(:all) do
        setup_account
        put "/accounts/#{@id}", '{"account":{"name":"LulzCo'
      end

      it_behaves_like "invalid json"
    end

    context "without a name" do
      before(:all) do
        setup_account
        put "/accounts/#{@id}", {account: { name: ""}}
      end

      it { expect_status(400) }
      it { expect_json("errors", {name: ["must not be blank"]}) }
    end

    context "with a duplicate name" do
      before(:all) do
        setup_account
        post "/accounts", {account: { name: "BooBoo Butt"}}
        expect_status(200)
        put "/accounts/#{@id}", {account: { name: "BooBoo Butt"}}
      end

      it { expect_status(200) }
      it { expect_json("account.name", "BooBoo Butt")}
    end

    context "non-existing" do
      before(:all) do
        setup_account
        put "/accounts/#{@id+1}", {account: { name: "BooBoo Butt"}}
      end

      it_behaves_like "a missing account"
    end
  end

  describe "delete" do
    before(:all) do
      clear_db!
      post "/accounts", account: fixtures[:accounts][:lulz]
      expect_status(200)
      @id = json_body[:account][:id]
    end

    context "existing" do
      before(:all) do
        delete "/accounts/#{@id}"
      end

      it { expect_status(200) }
      it { expect(body).to be_empty }

      it "should remove the item" do
        get "/accounts/#{@id}"
        expect_status(404)
      end
    end

    context "non-existing" do
      before(:all) do
        delete "/accounts/#{@id+1}"
      end

      it_behaves_like "a missing account"
    end
  end
end
