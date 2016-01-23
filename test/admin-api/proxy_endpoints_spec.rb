require_relative "./spec_helper"

shared_examples "empty proxy_endpoints" do
  it { expect_status(200) }
  it { expect_json_types({ proxy_endpoints: :array }) }
  it { expect_json("proxy_endpoints", []) }
end

shared_examples "a missing proxy_endpoint" do
  it { expect_status(404) }
  it { expect_json("error", "No proxy endpoint matches") }
end

shared_examples "a valid proxy_endpoint" do
  it { expect_status(200) }
  it { expect_json_types("proxy_endpoint", {
    id:             :int,
    api_id:         :int,
    environment_id: :int,
    name:           :string,
    description:    :string,
    active:         :bool,
    cors_enabled:   :bool,
    routes:         :array_or_null,
    components:     :array_or_null,
    tests:          :array_or_null,
  }) }
end

describe "proxy_endpoints" do
  before(:all) do
    clear_db!

    @geff = fixtures[:users][:geff]
    @poter = fixtures[:users][:poter]

    # Post Geff's account.
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

    # Post a new Shared Component.
    post "/apis/#{@existent_api_id}/shared_components",
      shared_component: shared_component_for(
        @existent_api_id,
        @existent_remote_endpoint_id,
        @geff[:id],
        :single,
      )
    expect_status(200)
    @existent_shared_component = json_body[:shared_component]

    # Post another Shared Component.
    post "/apis/#{@existent_api_id}/shared_components",
      shared_component: shared_component_for(
        @existent_api_id,
        @existent_remote_endpoint_id,
        @geff[:id],
        :js,
      )
    expect_status(200)
    @second_shared_component_id = json_body[:shared_component][:id]

    # Create a new account.
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

    # Post an environment to @poter's API.
    post "/apis/#{@other_account_api_id}/environments",
      environment: fixtures[:environments][:basic]
    expect_status(200)
    @other_account_env_id = json_body[:environment][:id]
    # Populate the environment_id field of @poter's environment_data.
    poter_env_data = fixtures[:environment_data][:basic].merge({
      environment_id: @other_account_env_id,
    })
    # Set it as the environment_data for @poter's remote endpoint.
    poter_re = fixtures[:remote_endpoints][:basic].merge({
      environment_data: [poter_env_data],
    })

    # Post the new Remote Endpoint to @poter's new API.
    post "/apis/#{@other_account_api_id}/remote_endpoints",
      remote_endpoint: poter_re
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

    # Post a new Proxy Endpoint to @poter's API.
    post "/apis/#{@other_account_api_id}/proxy_endpoints",
      proxy_endpoint: proxy_endpoint_for(
        @other_account_api_id,
        @other_account_env_id,
        @other_account_remote_endpoint_id,
        @other_account_id,
        [{ shared_component_id: @other_account_shared_component_id }],
        :simple,
      ).merge({ name: 'a special other name' })
    expect_status(200)
    @other_account_proxy_endpoint_id = json_body[:proxy_endpoint][:id]

    logout!
  end

  context "logged out" do
    before(:all) do
      logout!
    end

    context "security" do
      before(:all) do
        get "/apis/#{@existent_api_id}/proxy_endpoints"
      end

      it { expect_status(401) }
      it { expect_json("error", "Unauthorized") }
    end

    context "cors preflight" do
      it "should show options for collection" do
        options "/apis/#{@existent_api_id}/proxy_endpoints"
        expect_status 200
        expect(headers[:access_control_allow_methods]).to eq("GET, POST, OPTIONS")
      end

      it "should show options for instance" do
        options "/apis/#{@existent_api_id}/proxy_endpoints/1"
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
          clear_proxy_endpoints!
          get "/apis/#{@existent_api_id}/proxy_endpoints"
        end

        it { expect_status(200) }
        it { expect_json_types({ proxy_endpoints: :array }) }
        it { expect_json("proxy_endpoints", []) }
      end

      context "with data" do
        before(:all) do
          clear_proxy_endpoints!
        end

        it "should return all proxy_endpoints in api" do
          post "/apis/#{@existent_api_id}/proxy_endpoints",
            proxy_endpoint: proxy_endpoint_for(
              @existent_api_id,
              @existent_env_id,
              @existent_remote_endpoint_id,
              @foo_account_id,
              [{ shared_component_id: @existent_shared_component[:id] }],
              :simple,
            )
          expect_status(200)
          expect_count_to_equal(1)

          post "/apis/#{@existent_api_id}/proxy_endpoints",
            proxy_endpoint: proxy_endpoint_for(
              @existent_api_id,
              @existent_env_id,
              @existent_remote_endpoint_id,
              @foo_account_id,
              [{ shared_component_id: @existent_shared_component[:id] }],
              :simple,
            ).merge({ name: 'Second' })
          expect_status(200)
          expect_count_to_equal(2)
        end

        def expect_count_to_equal(num)
          get "/apis/#{@existent_api_id}/proxy_endpoints"
          expect_json_sizes("proxy_endpoints", num)
        end
      end
    end

    describe "create" do
      context "with valid data" do
        before(:all) do
          clear_proxy_endpoints!
          post "/apis/#{@existent_api_id}/proxy_endpoints",
            proxy_endpoint: proxy_endpoint_for(
              @existent_api_id,
              @existent_env_id,
              @existent_remote_endpoint_id,
              @foo_account_id,
              [{ shared_component_id: @existent_shared_component[:id] }],
              :simple,
            )
        end

        it_behaves_like "a valid proxy_endpoint"
      end

      context "with invalid json" do
        before(:all) do
          post "/apis/#{@existent_api_id}/proxy_endpoints", "{"
        end

        it_behaves_like "invalid json"
      end

      context "without a name" do
        before(:all) do
          clear_proxy_endpoints!
          post "/apis/#{@existent_api_id}/proxy_endpoints",
            proxy_endpoint: proxy_endpoint_for(
              @existent_api_id,
              @existent_env_id,
              @existent_remote_endpoint_id,
              @foo_account_id,
              [{ shared_component_id: @existent_shared_component[:id] }],
              :simple,
            ).without(:name)
        end

        it { expect_status(400) }
        it { expect_json("errors", { name: ["must not be blank"] }) }
      end

      context "with invalid routes" do
        before(:all) do
          clear_proxy_endpoints!
          post "/apis/#{@existent_api_id}/proxy_endpoints",
            proxy_endpoint: proxy_endpoint_for(
              @existent_api_id,
              @existent_env_id,
              @existent_remote_endpoint_id,
              @foo_account_id,
              [{ shared_component_id: @existent_shared_component[:id] }],
              :simple,
            ).without(:routes)
        end

        it { expect_status(400) }
        it { expect_json("errors", { routes: ['are invalid'] }) }
      end

      context "with the same name as another proxy_endpoint on account" do
        before(:all) do
          clear_proxy_endpoints!
          ordinary = proxy_endpoint_for(
            @existent_api_id,
            @existent_env_id,
            @existent_remote_endpoint_id,
            @foo_account_id,
            [{ shared_component_id: @existent_shared_component[:id] }],
            :simple,
          )
          post "/apis/#{@existent_api_id}/proxy_endpoints",
            proxy_endpoint: ordinary
          expect_status(200)
          post "/apis/#{@existent_api_id}/proxy_endpoints",
            proxy_endpoint: ordinary
        end

        it { expect_status(400) }
        it { expect_json("errors", { name: ['is already taken'] }) }
      end

      context "with the same name as a proxy_endpoint on another account" do
        before(:all) do
          clear_proxy_endpoints!
          post "/apis/#{@existent_api_id}/proxy_endpoints",
            proxy_endpoint: proxy_endpoint_for(
              @existent_api_id,
              @existent_env_id,
              @existent_remote_endpoint_id,
              @foo_account_id,
              [{ shared_component_id: @existent_shared_component[:id] }],
              :simple,
            ).merge({ name: 'boop' })
        end

        it_behaves_like "a valid proxy_endpoint"
      end
    end

    describe "show" do
      before(:all) do
        clear_proxy_endpoints!
        post "/apis/#{@existent_api_id}/proxy_endpoints",
          proxy_endpoint: proxy_endpoint_for(
            @existent_api_id,
            @existent_env_id,
            @existent_remote_endpoint_id,
            @foo_account_id,
            [{ shared_component_id: @existent_shared_component[:id] }],
            :simple,
          )
        expect_status(200)
        @expect_pe = json_body[:proxy_endpoint]
        @pe_id = @expect_pe[:id]
      end

      context "existing" do
        before(:all) do
          get "/apis/#{@existent_api_id}/proxy_endpoints/#{@pe_id}"
        end

        it_behaves_like "a valid proxy_endpoint"
        it { expect_json("proxy_endpoint", @expect_pe) }
      end

      context "non-existing" do
        before(:all) do
          get "/apis/#{@existent_api_id}/proxy_endpoints/#{@pe_id+100}"
        end

        it_behaves_like "a missing proxy_endpoint"
      end

      context "mismatched account" do
        before(:all) do
          get "/apis/#{@existent_api_id}/proxy_endpoints/#{@other_account_proxy_endpoint_id}"
        end

        it_behaves_like "a missing proxy_endpoint"
      end
    end

    describe "update" do
      before(:all) do
        clear_proxy_endpoints!
        post "/apis/#{@existent_api_id}/proxy_endpoints",
          proxy_endpoint: proxy_endpoint_for(
            @existent_api_id,
            @existent_env_id,
            @existent_remote_endpoint_id,
            @foo_account_id,
            [ fixtures[:components][:simple],
              { shared_component_id: @existent_shared_component[:id] } ],
            :simple,
          )
        expect_status(200)
        @ordinary = json_body[:proxy_endpoint]
      end

      context "with an updated name" do
        before(:all) do
          @new_pe = @ordinary.merge({ name: 'Updated proxy' })
          put "/apis/#{@existent_api_id}/proxy_endpoints/#{@ordinary[:id]}",
            proxy_endpoint: @new_pe
        end

        it_behaves_like "a valid proxy_endpoint"
        it { expect_json("proxy_endpoint", @new_pe) }
      end

      context 'with an updated non-shared component' do
        before(:all) do
          @new_pe = @ordinary.merge({
            components: [
              @ordinary[:components][0].merge({
                data: 'some other code string',
              }),
              @ordinary[:components][1],
            ],
          })

          put "/apis/#{@existent_api_id}/proxy_endpoints/#{@ordinary[:id]}",
            proxy_endpoint: @new_pe
        end

        it_behaves_like "a valid proxy_endpoint"
        it { expect_json("proxy_endpoint", @new_pe) }
      end

      context "with rearranged components" do
        before(:all) do
          @new_pe = @ordinary.merge({
            components: [ @ordinary[:components][1],
                          @ordinary[:components][0] ]
          })
          put "/apis/#{@existent_api_id}/proxy_endpoints/#{@ordinary[:id]}",
            proxy_endpoint: @new_pe
        end

        it_behaves_like "a valid proxy_endpoint"
        it { expect_json("proxy_endpoint", @new_pe) }
      end

      context "with an added non-shared component" do
        before(:all) do
          new_pe = @ordinary.merge({
            components: [ @ordinary[:components][0],
                          @ordinary[:components][1],
                          fixtures[:components][:simple] ]
          })
          put "/apis/#{@existent_api_id}/proxy_endpoints/#{@ordinary[:id]}",
            proxy_endpoint: new_pe
          @expect = new_pe.merge({ components: [
            @ordinary[:components][0],
            @ordinary[:components][1],
            json_body[:proxy_endpoint][:components][2],
          ] })
        end

        it_behaves_like "a valid proxy_endpoint"
        it { expect_json("proxy_endpoint", @expect) }
      end

      context "with an added shared component in the middle" do
        before(:all) do
          new_pe = @ordinary.merge({
            components: [ @ordinary[:components][0],
	                  { shared_component_id: @second_shared_component_id },
                          @ordinary[:components][1] ]
          })

          put "/apis/#{@existent_api_id}/proxy_endpoints/#{@ordinary[:id]}",
            proxy_endpoint: new_pe

          @expect = new_pe.merge({ components: [
            @ordinary[:components][0],
      	    json_body[:proxy_endpoint][:components][1].merge({
              shared_component_id: @second_shared_component_id,
              proxy_endpoint_component_id: @second_shared_component_id,
            }),
            @ordinary[:components][1],
          ] })
        end

        it_behaves_like "a valid proxy_endpoint"
        it { expect_json("proxy_endpoint", @expect) }
      end

      context "with invalid json" do
        before(:all) do
          put "/apis/#{@existent_api_id}/proxy_endpoints/#{@geff[:id]}",
            '{"proxy_endpoint":{"name":"LulzCo'
        end

        it_behaves_like "invalid json"
      end

      context "without a name" do
        before(:all) do
          put "/apis/#{@existent_api_id}/proxy_endpoints/#{@ordinary[:id]}",
            proxy_endpoint: @ordinary.without(:name)
        end

        it { expect_status(400) }
        it { expect_json("errors", { name: ["must not be blank"] }) }
      end

      context "with the same name as another proxy_endpoint on api" do
        before(:all) do
          another = proxy_endpoint_for(
            @existent_api_id,
            @existent_env_id,
            @existent_remote_endpoint_id,
            @foo_account_id,
            [{ shared_component_id: @existent_shared_component[:id] }],
            :simple,
          ).merge({ name: 'beep' })

          post "/apis/#{@existent_api_id}/proxy_endpoints",
            proxy_endpoint: another
          expect_status(200)

          put "/apis/#{@existent_api_id}/proxy_endpoints/#{@ordinary[:id]}",
            proxy_endpoint: @ordinary.merge({ name: another[:name] })
        end

        it { expect_status(400) }
        it { expect_json("errors", { name: ["is already taken"] }) }
      end

      context "with the same name as a proxy_endpoint on another account" do
        before(:all) do
          put "/apis/#{@existent_api_id}/proxy_endpoints/#{@ordinary[:id]}",
            proxy_endpoint: @ordinary.merge({ name: 'a special other name' })
        end

        it_behaves_like "a valid proxy_endpoint"
      end

      context "non-existing" do
        before(:all) do
          put "/apis/#{@existent_api_id}/proxy_endpoints/#{@ordinary[:id]+100}",
            proxy_endpoint: @ordinary
        end

        it_behaves_like "a missing proxy_endpoint"
      end

      context "mismatched account" do
        before(:all) do
          put "/apis/#{@existent_api_id}/proxy_endpoints/#{@other_account_proxy_endpoint_id}", proxy_endpoint: @ordinary
        end

        it_behaves_like "a missing proxy_endpoint"
      end
    end

    describe "delete" do
      before(:all) do
        clear_proxy_endpoints!
        post "/apis/#{@existent_api_id}/proxy_endpoints",
          proxy_endpoint: proxy_endpoint_for(
            @existent_api_id,
            @existent_env_id,
            @existent_remote_endpoint_id,
            @foo_account_id,
            [{ shared_component_id: @existent_shared_component[:id] }],
            :simple,
          )
        expect_status(200)
        @ordinary = json_body[:proxy_endpoint]
      end

      context "existing" do
        before(:all) do
          delete "/apis/#{@existent_api_id}/proxy_endpoints/#{@ordinary[:id]}"
        end

        it { expect_status(200) }
        it { expect(body).to be_empty }

        it "should remove the item" do
          get "/apis/#{@existent_api_id}/proxy_endpoints/#{@ordinary[:id]}"
          expect_status(404)
        end
      end

      context "non-existing" do
        before(:all) do
          delete "/apis/#{@existent_api_id}/proxy_endpoints/#{@ordinary[:id]+1}"
        end

        it_behaves_like "a missing proxy_endpoint"
      end

      context "mismatched account" do
        before(:all) do
          delete "/apis/#{@existent_api_id}/proxy_endpoints/#{@other_account_proxy_endpoint_id}"
        end

        it_behaves_like "a missing proxy_endpoint"
      end
    end
  end
end
