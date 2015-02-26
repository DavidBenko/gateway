require_relative "./spec_helper"

shared_examples "empty apis" do
  it { expect_status(200) }
  it { expect_json_types({apis: :array}) }
  it { expect_json("apis", []) }
end

shared_examples "a missing api" do
  it { expect_status(404) }
  it { expect_json("error", "No api matches") }
end

shared_examples "a valid api" do
  it { expect_status(200) }
  it { expect_json_types("api", {id: :int, 
                                 name: :string,
                                 description: :string, 
                                 cors_allow_origin: :string,
                                 cors_allow_headers: :string,
                                 cors_allow_credentials: :boolean,
                                 cors_request_headers: :string,
                                 cors_max_age: :int
                                 })}
end

describe "apis" do
  before(:all) do
    clear_db!

    @geff = fixtures[:users][:geff]
    @poter = fixtures[:users][:poter]

    post "/accounts", account: fixtures[:accounts][:foo]
    expect_status(200)
    @foo_id = json_body[:account][:id]
    post "/accounts/#{@foo_id}/users", user: @geff
    expect_status(200)
    @geff.merge!(id: json_body[:user][:id])

    post "/accounts", account: fixtures[:accounts][:bar]
    expect_status(200)
    post "/accounts/#{json_body[:account][:id]}/users", user: @poter
    expect_status(200)
    login @poter[:email], @poter[:password]
    post "/apis", api: fixtures[:apis][:gadgets]
    expect_status(200)
    @other_account_api_id = json_body[:api][:id]
    logout!
  end

  context "logged out" do
    before(:all) do
      logout!
    end

    context "security" do
      before(:all) do
        get "/apis"
      end

      it { expect_status(401) }
      it { expect_json("error", "Unauthorized") }
    end

    context "cors preflight" do
      it "should show options for collection" do
        options "/apis"
        expect_status 200
        expect(headers[:access_control_allow_methods]).to eq("GET, POST, OPTIONS")
      end

      it "should show options for instance" do
        options "/apis/1"
        expect_status 200
        expect(headers[:access_control_allow_methods]).to eq("GET, PUT, DELETE, OPTIONS")
      end
    end
  end

  context "logged in" do
    before(:all) do
      login @geff[:email], @geff[:password]
    end

    describe "index" do
      context "empty" do
        before(:all) do
          clear_apis!
          get "/apis"
        end

        it { expect_status(200) }
        it { expect_json_types({apis: :array}) }
        it { expect_json("apis", []) }
      end

      context "with data" do
        before(:all) do
          clear_apis!
        end

        it "should return all apis in account" do
          post "/apis", api: fixtures[:apis][:widgets]
          expect_count_to_equal(1)
          post "/apis", api: fixtures[:apis][:gadgets]
          expect_count_to_equal(2)
        end

        def expect_count_to_equal(num)
          get "/apis"
          expect_json_sizes("apis", num)
        end
      end
    end

    describe "create" do
      context "with valid data" do
        before(:all) do
          clear_apis!
          post "/apis", api: fixtures[:apis][:widgets]
        end

        it_behaves_like "a valid api"
      end

      context "with invalid json" do
        before(:all) do
          post "/apis", "{"
        end

        it_behaves_like "invalid json"
      end

      context "without a name" do
        before(:all) do
          clear_apis!
          post "/apis", {api: fixtures[:apis][:widgets].without(:name)}
        end

        it { expect_status(400) }
        it { expect_json("errors", {name: ["must not be blank"]}) }
      end

      context "with the same name as another api on account" do
        before(:all) do
          clear_apis!
          post "/apis", {api: fixtures[:apis][:gadgets]}
          expect_status(200)
          post "/apis", {api: fixtures[:apis][:gadgets]}
        end

        it { expect_status(400) }
        it { expect_json("errors", {name: ["is already taken"]}) }
      end

      context "with the same name as an api on another account" do
        before(:all) do
          clear_apis!
          post "/apis", {api: fixtures[:apis][:gadgets]}
        end

        it_behaves_like "a valid api"
      end
    end

    describe "show" do
      before(:all) do
        clear_apis!
        post "/apis", api: fixtures[:apis][:widgets]
        expect_status 200
        @id = json_body[:api][:id]
      end

      context "existing" do
        before(:all) do
          get "/apis/#{@id}"
        end

        it_behaves_like "a valid api"
        it { expect_json("api", {id: @id, name: "Widgets",
                                description: "Lots of widgets here",
                                cors_allow_origin: "*"}) }
      end

      context "non-existing" do
        before(:all) do
          get "/apis/#{@id+100}"
        end

        it_behaves_like "a missing api"
      end

      context "mismatched account" do
        before(:all) do
          get "/apis/#{@other_account_api_id}"
        end

        it_behaves_like "a missing api"
      end
    end

    describe "update" do
      before(:all) do
        clear_apis!
        @widgets = fixtures[:apis][:widgets]
        post "/apis", api: @widgets
        expect_status 200
        @widgets.merge!({id: json_body[:api][:id]})
      end

      context "with valid data" do
        before(:all) do
          wadgets = @widgets.dup
          wadgets[:name] = "Wadgets"
          put "/apis/#{wadgets[:id]}", { api: wadgets }
        end

        it_behaves_like "a valid api"
        it { expect_json("api", {id: @widgets[:id], name: "Wadgets",
                                description: "Lots of widgets here",
                                cors_allow_origin: "*"}) }
      end

      context "with invalid json" do
        before(:all) do
          put "/apis/#{@geff[:id]}", '{"api":{"name":"LulzCo'
        end

        it_behaves_like "invalid json"
      end

      context "without a name" do
        before(:all) do
          put "/apis/#{@widgets[:id]}", { api: @widgets.without(:name) }
        end

        it { expect_status(400) }
        it { expect_json("errors", {name: ["must not be blank"]}) }
      end

      context "with the same name as another api on account" do
        before(:all) do
          goodgets = fixtures[:apis][:gadgets]
          goodgets[:name] = "Goodgets"
          post "/apis", {api: goodgets}
          expect_status(200)
          woodgets = @widgets.dup
          woodgets[:name] = goodgets[:name]
          put "/apis/#{@widgets[:id]}", api: woodgets
        end

        it { expect_status(400) }
        it { expect_json("errors", {name: ["is already taken"]}) }
      end

      context "with the same name as an api on another account" do
        before(:all) do
          post "/apis", {api: fixtures[:apis][:gadgets]}
        end

        it_behaves_like "a valid api"
      end

      context "non-existing" do
        before(:all) do
          put "/apis/#{@widgets[:id]+100}", api: @widgets
        end

        it_behaves_like "a missing api"
      end

      context "mismatched account" do
        before(:all) do
          put "/apis/#{@other_account_api_id}", api: @widgets
        end

        it_behaves_like "a missing api"
      end
    end

    describe "delete" do
      before(:all) do
        clear_apis!
        @widgets = fixtures[:apis][:widgets]
        post "/apis", api: @widgets
        expect_status 200
        @widgets.merge!({id: json_body[:api][:id]})
      end

      context "existing" do
        before(:all) do
          delete "/apis/#{@widgets[:id]}"
        end

        it { expect_status(200) }
        it { expect(body).to be_empty }

        it "should remove the item" do
          get "/apis/#{@widgets[:id]}"
          expect_status(404)
        end
      end

      context "non-existing" do
        before(:all) do
          delete "/apis/#{@widgets[:id]+1}"
        end

        it_behaves_like "a missing api"
      end

      context "mismatched account" do
        before(:all) do
          delete "/apis/#{@other_account_api_id}"
        end

        it_behaves_like "a missing api"
      end
    end
  end
end
