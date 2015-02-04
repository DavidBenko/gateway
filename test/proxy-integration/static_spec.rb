require_relative "./spec_helper"

describe "hello-world.json" do
  before(:all) do
    get "/"
  end
  
  it { expect_response(200) }
end