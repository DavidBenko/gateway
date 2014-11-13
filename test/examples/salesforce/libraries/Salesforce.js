/**
 * Sets up a global session
 */
var session = new AP.Session("salesforce");

/**
 * Sets the template settings to be Mustache like.
 */
_.templateSettings = {
  interpolate : /\{\{(.+?)\}\}/g
};

var Salesforce = Salesforce || {};

/**
 * Returns a new HTTP request set up to be SOAP-y.
 * If already authenticated, sets the URL to what was returned.
 */
Salesforce.newRequest = function() {
	var request = new AP.HTTP.Request();
	request.method = "POST";
	request.headers["Content-Type"] = "text/xml;charset=UTF-8";
	request.headers["SOAPAction"] = '""';
	if (session.isSet("serverURL")) {
		request.url = session.get("serverURL");
	}
	return request;
}

Salesforce.isLoggedIn = function() {
	return session.isSet("sessionID");
}

Salesforce.unauthorizedResponse = function() {
	var response = new AP.HTTP.Response();	
	response.statusCode = 401;
	response.setJSONBodyPretty({"error": "Unauthorized"});
	return response;
}