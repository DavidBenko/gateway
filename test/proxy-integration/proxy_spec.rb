require_relative "./spec_helper"

describe "hello-world.json" do
  before(:all) do
    get "/proxy"
  end
  
  it { expect_status(200) }
  it { expect(body).to eq("Hello, world!\n")}
end