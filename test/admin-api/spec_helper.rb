require 'airborne'

Airborne.configure do |config|
  config.base_url = "http://admin:admin@localhost:5000/admin"
end

# Add the 'except' method to get a hash less one of its keys.
class Hash
  def except(*ks)
    tap do |h|
      ks.each { |k| h.delete(k) }
    end
  end
end

def api_export_for(api, version, envs, remote_eps, shared_comps, proxy_eps)
  index_map = {
    environments: Hash[envs.collect.with_index(1) { |e, i| [e[:id], i] }],
    proxy_endpoints:
     Hash[proxy_eps.collect.with_index(1) { |pe, i| [pe[:id], i] }],
    remote_endpoints:
     Hash[remote_eps.collect.with_index(1) { |r, i| [r[:id], i] }],
    shared_components:
     Hash[shared_comps.collect.with_index(1) { |sc, i| [sc[:id], i] }],
    environment_data: Hash[
     remote_eps.flat_map do |re|
       re[:environment_data].collect.with_index(1) do |ed, i|
         [ed[:id], i]
       end
     end]
  }

  strip_call = lambda do |call|
    unless call[:before].nil?
      call[:before] = call[:before].collect { |xf| xf.except(:id) }
    end
    unless call[:after].nil?
      call[:after] = call[:after].collect { |xf| xf.except(:id) }
    end

    call.merge(remote_endpoint_index:
                index_map[:remote_endpoints][call[:remote_endpoint_id]])
    .except(:id, :remote_endpoint, :remote_endpoint_id)
  end

  new_reps = remote_eps.collect do |re|
    re.merge(id: 0,
             api_id: 0,
             environment_data:
              re[:environment_data].collect.with_index(1) do |ed, i|
                index_map[:environment_data][ed[:id]] = i
                ed.merge(environment_index:
                          index_map[:environments][ed[:environment_id]],
                         remote_endpoint_id: 0)
                  .except(:id, :environment_id, :links)
              end)
    .except(:id, :api_id)
  end

  new_shared_comps = shared_comps.collect do |sc|
    unless sc[:before].nil?
      sc[:before] = sc[:before].collect { |xf| xf.except(:id) }
    end
    unless sc[:after].nil?
      sc[:after] = sc[:after].collect { |xf| xf.except(:id) }
    end

    sc[:call].nil? || sc[:call] = strip_call.call(sc[:call])
    sc[:calls].nil? || sc[:calls] = sc[:calls].map(&strip_call)

    sc.except(:id,
              :api_id,
              :proxy_endpoint_component_id,
              :proxy_endpoint_component_reference_id)
  end

  new_proxy_eps = proxy_eps.collect do |pe|
    new_comps = pe[:components].collect do |comp|
      unless comp[:before].nil?
        comp[:before] = comp[:before].collect { |xf| xf.except(:id) }
      end
      unless comp[:after].nil?
        comp[:after] = comp[:after].collect { |xf| xf.except(:id) }
      end

      comp[:call].nil? || comp[:call] = strip_call.call(comp[:call])
      comp[:calls].nil? || comp[:calls] = comp[:calls].collect(&strip_call)

      unless comp[:shared_component_id].nil?
        comp[:shared_component_index] =
         index_map[:shared_components][comp[:shared_component_id]]
      end

      comp.except(:id,
                  :proxy_endpoint_component_id,
                  :proxy_endpoint_component_reference_id,
                  :shared_component_id)
    end

    pe.merge(environment_index: index_map[:environments][pe[:environment_id]],
             components: new_comps)
    .except(:id, :api_id, :environment_id)
  end

  api.merge(export_version: version,
            remote_endpoints: new_reps,
            environments: [envs.collect { |e| e.except(:api_id, :id) },
                           fixtures[:environments][:development]].flatten,
            shared_components: new_shared_comps,
            proxy_endpoints: new_proxy_eps,
            base_url: '')
    .except(:id)
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

def proxy_endpoint_for(api_id, env_id, remote_id, acc_id, components, keyword)
  fixtures[:proxy_endpoints][keyword].merge({
    api_id:             api_id,
    environment_id:     env_id,
    remote_endpoint_id: remote_id,
    account_id:         acc_id,
    components:         components,
  })
end

def clear_db!
  get "/accounts"
  json_body[:accounts].each do |account|
    delete "/accounts/#{account[:id]}"
  end
end

def clear_users!(id, options={})
  cept = options[:cept]
  if cept.kind_of? Fixnum
    cept = [cept]
  end

  get "/accounts/#{id}/users"
  json_body[:users].each do |user|
    next if cept && cept.include?(user[:id])
    delete "/accounts/#{id}/users/#{user[:id]}"
  end
end

def clear_apis!
  get "/apis"
  json_body[:apis].each do |api|
    delete "/apis/#{api[:id]}"
  end
end

def clear_proxy_endpoints!
  get "/apis"
  json_body[:apis].each do |api|
    get "/apis/#{api[:id]}/proxy_endpoints"
    json_body[:proxy_endpoints].each do |shared|
      delete "/apis/#{api[:id]}/proxy_endpoints/#{shared[:id]}"
    end
  end
end

def clear_push_channels!
  get "/push_channels"
  json_body[:push_channels].each do |push_channel|
    delete "/push_channels/#{push_channel[:id]}"
  end
end

def clear_push_devices!
  get "/push_channels"
  json_body[:push_channels].each do |push_channel|
    get "/push_channels/#{push_channel[:id]}/push_devices"
    json_body[:push_devices].each do |push_device|
      delete "/push_channels/#{push_channel[:id]}/push_devices/#{push_device[:id]}"
    end
  end
end

def clear_push_messages!
  get "/push_channels"
  json_body[:push_channels].each do |push_channel|
    get "/push_channels/#{push_channel[:id]}/push_devices"
    json_body[:push_devices].each do |push_device|
      get "/push_channels/#{push_channel[:id]}/push_devices/#{push_device[:id]}/push_messages"
      json_body[:push_messages].each do |push_message|
        delete "/push_channels/#{push_channel[:id]}/push_devices/#{push_device[:id]}/push_messages/#{push_message[:id]}"
      end
    end
  end
end

def login(email, pw)
  post "/sessions", {email: email, password: pw}
  expect_status(200)
  cookie = response.cookies.first
  Airborne.configuration.headers = {cookies: { cookie[0] => cookie[1].gsub("%3D","=")}}
end

def logout!
  Airborne.configuration.headers = nil
end

def fixtures
  fixts = {
    accounts: {
      lulz: { name: 'LulzCorp' },
      foo:  { name: 'Foo Corp' },
      bar:  { name: 'Bar Corp' },
    },
    users: {
      geff:  { name: 'Geff',  email: 'g@ffery.com', password: 'password', password_confirmation: 'password', admin: true, confirmed: true },
      brain: { name: 'Brain', email: 'br@in.com',   password: 'password', password_confirmation: 'password', confirmed: true },
      poter: { name: 'Poter', email: 'p@ter.com',   password: 'password', password_confirmation: 'password', confirmed: true },
    },
    environments: {
      basic: {
        name: 'Basic',
        description: 'A basic environment',
        data: {method: 'POST'},
        session_name: 'session',
        session_type: 'client',
        session_auth_key: 'auth-key',
        session_encryption_key: 'encryption-key',
        session_auth_key_rotate: '???',
        session_encryption_key_rotate: '!!!',
        show_javascript_errors: true,
      },
      development: {
        name: 'Development',
        description: '',
        data: nil,
        session_type: 'client',
        session_header: 'X-Session-Id',
        session_name: '',
        session_auth_key: '',
        session_encryption_key: '',
        session_auth_key_rotate: '',
        session_encryption_key_rotate: '',
        show_javascript_errors: false
      }
    },
    environment_data: {
      basic: {
        name: 'Basic',
        description: 'A basic environment.',
        data: {method: 'POST'},
        type: 'http',
        environment_id: 1,
        session_name: 'session',
        session_auth_key: 'auth-key',
        session_encryption_key: 'encryption-key',
        session_auth_key_rotate: '???',
        session_encryption_key_rotate: '!!!',
        show_javascript_errors: true,
      },
      push: {
        name: 'Push',
        type: 'push',
        data: {publish_endpoint: true, subscribe_endpoint: true, unsubscribe_endpoint: true},
      },
    },
    transformations: {
      empty: {
        type: 'js',
        data: '',
      },
      basic: {
        type: 'js',
        data: 'some_basic_javascript();',
      },
      normal: {
        type: 'js',
        data: 'some_normal_javascript();',
      },
    },
    apis: {
      widgets: {
        name: 'Widgets',
        description: 'Lots of widgets here',
        cors_allow_origin: '*',
        cors_allow_headers: 'content-type, accept',
        cors_allow_credentials: true,
        cors_request_headers: '*',
        cors_max_age: 600
      },
      gadgets: {
        name: 'Gadgets',
        description: 'No widgets',
        cors_allow_origin: '*',
        cors_allow_headers: 'content-type, accept',
        cors_allow_credentials: true,
        cors_request_headers: '*',
        cors_max_age: 600
      },
    },
    components: {
      simple: {
        conditional: 'var foo = function () {\n\n};',
        conditional_positive: true,
        type: 'js',
        data: 'code string',
      },
    },
  }

  fixts[:hosts] = {
    basic: {
      name: "localhost",
      hostname: "localhost",
    },
  }

  fixts[:remote_endpoints] = {
    basic: {
      name: 'Basic',
      codename: 'basic',
      description: 'A simple remote endpoint.',
      type: 'http',
      data: {
        method:'GET',
        url:'http://localhost:8080',
        body:'',
        headers: { a: 2 },
        query: { b: 'c' },
      },
    },
    push: {
      name: 'Push',
      codename: 'push',
      description: 'A push remote endpoint.',
      type: 'push',
      data: {
        push_platforms: [
          {
             name: "Test GCM",
             codename: "test-gcm",
             type: "gcm",
             development: false,
             api_key: "AIzaSyCPc5PN7PkKT7BGj-b60XAmEpp5f9N1oNY",
          },
        ],
        publish_endpoint: true, subscribe_endpoint: true, unsubscribe_endpoint: true,
      },
    },
  }

  fixts[:calls] = {
    basic_notrans: {
      endpoint_name_override: '',
      conditional: 'something conditional',
      conditional_positive: true,
      before: [
        fixts[:transformations][:empty],
      ],
      after: [
        fixts[:transformations][:empty],
      ],
    },
    basic: {
      endpoint_name_override: '',
      conditional: 'something conditional',
      conditional_positive: true,
      before: [
        fixts[:transformations][:basic],
        fixts[:transformations][:normal],
      ],
      after: [
        fixts[:transformations][:basic],
        fixts[:transformations][:normal],
      ],
    },
    normal: {
      endpoint_name_override: '',
      conditional: 'something conditional',
      conditional_positive: true,
      before: [
        fixts[:transformations][:basic],
        fixts[:transformations][:normal],
      ],
      after: [
        fixts[:transformations][:basic],
        fixts[:transformations][:normal],
      ],
    },
  }

  fixts[:shared_components] = {
    single: {
      name: 'Ordinary single component',
      description: 'An utterly unremarkable shared_component',
      type: 'single',
      conditional: 'x == 5;',
      conditional_positive: true,
      before: [
        fixts[:transformations][:basic],
        fixts[:transformations][:normal],
      ],
      after: [
        fixts[:transformations][:basic],
        fixts[:transformations][:normal],
      ],
      call: fixts[:calls][:basic],
      data: {},
    },
    single_notrans: {
      name: 'Single notrans',
      description: 'A shared_component without transformations',
      type: 'single',
      conditional: 'x == 5;',
      conditional_positive: true,
      before: [
        fixts[:transformations][:empty],
      ],
      after: [
        fixts[:transformations][:empty],
      ],
      call: fixts[:calls][:basic_notrans],
      data: {},
    },
    multi: {
      name: 'Less Ordinary multi component',
      description: 'A somewhat less ordinary shared_component',
      type: 'multi',
      conditional: 'x == 5;',
      conditional_positive: true,
      before: [
        fixts[:transformations][:basic],
        fixts[:transformations][:normal],
      ],
      after: [
        fixts[:transformations][:basic],
        fixts[:transformations][:normal],
      ],
      calls: [
        fixts[:calls][:basic],
        fixts[:calls][:normal],
      ],
    },
    js: {
      name: 'Javascripty',
      description: 'A JavaScripty shared_component',
      type: 'multi',
      conditional: 'x == 5;',
      conditional_positive: true,
      before: [
        fixts[:transformations][:basic],
        fixts[:transformations][:normal],
      ],
      after: [
        fixts[:transformations][:basic],
        fixts[:transformations][:normal],
      ],
      calls: [
        fixts[:calls][:basic],
        fixts[:calls][:normal],
      ],
    },
  }

  fixts[:proxy_endpoints] = {
    simple: {
      name: 'Simple Proxy',
      description: 'A simple Proxy Endpoint',
      active: true,
      cors_enabled: true,
      routes: [{
        path: '/proxy',
        methods: ['GET'],
      }],
      components: [fixts[:components][:simple]],
      tests: [],
    },
  }

  fixts[:push_channels] = {
    basic: {
      name: "A Push Channel",
      expires: Time.now.to_i + 86400,
    },
  }

  fixts[:push_manual_messages] = {
    basic: {
      payload: { "test-gcm": { default: 'Test'} }
    }
  }

  fixts[:push_devices] = {
    basic: {
      name: "A Push Device",
      type: "test-gcm",
      token: "cqvkjqoUL9A:APA91bEFS9knUbRH_X9_4UzuCdIpUp7iXQUCvmQ8zf1OepQBOEpPKkDNkjslVIqiehRN8WVi2R3hyUmK5FZ14qHMMkPQBq1pEPH2aokuFk4jAIwPEiQSCjcqvkjqoUL9A:APA91bEFS9knUbRH_X9_4UzuCdIpUp7iXQUCvmQ8zf1OepQBOEpPKkDNkjslVIqiehRN8WVi2R3hyUmK5FZ14qHMMkPQBq1pEPH2aokuFk4jAIwPEiQSCj",
      expires: Time.now.to_i + 86400,
      qos: 0,
    },
  }

  fixts[:push_messages] = {
    basic: {
      stamp: Time.now.to_i - 86400,
      data: {
        aps: {
          alert: {
            body: "A test Message",
          },
          'url-args': [],
        },
      },
    },
  }

  fixts[:push_subscribe] = {
    basic: {
      platform: "test-gcm",
	    channel: "test",
      period: 31536000,
      name: "test-gcm",
      token: "cqvkjqoUL9A:APA91bEFS9knUbRH_X9_4UzuCdIpUp7iXQUCvmQ8zf1OepQBOEpPKkDNkjslVIqiehRN8WVi2R3hyUmK5FZ14qHMMkPQBq1pEPH2aokuFk4jAIwPEiQSCj-Ywu9bNVoGrl-ZXMjeqzPw",
    },
  }
  return fixts
end

class Hash
  def without(*keys)
    cpy = self.dup
    keys.each { |key| cpy.delete(key) }
    cpy
  end
end

shared_examples "invalid json" do
  it { expect_status(400) }
  it { expect_json("error", "unexpected end of JSON input") }
end
