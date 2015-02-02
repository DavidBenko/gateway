require_relative "./spec_helper"

shared_examples "empty users" do
  it { expect_status(200) }
  it { expect_json_types({users: :array}) }
  it { expect_json("users", []) }
end

shared_examples "a missing user" do
  it { expect_status(404) }
  it { expect_json("error", "No user matches") }
end

shared_examples "a valid user" do
  it { expect_status(200) }
  it { expect_json_types("user", {id: :int, name: :string, email: :string}) }
  it { expect_json("user.password", nil)}
  it { expect_json("user.password_confirmation", nil)}
  it { expect_json("user.hashed_password", nil)}
end

describe "users" do
  before(:all) do
    clear_db!

    @geff = fixtures[:users][:geff]
    @poter = fixtures[:users][:poter]
    @brain = fixtures[:users][:brain]

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
    @poter.merge!(id: json_body[:user][:id])
  end

  context "logged out" do
    before(:all) do
      logout!
      get "/users"
    end

    it { expect_status(401) }
    it { expect_json("error", "Unauthorized") }
  end

  context "logged in" do
    before(:all) do
      login @geff[:email], @geff[:password]
    end

    describe "index" do
      context "with data" do
        before(:all) do
          clear_most_users!
        end

        it "should return all users in account" do
          expect_count_to_equal(1)
          post "/users", user: @brain
          expect_count_to_equal(2)
        end

        def expect_count_to_equal(num)
          get "/users"
          expect_json_sizes("users", num)
        end
      end
    end

    describe "create" do
      context "with valid data" do
        before(:all) do
          clear_most_users!
          post "/users", user: @brain
        end

        it_behaves_like "a valid user"
      end

      context "with invalid json" do
        before(:all) do
          post "/users", "{"
        end

        it_behaves_like "invalid json"
      end

      context "without a name" do
        before(:all) do
          clear_most_users!
          post "/users", {user: @brain.without(:name)}
        end

        it { expect_status(400) }
        it { expect_json("errors", {name: ["must not be blank"]}) }
      end

      context "without a email" do
        before(:all) do
          clear_most_users!
          post "/users", {user: @brain.without(:email)}
        end

        it { expect_status(400) }
        it { expect_json("errors", {email: ["must not be blank"]}) }
      end

      context "with a duplicate email" do
        before(:all) do
          clear_most_users!
          post "/users", user: @brain
          expect_status(200)
          post "/users", user: @brain
        end

        it { expect_status(400) }
        it { expect_json("errors", {email: ["is already taken"]}) }
      end

      context "without a password" do
        before(:all) do
          post "/users", {user: @brain.without(:password)}
        end

        it { expect_status(400) }
        it { expect_json("errors", {password: ["must not be blank"]}) }
      end

      context "without a password confirmation" do
        before(:all) do
          post "/users", {user: @brain.without(:password_confirmation)}
        end

        it { expect_status(400) }
        it { expect_json("errors", {password_confirmation: ["must match password"]}) }
      end

      context "without a matching password confirmation" do
        before(:all) do
          brain = @brain.dup
          brain[:password_confirmation] += "x"
          post "/users", user: brain
        end

        it { expect_status(400) }
        it { expect_json("errors", {password_confirmation: ["must match password"]}) }
      end
    end

    describe "show" do
      before(:all) do
        clear_most_users!
        @id = @geff[:id]
      end

      context "existing" do
        before(:all) do
          get "/users/#{@id}"
        end

        it_behaves_like "a valid user"
        it { expect_json("user", {id: @id, name: "Geff", email: "g@ffery.com"}) }
      end

      context "non-existing" do
        before(:all) do
          get "/users/#{@id+100}"
        end

        it_behaves_like "a missing user"
      end

      context "mismatched account" do
        before(:all) do
          get "/users/#{@poter[:id]}"
        end

        it_behaves_like "a missing user"
      end
    end

    describe "update" do
      context "with valid data" do
        before(:all) do
          garf = @geff.without(:password, :password_confirmation)
          garf[:name] = "Garf"
          garf[:email] = "g@rffry.com"
          put "/users/#{@geff[:id]}", { user: garf }
        end

        it_behaves_like "a valid user"
        it { expect_json("user", {id: @geff[:id], name: "Garf", email: "g@rffry.com"}) }
      end

      context "without updating password" do
        before(:all) do
          geff = @geff.without(:password, :password_confirmation)
          put "/users/#{@geff[:id]}", { user: geff }
        end

        it_behaves_like "a valid user"
        it { expect_json("user", {id: @geff[:id], name: "Geff", email: "g@ffery.com"}) }

        it "should not change the password" do
          post "/sessions", {email: "g@ffery.com", password: "password"}
          expect_status(200)
        end
      end

      context "updating password" do
        before(:all) do
          geff = @geff.dup
          geff[:password] = "newpassword"
          geff[:password_confirmation] = "newpassword"
          put "/users/#{@geff[:id]}", { user: geff }
        end

        it_behaves_like "a valid user"
        it { expect_json("user", {id: @geff[:id], name: "Geff", email: "g@ffery.com"}) }

        it "should change the password" do
          post "/sessions", {email: "g@ffery.com", password: "password"}
          expect_status(400)
          post "/sessions", {email: "g@ffery.com", password: "newpassword"}
          expect_status(200)
        end
      end

      context "with invalid json" do
        before(:all) do
          put "/users/#{@geff[:id]}", '{"user":{"name":"LulzCo'
        end

        it_behaves_like "invalid json"
      end

      context "without a name" do
        before(:all) do
          geff = @geff.without(:name)
          put "/users/#{@geff[:id]}", { user: geff }
        end

        it { expect_status(400) }
        it { expect_json("errors", {name: ["must not be blank"]}) }
      end

      context "without a email" do
        before(:all) do
          geff = @geff.without(:email)
          put "/users/#{@geff[:id]}", { user: geff }
        end

        it { expect_status(400) }
        it { expect_json("errors", {email: ["must not be blank"]}) }
      end

      context "with a duplicate email" do
        before(:all) do
          post "/users", user: @brain
          expect_status(200)

          geff = @geff.dup
          geff[:email] = @brain[:email]
          put "/users/#{@geff[:id]}", { user: geff }
        end

        it { expect_status(400) }
        it { expect_json("errors", {email: ["is already taken"]}) }
      end

      context "without a password confirmation" do
        before(:all) do
          geff = @geff.without(:password_confirmation)
          put "/users/#{@geff[:id]}", { user: geff }
        end

        it { expect_status(400) }
        it { expect_json("errors", {password_confirmation: ["must match password"]}) }
      end

      context "without a matching password confirmation" do
        before(:all) do
          geff = @geff.dup
          geff[:password_confirmation] = geff[:password]+"x"
          put "/users/#{@geff[:id]}", { user: geff }
        end

        it { expect_status(400) }
        it { expect_json("errors", {password_confirmation: ["must match password"]}) }
      end

      context "non-existing" do
        before(:all) do
          put "/users/#{@geff[:id]+100}", user: @geff
        end

        it_behaves_like "a missing user"
      end

      context "mismatched account" do
        before(:all) do
          put "/users/#{@poter[:id]}", user: @geff
        end

        it_behaves_like "a missing user"
      end
    end

    describe "delete" do
      before(:all) do
        clear_most_users!
        post "/users", user: @brain
        expect_status(200)
        @id = json_body[:user][:id]
      end

      context "existing" do
        before(:all) do
          delete "/users/#{@id}"
        end

        it { expect_status(200) }
        it { expect(body).to be_empty }

        it "should remove the item" do
          get "/users/#{@id}"
          expect_status(404)
        end
      end

      context "non-existing" do
        before(:all) do
          delete "/users/#{@id+1}"
        end

        it_behaves_like "a missing user"
      end

      context "mismatched account" do
        before(:all) do
          delete "/users/#{@poter[:id]}"
        end

        it_behaves_like "a missing user"
      end
    end
  end

  def clear_most_users!
    clear_users!(@foo_id, cept: @geff[:id])
  end
end
