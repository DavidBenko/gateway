if (request.headers["X-Sharedsecret"] != "12345") {
  response.statusCode = 401;
  response.setJSONBodyPretty({error: "Access denied!"});
} else {
  response.body = "Super Secret Information\n";
}
