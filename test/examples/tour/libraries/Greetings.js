function RandomGreeting() {
  var greetings = ["Hello", "Hi", "Sup?", "How's it going?", "Greetings", "Salutations"];
  return greetings[Math.floor(Math.random()*6)]
}

var Greetings = Greetings || {};

Greetings.Response = function() {
  this.statusCode = 200;
  this.body = null;
  this.headers = {"Content-Type": "application/json"};
}

Greetings.Response.prototype.setBody = function(greeting) {
  this.body = JSON.stringify({"greeting": greeting}, null, "   ") + "\n";
}
