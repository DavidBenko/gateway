require_relative "./spec_helper"

shared_examples "empty account users" do
  it { expect_status(200) }
  it { expect_json_types({users: :array}) }
  it { expect_json("users", []) }
end

shared_examples "a missing account user" do
  it { expect_status(404) }
  it { expect_json("error", "No user matches") }
end

shared_examples "a valid account user" do
  it { expect_status(200) }
  it { expect_json_types("user", {id: :int, name: :string, email: :string}) }
  it { expect_json("user.password", nil)}
  it { expect_json("user.password_confirmation", nil)}
  it { expect_json("user.hashed_password", nil)}
end

describe "users via accounts" do
  before(:all) do
    clear_db!

    post "/accounts", fixtures[:accounts][:foo]
    @foo_id = json_body[:account][:id]
    post "/accounts", fixtures[:accounts][:bar]
    @bar_id = json_body[:account][:id]
    @nonexistent_id = @bar_id + 1
  end

  describe "index" do
    context "empty" do
      before(:all) do
        get "/accounts/#{@foo_id}/users"
      end

      it_behaves_like "empty account users"
    end

    context "with data" do
      before(:all) do
        clear_users!(@foo_id)
      end

      it "should return all users in account" do
        expect_count_to_equal(0)
        post "/accounts/#{@foo_id}/users", fixtures[:users][:geff]
        expect_count_to_equal(1)
        post "/accounts/#{@foo_id}/users", fixtures[:users][:brain]
        expect_count_to_equal(2)
      end

      def expect_count_to_equal(num)
        get "/accounts/#{@foo_id}/users"
        expect_json_sizes("users", num)
      end
    end

    context "non-existent account" do
      before(:all) do
        get "/accounts/#{@nonexistent_id}/users"
      end

      # This is arguably inaccurate data, but I"m not sure
      # returning 404 is worth adding extra DB calls.
      it_behaves_like "empty account users"
    end
  end

  describe "create" do
    context "with valid data" do
      before(:all) do
        clear_users!(@foo_id)
        post "/accounts/#{@foo_id}/users", fixtures[:users][:geff]
      end

      it_behaves_like "a valid account user"
    end

    context "with invalid json" do
      before(:all) do
        post "/accounts/#{@foo_id}/users", "{"
      end

      it_behaves_like "invalid json"
    end

    context "without a name" do
      before(:all) do
        clear_users!(@foo_id)
        geff = fixtures[:users][:geff][:user]
        post "/accounts/#{@foo_id}/users", {user: geff.without(:name)}
      end

      it { expect_status(400) }
      it { expect_json("errors", {name: ["must not be blank"]}) }
    end

    context "without a email" do
      before(:all) do
        clear_users!(@foo_id)
        geff = fixtures[:users][:geff][:user]
        post "/accounts/#{@foo_id}/users", {user: geff.without(:email)}
      end

      it { expect_status(400) }
      it { expect_json("errors", {email: ["must not be blank"]}) }
    end

    context "with a duplicate email" do
      before(:all) do
        clear_users!(@foo_id)
        post "/accounts/#{@foo_id}/users", fixtures[:users][:geff]
        expect_status(200)
        brain = fixtures[:users][:brain]
        brain[:user][:email] = fixtures[:users][:geff][:user][:email]
        post "/accounts/#{@foo_id}/users",  brain
      end

      it { expect_status(400) }
      it { expect_json("errors", {email: ["is already taken"]}) }
    end

    context "without a password" do
      before(:all) do
        geff = fixtures[:users][:geff][:user]
        post "/accounts/#{@foo_id}/users", {user: geff.without(:password)}
      end

      it { expect_status(400) }
      it { expect_json("errors", {password: ["must not be blank"]}) }
    end

    context "without a password confirmation" do
      before(:all) do
        geff = fixtures[:users][:geff][:user]
        post "/accounts/#{@foo_id}/users", {user: geff.without(:password_confirmation)}
      end

      it { expect_status(400) }
      it { expect_json("errors", {password_confirmation: ["must match password"]}) }
    end

    context "without a matching password confirmation" do
      before(:all) do
        geff = fixtures[:users][:geff]
        geff[:user][:password_confirmation] = geff[:user][:password_confirmation]+"x"
        post "/accounts/#{@foo_id}/users", geff
      end

      it { expect_status(400) }
      it { expect_json("errors", {password_confirmation: ["must match password"]}) }
    end
  end

  describe "show" do
    before(:all) do
      clear_users!(@foo_id)
      post "/accounts/#{@foo_id}/users", fixtures[:users][:geff]
      expect_status(200)
      @id = json_body[:user][:id]
    end

    context "existing" do
      before(:all) do
        get "/accounts/#{@foo_id}/users/#{@id}"
      end

      it_behaves_like "a valid account user"
      it { expect_json("user", {id: @id, name: "Geff", email: "g@ffery.com"}) }
    end

    context "non-existing" do
      before(:all) do
        get "/accounts/#{@foo_id}/users/#{@id+1}"
      end

      it_behaves_like "a missing account user"
    end

    context "mismatched account" do
      before(:all) do
        post "/accounts/#{@bar_id}/users", fixtures[:users][:brain]
        expect_status(200)
        @id2 = json_body[:user][:id]
        get "/accounts/#{@foo_id}/users/#{@id}"
        expect_status(200)
        get "/accounts/#{@foo_id}/users/#{@id2}"
      end

      it_behaves_like "a missing account user"
    end
  end

  describe "update" do
    def setup_user
      clear_users!(@foo_id)
      clear_users!(@bar_id)
      post "/accounts/#{@foo_id}/users", fixtures[:users][:geff]
      expect_status(200)
      @id = json_body[:user][:id]
    end

    context "with valid data" do
      before(:all) do
        setup_user
        garf = fixtures[:users][:geff].without(:password, :password_confirmation)
        garf[:name] = "Garf"
        garf[:email] = "g@rffry.com"
        put "/accounts/#{@foo_id}/users/#{@id}", { user: garf }
      end

      it_behaves_like "a valid account user"
      it { expect_json("user", {id: @id, name: "Garf", email: "g@rffry.com"}) }
    end

    context "without updating password" do
      before(:all) do
        setup_user

        geff = fixtures[:users][:geff][:user].without(:password, :password_confirmation)
        put "/accounts/#{@foo_id}/users/#{@id}", { user: geff }
      end

      it_behaves_like "a valid account user"
      it { expect_json("user", {id: @id, name: "Geff", email: "g@ffery.com"}) }

      it "should not change the password" do
        post "/sessions", {email: "g@ffery.com", password: "password"}
        expect_status(200)
      end
    end

    context "updating password" do
      before(:all) do
        setup_user

        geff = fixtures[:users][:geff][:user]
        geff[:password] = "newpassword"
        geff[:password_confirmation] = "newpassword"
        put "/accounts/#{@foo_id}/users/#{@id}", { user: geff }
      end

      it_behaves_like "a valid account user"
      it { expect_json("user", {id: @id, name: "Geff", email: "g@ffery.com"}) }

      it "should change the password" do
        post "/sessions", {email: "g@ffery.com", password: "password"}
        expect_status(400)
        post "/sessions", {email: "g@ffery.com", password: "newpassword"}
        expect_status(200)
      end
    end

    context "with invalid json" do
      before(:all) do
        setup_user

        put "/accounts/#{@foo_id}/users/#{@id}", '{"user":{"name":"LulzCo'
      end

      it_behaves_like "invalid json"
    end

    context "without a name" do
      before(:all) do
        setup_user

        geff = fixtures[:users][:geff][:user].without(:name)
        put "/accounts/#{@foo_id}/users/#{@id}", { user: geff }
      end

      it { expect_status(400) }
      it { expect_json("errors", {name: ["must not be blank"]}) }
    end

    context "without a email" do
      before(:all) do
        setup_user

        geff = fixtures[:users][:geff][:user].without(:email)
        put "/accounts/#{@foo_id}/users/#{@id}", { user: geff }
      end

      it { expect_status(400) }
      it { expect_json("errors", {email: ["must not be blank"]}) }
    end

    context "with a duplicate email" do
      before(:all) do
        setup_user
        post "/accounts/#{@foo_id}/users", fixtures[:users][:brain]
        expect_status(200)

        geff = fixtures[:users][:geff][:user]
        geff[:email] = fixtures[:users][:brain][:user][:email]
        put "/accounts/#{@foo_id}/users/#{@id}", { user: geff }
      end

      it { expect_status(400) }
      it { expect_json("errors", {email: ["is already taken"]}) }
    end

    context "without a password confirmation" do
      before(:all) do
        setup_user

        geff = fixtures[:users][:geff][:user].without(:password_confirmation)
        put "/accounts/#{@foo_id}/users/#{@id}", { user: geff }
      end

      it { expect_status(400) }
      it { expect_json("errors", {password_confirmation: ["must match password"]}) }
    end

    context "without a matching password confirmation" do
      before(:all) do
        setup_user

        geff = fixtures[:users][:geff][:user]
        geff[:password_confirmation] = geff[:password]+"x"
        put "/accounts/#{@foo_id}/users/#{@id}", { user: geff }
      end

      it { expect_status(400) }
      it { expect_json("errors", {password_confirmation: ["must match password"]}) }
    end

    context "non-existing" do
      before(:all) do
        setup_user

        put "/accounts/#{@foo_id}/users/#{@id+1}", fixtures[:users][:geff]
      end

      it_behaves_like "a missing account user"
    end

    context "mismatched account" do
      before(:all) do
        setup_user

        put "/accounts/#{@bar_id}/users/#{@id}", fixtures[:users][:geff]
      end

      it_behaves_like "a missing account user"
    end
  end

  describe "delete" do
    before(:all) do
      clear_users!(@foo_id)
      clear_users!(@bar_id)
      post "/accounts/#{@foo_id}/users", fixtures[:users][:geff]
      expect_status(200)
      @id = json_body[:user][:id]
    end

    context "existing" do
      before(:all) do
        delete "/accounts/#{@foo_id}/users/#{@id}"
      end

      it { expect_status(200) }
      it { expect(body).to be_empty }

      it "should remove the item" do
        get "/accounts/#{@foo_id}/users/#{@id}"
        expect_status(404)
      end
    end

    context "non-existing" do
      before(:all) do
        delete "/accounts/#{@foo_id}/users/#{@id+1}"
      end

      it_behaves_like "a missing account user"
    end

    context "mismatched account" do
      before(:all) do
        delete "/accounts/#{@bar_id}/users/#{@id+1}"
      end

      it_behaves_like "a missing account user"
    end
  end
end
