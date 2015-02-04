/**
 * Requests a truncated version of the accounts from Salesforce by
 * delegating to its SOAP API.
 * 
 * After logging in:
 * 
 * $ curl -b cookies.txt -c cookies.txt http://localhost:5000/accounts?limit=1
 * [
 *    {
 *       "id": "001d000001MEZm4AAH",
 *       "name": "Aaron"
 *    }
 * ]
 * 
 * $ curl -b cookies.txt -c cookies.txt "http://localhost:5000/accounts?limit=1&offset=100"
 * [
 *    {
 *       "id": "001d000001MEgKJAA1",
 *       "name": "Adidas Corporation"
 *    }
 * ]
 * 
 * Or, if logged out:
 * $ curl -b cookies.txt -c cookies.txt http://localhost:5000/accounts?limit=1
 * {
 *    "error": "Unauthorized"
 * }
 * 
 */

include("Salesforce");

function main(proxyRequest) {
	if (!Salesforce.isLoggedIn()) {
		return Salesforce.unauthorizedResponse();
	}
	
	var info = {
		sessionID: session.get("sessionID"),
		limit: Math.min(proxyRequest.params.limit || 100, 100),
		offset: proxyRequest.params.offset || 0
	};
	
	var body = '<?xml version="1.0" encoding="utf-8"?>\
	<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:enterprise.soap.sforce.com">\
	  <soapenv:Header>\
        <urn:SessionHeader>\
          <urn:sessionId>{{sessionID}}</urn:sessionId>\
        </urn:SessionHeader>\
      </soapenv:Header>\
	  <soapenv:Body>\
	    <urn:query>\
          <urn:queryString>SELECT Id, Name FROM Account ORDER BY Name LIMIT {{limit}} OFFSET {{offset}}</urn:queryString>\
	    </urn:query>\
	  </soapenv:Body>\
	</soapenv:Envelope>';
	
	var request = Salesforce.newRequest();
	request.body = _.template(body)(info);
	var proxyResponse = AP.makeRequest(request);
			
	// Prepare response
	var response = new AP.HTTP.Response();
	
	// Check for errors
	if (proxyResponse.statusCode != 200) {
		response.statusCode = 400;
		response.setJSONBodyPretty({"error": "Bad request."});
		return response;
	}
	
	var responseBody = proxyResponse.body;
	
	var regex = /<records xsi:type=\"sf:Account\"><sf:Id>(.+?)<\/sf:Id><sf:Name>(.+?)<\/sf:Name><\/records>/g;
	var match;
	var accounts = [];
	while ((match = regex.exec(responseBody)) !== null) {
		accounts.push({id: match[1], name: match[2]})
	}
	response.setJSONBodyPretty(accounts);
	return response;
}
