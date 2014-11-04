function main(proxyRequest) {
	var info = {
		sessionID: proxyRequest.headers["X-Sessionid"],
		serverURL: proxyRequest.headers["X-Serverurl"],
		limit: Math.min(proxyRequest.params.limit || 100, 100),
		offset: proxyRequest.params.offset || 0
	}
	
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
	
	var request = new AP.HTTP.Request();
	request.method = "POST";
	request.url = info.serverURL;
	request.headers["Content-Type"] = "text/xml;charset=UTF-8";
	request.headers["SOAPAction"] = '""';
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
