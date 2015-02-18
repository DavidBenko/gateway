# Monkey Patch Airborne to work with non-JSON returning APIs
# This lets all of our integration tests sort of rely on the same
# tech and have the same interface.

module Airborne
  def json_body
    @json_body ||= parse_json
  end

  private
  
  def set_response(res)
    @response = res
    @body = res.body
    @headers = HashWithIndifferentAccess.new(res.headers) unless res.headers.nil?
    @json_body = parse_json rescue nil
  end
  
  def parse_json
    JSON.parse(@body, symbolize_names: true) unless @body.empty?
  end
end