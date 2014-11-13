/**
 * Logs the user out of the service by deleting the values from the session.
 * 
 * $ curl -X DELETE http://localhost:5000/sessions
 * {
 *    "success": true
 * }
 * 
 */

include("Salesforce");

function main(proxyRequest) {
	session.delete("serverURL");
	session.delete("sessionID");

	var response = new AP.HTTP.Response();
	response.setJSONBodyPretty({"success": true});
	return response;
}
