// To avoid boilerplate in endpoint code.
Acme = {};
Acme.Proxy = {};
Acme.Static = {};
Acme.Middleware = {};

include("Acme.Middleware.Logger")
include("AP.Middleware.CORS")

var App = new AP.Gateway();
App.middleware = [
	AP.Middleware.CORS,
	Acme.Middleware.Logger
];
