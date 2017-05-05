/**
 * AP is the root namespace for all Gateway-provided functionality.
 *
 * @namespace
 */
var AP = AP || {};

AP.prepareRequests = function() {
  var requests = [];
  var numCalls = arguments.length;
  for (var i = 0; i < numCalls; i++) {
    var call = arguments[i];
    if (!call.request) {
      requests.push(request);
    } else {
      requests.push(call.request);
    }
  }
  return JSON.stringify(requests);
}

AP.insertResponses = function(calls, responses) {
  var numCalls = calls.length;
  for (var i = 0; i < numCalls; i++) {
    var call = calls[i];
    call.response = responses[i];
    if (call.response.type == "mongodb") {
      if (call.response.error) {
        throw call.response.error;
      }
      results = call.response.data;
      for (var i in results) {
        AP.Mongo.unnormalizeObjectId(results[i]);
      }
    }
    if (numCalls == 1) {
      response = call.response;
    }
  }
}

console.log = function() {
  data = "";
  space = "";
  for (var i = 0; i < arguments.length; i++) {
    data += space + String(arguments[i]);
    space = " ";
  }
  log(data);
}

prettify = function(o) {
  return "\n" + JSON.stringify(o, null, "   ") + "\n"
}
