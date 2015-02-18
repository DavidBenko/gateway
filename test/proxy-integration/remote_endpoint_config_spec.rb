require_relative "./spec_helper"

describe "remote-endpoint-data.json" do
  context "query" do    
    before(:all) do
      get "/remote-test?test=query"
    end
  
    it { expect_status(200) }
    it { expect_json("query", {foo: "bar", baz: "baf"} ) }
  end
  
  context "headers" do    
    before(:all) do
      get "/remote-test?test=headers"
    end
  
    it { expect_status(200) }
    it { expect(json_body[:headers][:"X-Pasta"]).to eq("spaghetti") }
    it { expect(json_body[:headers][:"X-Addition"]).to eq("meatballs") }
  end
end