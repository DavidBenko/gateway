require_relative "./spec_helper"

describe "remote-endpoint-data.json" do
  context "query" do    
    before(:all) do
      get "/remote-test?test=query"
    end
  
    it { expect_status(200) }
    it { expect_json("query", {foo: "bar", baz: "baf"} ) }
  end
end