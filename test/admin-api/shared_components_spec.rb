require_relative "./spec_helper"

def clear_shared_components!
  get "/apis"
  json_body[:apis].each do |api|
    get "/apis/#{api[:id]}/shared_components"
    json_body[:shared_components].each do |shared|
      delete "/apis/#{api[:id]}/shared_components/#{shared[:id]}"
    end
  end
end

def shared_component_for(api_id, remote_id, acc_id, keyword)
  sh = fixtures[:shared_components][keyword]

  # Insert remote_id in each call as remote_endpoint_id.
  if !sh[:call].nil? then
    sh[:call][:remote_endpoint_id] = remote_id
  elsif !sh[:calls].nil? then
    sh[:calls].each do |call|
      call[:remote_endpoint_id] = remote_id
    end
  end

  # Insert remote_endpoint_id, api_id, and account_id for the shared_component.
  return sh.merge({
    api_id:             api_id,
    remote_endpoint_id: remote_id,
    account_id:         acc_id,
  })
end

shared_examples "empty " do
  it { expect_status(200) }
  it { expect_json_types({shared_components: :array}) }
  it { expect_json("shared_components", []) }
end

shared_examples "a missing shared_component" do
  it { expect_status(404) }
  it { expect_json("error", "No shared component matches") }
end

shared_examples "a valid shared_component" do
  it { expect_status(200) }
  it { expect_json_types("shared_component", {
    id: :int,
    api_id: :int,
    description: :string,
    type: :string,
    conditional: :string,
    conditional_positive: :bool,
    before: :array_or_null,
    after: :array_or_null,
    call: :object_or_null,
    calls: :array_or_null,
  })}
end

describe "shared_components" do
  before(:all) do
    clear_db!

    @geff = fixtures[:users][:geff]
    @poter = fixtures[:users][:poter]

    post "/accounts", account: fixtures[:accounts][:foo]
    expect_status(200)
    @foo_account_id = json_body[:account][:id]
    post "/accounts/#{@foo_account_id}/users", user: @geff
    expect_status(200)
    @geff.merge!(id: json_body[:user][:id])
    login @geff[:email], @geff[:password]

    # Post a new API.
    post "/apis", api: fixtures[:apis][:widgets]
    expect_status(200)
    @existent_api_id = json_body[:api][:id]

    # Post an environment to the new API.
    post "/apis/#{@existent_api_id}/environments",
      environment: fixtures[:environments][:basic]
    expect_status(200)
    @existent_env_id = json_body[:environment][:id]

    # Populate the environment_id field of this environment_data.
    env_data = fixtures[:environment_data][:basic].merge({
      environment_id: @existent_env_id,
    })
    # Set it as the environment_data for the remote endpoint.
    re = fixtures[:remote_endpoints][:basic].merge({
      environment_data: [env_data],
    })
    # Post the new remote endpoint.
    post "/apis/#{@existent_api_id}/remote_endpoints",
      remote_endpoint: re
    expect_status(200)
    @existent_remote_endpoint_id = json_body[:remote_endpoint][:id]
    logout!

    post "/accounts", account: fixtures[:accounts][:bar]
    expect_status(200)
    @other_account_id = json_body[:account][:id]
    post "/accounts/#{@other_account_id}/users", user: @poter
    expect_status(200)
    login @poter[:email], @poter[:password]

    # Post a new API to @poter's account.
    post "/apis", api: fixtures[:apis][:gadgets]
    expect_status(200)
    @other_account_api_id = json_body[:api][:id]

    # Post a new Remote Endpoint to @poter's new API.
    post "/apis/#{@other_account_api_id}/remote_endpoints",
      remote_endpoint: fixtures[:remote_endpoints][:basic].merge({ name: 'boop' })
    expect_status(200)
    @other_account_remote_endpoint_id = json_body[:remote_endpoint][:id]

    # Post a new Shared Component to @poter's API.
    post "/apis/#{@other_account_api_id}/shared_components",
      shared_component: shared_component_for(
        @other_account_api_id,
        @other_account_remote_endpoint_id,
        @other_account_id,
        :single,
      )
    expect_status(200)
    @other_account_shared_component_id = json_body[:shared_component][:id]
    logout!
  end

  context "logged out" do
    before(:all) do
      logout!
    end

    context "security" do
      before(:all) do
        get "/apis/#{@existent_api_id}/shared_components"
      end

      it { expect_status(401) }
      it { expect_json("error", "Unauthorized") }
    end

    context "cors preflight" do
      it "should show options for collection" do
        options "/apis/#{@existent_api_id}/shared_components"
        expect_status 200
        expect(headers[:access_control_allow_methods]).to eq("GET, POST, OPTIONS")
      end

      it "should show options for instance" do
        options "/apis/#{@existent_api_id}/shared_components/1"
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
          clear_shared_components!
          get "/apis/#{@existent_api_id}/shared_components"
        end

        it { expect_status(200) }
        it { expect_json_types({ shared_components: :array }) }
        it { expect_json("shared_components", []) }
      end

      context "with data" do
        before(:all) do
          clear_shared_components!
        end

        it "should return all shared_components in api" do
          post "/apis/#{@existent_api_id}/shared_components",
            shared_component: shared_component_for(
              @existent_api_id,
              @existent_remote_endpoint_id,
              @foo_account_id,
              :single,
            )
          expect_count_to_equal(1)

          post "/apis/#{@existent_api_id}/shared_components",
            shared_component: shared_component_for(
              @existent_api_id,
              @existent_remote_endpoint_id,
              @foo_account_id,
              :multi,
            )
          expect_count_to_equal(2)
        end

        def expect_count_to_equal(num)
          get "/apis/#{@existent_api_id}/shared_components"
          expect_json_sizes("shared_components", num)
        end
      end
    end

    describe "create" do
      context "with valid data" do
        before(:all) do
          clear_shared_components!
          post "/apis/#{@existent_api_id}/shared_components",
            shared_component: shared_component_for(
              @existent_api_id,
              @existent_remote_endpoint_id,
              @foo_account_id,
              :single,
            )
        end

        it_behaves_like "a valid shared_component"
      end

      context "with invalid json" do
        before(:all) do
          post "/apis/#{@existent_api_id}/shared_components", "{"
        end

        it_behaves_like "invalid json"
      end

      context "without a name" do
        before(:all) do
          clear_shared_components!
          post "/apis/#{@existent_api_id}/shared_components",
            shared_component: shared_component_for(
              @existent_api_id,
              @existent_remote_endpoint_id,
              @foo_account_id,
              :single,
            ).without(:name)
        end

        it { expect_status(400) }
        it { expect_json("errors", { name: ["must not be blank"] }) }
      end

      context "with a wrong type" do
        before(:all) do
          clear_shared_components!
          ordinary =
          post "/apis/#{@existent_api_id}/shared_components",
            shared_component: shared_component_for(
              @existent_api_id,
              @existent_remote_endpoint_id,
              @foo_account_id,
              :single,
            ).without(:type)
        end

        it { expect_status(400) }
        it { expect_json("errors",
          { type: ["must be one of 'single', or 'multi', or 'js'"] }) }
      end

      context "with the same name as another shared_component on account" do
        before(:all) do
          clear_shared_components!
          ordinary = shared_component_for(
            @existent_api_id,
            @existent_remote_endpoint_id,
            @foo_account_id,
            :single,
          )
          post "/apis/#{@existent_api_id}/shared_components", {shared_component: ordinary}
          expect_status(200)
          post "/apis/#{@existent_api_id}/shared_components", {shared_component: ordinary}
        end

        it { expect_status(400) }
        it { expect_json("errors", {name: ["is already taken"]}) }
      end

      context "with the same name as a shared_component on another account" do
        before(:all) do
          clear_shared_components!
          ordinary = shared_component_for(
            @existent_api_id,
            @existent_remote_endpoint_id,
            @foo_account_id,
            :single,
          )
          post "/apis/#{@existent_api_id}/shared_components", {shared_component: ordinary}
        end

        it_behaves_like "a valid shared_component"
      end
    end

    describe "show" do
      before(:all) do
        clear_shared_components!
        post "/apis/#{@existent_api_id}/shared_components",
          shared_component: shared_component_for(
            @existent_api_id,
            @existent_remote_endpoint_id,
            @foo_account_id,
            :single,
          )
        expect_status(200)
        @expect_sh = json_body[:shared_component]
        @sh_id = @expect_sh[:id]
      end

      context "existing" do
        before(:all) do
          get "/apis/#{@existent_api_id}/shared_components/#{@sh_id}"
        end

        it_behaves_like "a valid shared_component"
        it { expect_json("shared_component", @expect_sh) }
      end

      context "non-existing" do
        before(:all) do
          get "/apis/#{@existent_api_id}/shared_components/#{@sh_id+100}"
        end

        it_behaves_like "a missing shared_component"
      end

      context "mismatched account" do
        before(:all) do
          get "/apis/#{@existent_api_id}/shared_components/#{@other_account_shared_component_id}"
        end

        it_behaves_like "a missing shared_component"
      end
    end

    describe "update" do
      before(:all) do
        clear_shared_components!
        post "/apis/#{@existent_api_id}/shared_components",
          shared_component: shared_component_for(
            @existent_api_id,
            @existent_remote_endpoint_id,
            @foo_account_id,
            :single,
          )
        expect_status(200)
        @ordinary = json_body[:shared_component]
      end

      context "with valid data" do
        before(:all) do
          also_ordinary = @ordinary.dup
          also_ordinary[:name] = 'Updated single'
          put "/apis/#{@existent_api_id}/shared_components/#{also_ordinary[:id]}",
            { shared_component: also_ordinary }
          expect_status(200)
          @expect_sh = json_body[:shared_component]
            .merge({ name: 'Updated single' })
        end

        it_behaves_like "a valid shared_component"
        it { expect_json("shared_component", @expect_sh) }
      end

      context "with invalid json" do
        before(:all) do
          put "/apis/#{@existent_api_id}/shared_components/#{@geff[:id]}",
            '{"shared_component":{"name":"LulzCo'
        end

        it_behaves_like "invalid json"
      end

      context "without a name" do
        before(:all) do
          put "/apis/#{@existent_api_id}/shared_components/#{@ordinary[:id]}",
            { shared_component: @ordinary.without(:name) }
        end

        it { expect_status(400) }
        it { expect_json("errors", { name: ["must not be blank"] }) }
      end

      context "with the same name as another shared_component on api" do
        before(:all) do
          another = shared_component_for(
            @existent_api_id,
            @existent_remote_endpoint_id,
            @foo_account_id,
            :multi,
          )

          post "/apis/#{@existent_api_id}/shared_components",
            shared_component: another
          expect_status(200)

          put "/apis/#{@existent_api_id}/shared_components/#{@ordinary[:id]}",
            shared_component: @ordinary.merge({ name: another[:name] })
        end

        it { expect_status(400) }
        it { expect_json("errors", {name: ["is already taken"]}) }
      end

      context "with the same name as a shared_component on another account" do
        before(:all) do
          post "/apis/#{@existent_api_id}/shared_components",
            shared_component: shared_component_for(
              @existent_api_id,
              @existent_remote_endpoint_id,
              @foo_account_id,
              :single,
            )
        end

        it_behaves_like "a valid shared_component"
      end

      context "non-existing" do
        before(:all) do
          put "/apis/#{@existent_api_id}/shared_components/#{@ordinary[:id]+100}", shared_component: @ordinary
        end

        it_behaves_like "a missing shared_component"
      end

      context "mismatched account" do
        before(:all) do
          put "/apis/#{@existent_api_id}/shared_components/#{@other_account_shared_component_id}", shared_component: @ordinary
        end

        it_behaves_like "a missing shared_component"
      end
    end

    describe "delete" do
      before(:all) do
        clear_shared_components!
        post "/apis/#{@existent_api_id}/shared_components",
          shared_component: shared_component_for(
            @existent_api_id,
            @existent_remote_endpoint_id,
            @foo_account_id,
            :single,
          )
        expect_status(200)
        @ordinary = json_body[:shared_component]
      end

      context "existing" do
        before(:all) do
          delete "/apis/#{@existent_api_id}/shared_components/#{@ordinary[:id]}"
        end

        it { expect_status(200) }
        it { expect(body).to be_empty }

        it "should remove the item" do
          get "/apis/#{@existent_api_id}/shared_components/#{@ordinary[:id]}"
          expect_status(404)
        end
      end

      context "non-existing" do
        before(:all) do
          delete "/apis/#{@existent_api_id}/shared_components/#{@ordinary[:id]+1}"
        end

        it_behaves_like "a missing shared_component"
      end

      context "mismatched account" do
        before(:all) do
          delete "/apis/#{@existent_api_id}/shared_components/#{@other_account_shared_component_id}"
        end

        it_behaves_like "a missing shared_component"
      end
    end
  end
end
