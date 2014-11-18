/**
 * This example shows using environment values in proxy code.
 *
 * $ curl localhost:5000/env
 * How am I? I'm fine.
 * And how are you?
 *
 */
 Acme.Proxy.Environment = function() {
   this.handle = function(request) {
    var response = new AP.HTTP.Response();
    var mood = AP.Environment.get("MOOD") || "not sure";
    var body = "How am I? I'm " + mood + ".\n";
  	response.body = body;
    return response;
  };
}
