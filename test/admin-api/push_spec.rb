require_relative "./spec_helper"

shared_examples "a missing push_channel" do
  it { expect_status(404) }
  it { expect_json("error", "No push channel matches") }
end

shared_examples "a valid push_channel" do
  it { expect_status(200) }
  it { expect_json_types("push_channel", {
    id:                 :int,
    account_id:         :int,
    api_id:             :int,
    remote_endpoint_id: :int,
    name:               :string,
    expires:            :int,
  }) }
end

shared_examples "a missing push_device" do
  it { expect_status(404) }
  it { expect_json("error", "No push device matches") }
end

shared_examples "a valid push_device" do
  it { expect_status(200) }
  it { expect_json_types("push_device", {
    id:                  :int,
    remote_endpoint_id:  :int,
    name:                :string,
    type:                :string,
    token:               :string,
    expires:             :int,
  }) }
end

shared_examples "a missing push_message" do
  it { expect_status(404) }
  it { expect_json("error", "No push message matches") }
end

shared_examples "a valid push_message" do
  it { expect_status(200) }
  it { expect_json_types("push_message", {
    id:             :int,
    push_device_id: :int,
    push_channel_id: :int,
    push_channel_message_id: :int,
    stamp:          :int,
    data:           :object,
  }) }
end

describe "push" do
  before(:all) do
    clear_db!

    @geff = fixtures[:users][:geff]

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

    # Post an host to the new API.
    post "/apis/#{@existent_api_id}/hosts",
      host: fixtures[:hosts][:basic]
    expect_status(200)
    @existent_host_id = json_body[:host][:id]

    # Populate the environment_id field of this environment_data.
    env_data = fixtures[:environment_data][:push].merge({
      environment_id: @existent_env_id,
    })
    # Set it as the environment_data for the push remote endpoint.
    re = fixtures[:remote_endpoints][:push].merge({
      environment_data: [env_data],
    })
    # Post the new remote endpoint.
    post "/apis/#{@existent_api_id}/remote_endpoints",
      remote_endpoint: re
    expect_status(200)
    @existent_remote_endpoint_id = json_body[:remote_endpoint][:id]
  end

  describe "create" do
    context "with valid data" do
      before(:all) do
        clear_push_channels!
        pc = fixtures[:push_channels][:basic].merge({
          account_id: @foo_account_id,
          api_id: @existent_api_id,
          remote_endpoint_id: @existent_remote_endpoint_id,
        })
        post "/push_channels",
          push_channel: pc
      end

      it_behaves_like "a valid push_channel"
    end

    context "with invalid json" do
      before(:all) do
        post "/push_channels", "{"
      end

      it_behaves_like "invalid json"
    end

    context "without a name" do
      before(:all) do
        clear_push_channels!
        pc = fixtures[:push_channels][:basic].merge({
          account_id: @foo_account_id,
          api_id: @existent_api_id,
          remote_endpoint_id: @existent_remote_endpoint_id,
        })
        post "/push_channels",
          push_channel: pc.without(:name)
      end

      it { expect_status(400) }
      it { expect_json("errors", { name: ["must not be blank"] }) }
    end

    context "with the same name as another push_channel on account" do
      before(:all) do
        clear_push_channels!
        pc = fixtures[:push_channels][:basic].merge({
          account_id: @foo_account_id,
          api_id: @existent_api_id,
          remote_endpoint_id: @existent_remote_endpoint_id,
        })
        post "/push_channels",
          push_channel: pc
        expect_status(200)
        post "/push_channels",
          push_channel: pc
      end

      it { expect_status(400) }
      it { expect_json("errors", { name: ['is already taken'] }) }
    end
  end

  describe "show" do
    before(:all) do
      clear_push_channels!
      pc = fixtures[:push_channels][:basic].merge({
        account_id: @foo_account_id,
        api_id: @existent_api_id,
        remote_endpoint_id: @existent_remote_endpoint_id,
      })
      post "/push_channels",
        push_channel: pc
      expect_status(200)
      @expect_pc = json_body[:push_channel]
      @pc_id = @expect_pc[:id]
    end

    context "existing" do
      before(:all) do
        get "/push_channels/#{@pc_id}"
      end

      it_behaves_like "a valid push_channel"
      it { expect_json("push_channel", @expect_pc) }
    end

    context "non-existing" do
      before(:all) do
        get "/push_channels/#{@pc_id+100}"
      end

      it_behaves_like "a missing push_channel"
    end
  end

  describe "update" do
    before(:all) do
      clear_push_channels!
      pc = fixtures[:push_channels][:basic].merge({
        account_id: @foo_account_id,
        api_id: @existent_api_id,
        remote_endpoint_id: @existent_remote_endpoint_id,
      })
      post "/push_channels",
        push_channel: pc
      expect_status(200)
      @ordinary = json_body[:push_channel]
    end

    context "with an updated name" do
      before(:all) do
        @new_pc = @ordinary.merge({ name: 'Updated push channel' })
        put "/push_channels/#{@ordinary[:id]}",
          push_channel: @new_pc
      end

      it_behaves_like "a valid push_channel"
      it { expect_json("push_channel", @new_pc) }
    end

    context "with invalid json" do
      before(:all) do
        put "/push_channels/#{@ordinary[:id]}",
          '{"push_channel":{"name":"LulzCo'
      end

      it_behaves_like "invalid json"
    end

    context "without a name" do
      before(:all) do
        put "/push_channels/#{@ordinary[:id]}",
          push_channel: @ordinary.without(:name)
      end

      it { expect_status(400) }
      it { expect_json("errors", { name: ["must not be blank"] }) }
    end

    context "non-existing" do
      before(:all) do
        put "/push_channels/#{@ordinary[:id]+100}",
          push_channel: @ordinary
      end

      it_behaves_like "a missing push_channel"
    end
  end

  describe "delete" do
    before(:all) do
      clear_push_channels!
      pc = fixtures[:push_channels][:basic].merge({
        account_id: @foo_account_id,
        api_id: @existent_api_id,
        remote_endpoint_id: @existent_remote_endpoint_id,
      })
      post "/push_channels",
        push_channel: pc
      expect_status(200)
      @ordinary = json_body[:push_channel]
    end

    context "existing" do
      before(:all) do
        delete "/push_channels/#{@ordinary[:id]}"
      end

      it { expect_status(200) }
      it { expect(body).to be_empty }

      it "should remove the item" do
        get "/push_channels/#{@ordinary[:id]}"
        expect_status(404)
      end
    end

    context "non-existing" do
      before(:all) do
        delete "/push_channels/#{@ordinary[:id]+1}"
      end

      it_behaves_like "a missing push_channel"
    end
  end

  describe "device" do
    before(:all) do
      clear_push_channels!
      pc = fixtures[:push_channels][:basic].merge({
        account_id: @foo_account_id,
        api_id: @existent_api_id,
        remote_endpoint_id: @existent_remote_endpoint_id,
      })
      post "/push_channels",
        push_channel: pc
      expect_status(200)
      @existent_push_channel_id = json_body[:push_channel][:id]
    end

    describe "create" do
      context "with valid data" do
        before(:all) do
          clear_push_devices!
          pd = fixtures[:push_devices][:basic].merge({
            push_channel_id: @existent_push_channel_id,
          })
          post "/push_channels/#{@existent_push_channel_id}/push_devices",
            push_device: pd
        end

        it_behaves_like "a valid push_device"
      end

      context "with invalid json" do
        before(:all) do
          post "/push_channels/#{@existent_push_channel_id}/push_devices", "{"
        end

        it_behaves_like "invalid json"
      end

      context "without a name" do
        before(:all) do
          clear_push_devices!
          pd = fixtures[:push_devices][:basic].merge({
            push_channel_id: @existent_push_channel_id,
          })
          post "/push_channels/#{@existent_push_channel_id}/push_devices",
            push_device: pd.without(:name)
        end

        it { expect_status(400) }
        it { expect_json("errors", { name: ["must not be blank"] }) }
      end

      # TODO Jeff: remove transparent handling of token lookup on device insert.
      # context "with the same token as another push_device on account" do
      #   before(:all) do
      #     clear_push_devices!
      #     pd = fixtures[:push_devices][:basic].merge({
      #       push_channel_id: @existent_push_channel_id,
      #     })
      #     post "/push_channels/#{@existent_push_channel_id}/push_devices",
      #       push_device: pd
      #     expect_status(200)
      #     post "/push_channels/#{@existent_push_channel_id}/push_devices",
      #       push_device: pd
      #   end
      #
      #   it { expect_status(400) }
      #   it { expect_json("errors", { token: ['is already taken'] }) }
      # end
    end

    describe "show" do
      before(:all) do
        clear_push_devices!
        pd = fixtures[:push_devices][:basic].merge({
          push_channel_id: @existent_push_channel_id,
        })
        post "/push_channels/#{@existent_push_channel_id}/push_devices",
          push_device: pd
        expect_status(200)
        @expect_pd = json_body[:push_device]
        @pd_id = @expect_pd[:id]
      end

      context "existing" do
        before(:all) do
          get "/push_channels/#{@existent_push_channel_id}/push_devices/#{@pd_id}"
        end
        it_behaves_like "a valid push_device"

        it { expect_json("push_device", @expect_pd) }
      end

      context "non-existing" do
        before(:all) do
          get "/push_channels/#{@existent_push_channel_id}/push_devices/#{@pd_id+100}"
        end

        it_behaves_like "a missing push_device"
      end
    end

    describe "update" do
      before(:all) do
        clear_push_devices!
        pd = fixtures[:push_devices][:basic].merge({
          push_channel_id: @existent_push_channel_id,
        })
        post "/push_channels/#{@existent_push_channel_id}/push_devices",
          push_device: pd
        expect_status(200)
        @ordinary = json_body[:push_device]
      end

      context "with an updated name" do
        before(:all) do
          @new_pd = @ordinary.merge({ name: 'Updated push device' })
          put "/push_channels/#{@existent_push_channel_id}/push_devices/#{@ordinary[:id]}",
            push_device: @new_pd
        end

        it_behaves_like "a valid push_device"
        it { expect_json("push_device", @new_pd) }
      end

      context "with invalid json" do
        before(:all) do
          put "/push_channels/#{@existent_push_channel_id}/push_devices/#{@ordinary[:id]}",
            '{"push_device":{"name":"LulzCo'
        end

        it_behaves_like "invalid json"
      end

      context "without a name" do
        before(:all) do
          put "/push_channels/#{@existent_push_channel_id}/push_devices/#{@ordinary[:id]}",
            push_device: @ordinary.without(:name)
        end

        it { expect_status(400) }
        it { expect_json("errors", { name: ["must not be blank"] }) }
      end

      context "non-existing" do
        before(:all) do
          put "/push_channels/#{@existent_push_channel_id}/push_devices/#{@ordinary[:id]+100}",
            push_device: @ordinary
        end

        it_behaves_like "a missing push_device"
      end
    end

    describe "delete" do
      before(:all) do
        clear_push_devices!
        pd = fixtures[:push_devices][:basic].merge({
          push_channel_id: @existent_push_channel_id,
        })
        post "/push_channels/#{@existent_push_channel_id}/push_devices",
          push_device: pd
        expect_status(200)
        @ordinary = json_body[:push_device]
      end

      context "existing" do
        before(:all) do
          delete "/push_channels/#{@existent_push_channel_id}/push_devices/#{@ordinary[:id]}"
        end

        it { expect_status(200) }
        it { expect(body).to be_empty }

        it "should remove the item" do
          get "/push_channels/#{@existent_push_channel_id}/push_devices/#{@ordinary[:id]}"
          expect_status(404)
        end
      end

      context "non-existing" do
        before(:all) do
          delete "/push_channels/#{@existent_push_channel_id}/push_devices/#{@ordinary[:id]+1}"
        end

        it_behaves_like "a missing push_device"
      end
    end

    describe "message" do
      before(:all) do
        clear_push_devices!
        pd = fixtures[:push_devices][:basic].merge({
          push_channel_id: @existent_push_channel_id,
        })
        post "/push_channels/#{@existent_push_channel_id}/push_devices",
          push_device: pd
        expect_status(200)
        @existent_push_device_id = json_body[:push_device][:id]
      end

      describe "create" do
        context "with valid data" do
          before(:all) do
            clear_push_messages!
          end

          it "should send a manual message" do
            pm = fixtures[:push_manual_messages][:basic]
            post "/push_channels/#{@existent_push_channel_id}/push_manual_messages",
              push_manual_message: pm.merge({environment: "Push"})
            expect_status(200)
          end
        end

        context "with invalid json" do
          before(:all) do
            post "/push_channels/#{@existent_push_channel_id}/push_manual_messages", "{"
          end

          it_behaves_like "invalid json"
        end
      end

      describe "show" do
        before(:all) do
          clear_push_messages!
          pm = fixtures[:push_manual_messages][:basic]
          post "/push_channels/#{@existent_push_channel_id}/push_manual_messages",
            push_manual_message: pm.merge({environment: "Push"})
          expect_status(200)
          get "/push_channels/#{@existent_push_channel_id}/push_devices/#{@existent_push_device_id}/push_messages"
          @expect_pm = json_body[:push_messages][0]
          @pm_id = @expect_pm[:id]
        end

        context "existing" do
          before(:all) do
            get "/push_channels/#{@existent_push_channel_id}/push_devices/#{@existent_push_device_id}/push_messages/#{@pm_id}"
          end

          it_behaves_like "a valid push_message"
          it { expect_json("push_message", @expect_pm) }
        end

        context "non-existing" do
          before(:all) do
            get "/push_channels/#{@existent_push_channel_id}/push_devices/#{@existent_push_device_id}/push_messages/#{@pm_id+100}"
          end

          it_behaves_like "a missing push_message"
        end
      end

      describe "delete" do
        before(:all) do
          clear_push_messages!
          pm = fixtures[:push_manual_messages][:basic]
          post "/push_channels/#{@existent_push_channel_id}/push_manual_messages",
            push_manual_message: pm.merge({environment: "Push"})
          expect_status(200)
          get "/push_channels/#{@existent_push_channel_id}/push_devices/#{@existent_push_device_id}/push_messages"
          @ordinary = json_body[:push_messages][0]
        end

        context "existing" do
          before(:all) do
            delete "/push_channels/#{@existent_push_channel_id}/push_devices/#{@existent_push_device_id}/push_messages/#{@ordinary[:id]}"
          end

          it { expect_status(200) }
          it { expect(body).to be_empty }

          it "should remove the item" do
            get "/push_channels/#{@existent_push_channel_id}/push_devices/#{@existent_push_device_id}/push_messages/#{@ordinary[:id]}"
            expect_status(404)
          end
        end

        context "non-existing" do
          before(:all) do
            delete "/push_channels/#{@existent_push_channel_id}/push_devices/#{@existent_push_device_id}/push_messages/#{@ordinary[:id]+1}"
          end

          it_behaves_like "a missing push_message"
        end
      end
    end
  end

  describe "public" do
    before(:all) do
      @tmp_url = Airborne.configuration.base_url
      Airborne.configuration.base_url = "http://localhost:5000"
    end
    after(:all) do
      Airborne.configuration.base_url = @tmp_url
    end

    describe "subscribe" do
      context "valid" do
        before(:all) do
          put "/push/push/subscribe",
            fixtures[:push_subscribe][:basic]
        end

        it { expect_status(200) }
      end
    end

    describe "unsubscribe" do
      before(:all) do
        put "/push/push/subscribe",
          fixtures[:push_subscribe][:basic]
        expect_status(200)
      end

      context "valid" do
        before(:all) do
          put "/push/push/unsubscribe",
            fixtures[:push_subscribe][:basic]
        end

        it { expect_status(200) }
      end
    end
  end
end
