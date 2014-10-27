var AP = AP || {};

AP.log = function(line) {
  __ap_log(line);
}

AP.makeRequest = function(request) {
  return this.makeRequests([request])[0];
}

AP.makeRequests = function(requests) {
  var typedRequests = [];
  var numRequests = requests.length;
  for (var i = 0; i < numRequests; i++) {
    var request = requests[i];
    typedRequests.push([request.__ap_type, JSON.stringify(request)]);
  }
  var rawResponse = __ap_makeRequests(JSON.stringify(typedRequests));
  return JSON.parse(rawResponse);
}
