/**
 * Logs the user into the service, by delegating to the Salesforce SOAP API.
 * 
 * $ curl -b cookies.txt -c cookies.txt \
 *      -d '{"email":"qa_alias1@anypresence.com","password":"badpass","token":"orbadtoken"}' \
 *      http://localhost:5000/sessions/new
 * {
 *    "error": "Invalid credentials."
 * }
 * 
 * $ curl -b cookies.txt -c cookies.txt \
 *      -d '{"email":"qa_alias1@anypresence.com","password":"<omitted>","token":"<omitted>"}' \
 *      http://localhost:5000/sessions/new
 * {
 *    "success": true
 * }
 * 
 */

include("Salesforce");

function main(proxyRequest) {
	var credentials = JSON.parse(proxyRequest.body);
	
	var body = '\<?xml version="1.0" encoding="utf-8"?>\
	<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:enterprise.soap.sforce.com">\
	   <soapenv:Body>\
	      <urn:login>\
	         <urn:username>{{email}}</urn:username>\
	         <urn:password>{{password}}{{token}}</urn:password>\
	      </urn:login>\
	   </soapenv:Body>\
	</soapenv:Envelope>';
	
	var request = Salesforce.newRequest();
	request.url = AP.Environment.get("loginURL");
	request.body = _.template(body)(credentials);
	var proxyResponse = AP.makeRequest(request);
	
	// Prepare response
	var response = new AP.HTTP.Response();
	
	// Check for errors
	if (proxyResponse.statusCode != 200) {
		response.statusCode = 401;
		response.setJSONBodyPretty({"error": "Invalid credentials."});
		return response;
	}
	
	var responseBody = proxyResponse.body;
	
	var sessionRegex = /<sessionId>(.*)<\/sessionId>/;
	var sessionMatch = sessionRegex.exec(responseBody);
	
	var serverRegex = /<serverUrl>(.*)<\/serverUrl>/;
	var serverMatch = serverRegex.exec(responseBody);
	
	// Make sure response was formatted as we expect
	if (sessionMatch == null || serverMatch == null) {
		response.statusCode = 500;
		response.setJSONBodyPretty({"error": "Error parsing response."});
		return response;
	}
	
	session.set("serverURL", serverMatch[1]);
	session.set("sessionID", sessionMatch[1]);
	
	response.setJSONBodyPretty({"success": true});
	return response;
}
