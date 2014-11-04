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
	
	var request = new AP.HTTP.Request();
	request.method = "POST";
	request.url = "https://login.salesforce.com/services/Soap/c/28.0";
	request.headers["Content-Type"] = "text/xml;charset=UTF-8";
	request.headers["SOAPAction"] = '""';
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
	
	response.setJSONBodyPretty({"serverURL": serverMatch[1], "sessionID": sessionMatch[1]});
	return response;
}
