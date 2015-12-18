require 'airborne'

Airborne.configure do |config|
  config.base_url = "http://localhost:5000/admin"
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
        session_auth_key: 'auth-key',
        session_encryption_key: 'encryption-key',
        session_auth_key_rotate: '???',
        session_encryption_key_rotate: '!!!',
        show_javascript_errors: true,
      },
    },
    environment_data: {
      basic: {
        name: 'Basic',
        description: 'A basic environment.',
        data: {method: 'POST'},
        environment_id: 1,
        session_name: 'session',
        session_auth_key: 'auth-key',
        session_encryption_key: 'encryption-key',
        session_auth_key_rotate: '???',
        session_encryption_key_rotate: '!!!',
        show_javascript_errors: true,
      },
    },
    transformations: {
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
  }

  fixts[:calls] = {
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
