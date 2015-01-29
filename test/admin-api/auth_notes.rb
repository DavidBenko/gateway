
  # it 'should validate types' do
  #   puts response.headers.inspect
  #   expect_json_types({accounts: :array})
  # end
  #
  # it "should do stuff" do
  #   post '/accounts', {:account => {:name => 'John Doe'}}
  #   account_id = json_body[:account][:id]
  #   puts account_id
  #   post "/accounts/#{account_id}/users", {:user => {:name => "Tester", :email => "test@foo.com", :password => "foobar", :password_confirmation => "foobar"}}
  #   user_id = json_body[:user][:id]
  #   puts json_body.inspect
  #   login "test@foo.com", "foobar"
  #   # puts cookies.inspect
  #   # cooks = { "__ap_gateway" => cookies.first[1].gsub("%3D","=")}
  #   # puts cooks.inspect
  #   get "/users/#{user_id}"#, {cookies: cooks}
  #   puts body.inspect
  #   # puts json_body.inspect
  # end

  # it 'should validate values' do
  #   get 'http://example.com/api/v1/simple_get' #json api that returns { "name" : "John Doe" }
  #   expect_json({:name => "John Doe"})
  # end
