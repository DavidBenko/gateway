/**
 * Counts how many times this endpoint has been called this session.
 *
 * $ curl -b cookies.txt -c cookies.txt localhost:5000/counter
 * You have called this endpoint 1 times.
 * $ curl -b cookies.txt -c cookies.txt localhost:5000/counter
 * You have called this endpoint 2 times.
 * $ curl -b cookies.txt -c cookies.txt localhost:5000/counter
 * You have called this endpoint 3 times.
 * $ curl -b cookies.txt -c cookies.txt localhost:5000/counter?clear=true
 * Cleared!
 * $ curl -b cookies.txt -c cookies.txt localhost:5000/counter
 * You have called this endpoint 1 times.
 *
 */
include("lib/Session");

Acme.Proxy.Counter = function() {
  this.handle = function(request) {
    var response = new AP.HTTP.Response();

    if (request.params.clear) {
        session.delete("num");
        response.body = "Cleared!\n";
        return response;
    }

    var num = 0;
    if (session.isSet("num")) {
        num = session.get("num");
    }
    num += 1;
    session.set("num", num);

    response.body = "You have called this endpoint " + num + " times.\n";
    return response;
  };
}
