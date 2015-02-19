require_relative "./spec_helper"

describe "hello-world.json" do
  before(:all) do
    get "/"
  end
  
  it { expect_status(200) }
  it { expect(body).to eq("Hello, world!\n")}
end

describe "basic-library.json" do
  before(:all) do
    get "/library"
  end
  
  it { expect_status(200) }
  it { expect_json({library: true}) }
end


describe "env.json" do
  before(:all) do
    get "/env"
  end
  
  it { expect_status(200) }
  it { expect_json({env: { env: "dev"}}) }
end