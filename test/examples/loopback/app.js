// To avoid boilerplate in endpoint code.
Acme = {};
Acme.Proxy = {};
Acme.Static = {};
Acme.Middleware = {};

include("Acme.Middleware.Logger")

var App = new AP.Gateway();
App.middleware = [
	Acme.Middleware.Logger
];
