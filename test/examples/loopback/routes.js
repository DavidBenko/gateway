/**
 * These routes mock out a back end, so that you don't have to
 * run another server for little experimentation.
 */
 router.get("/", "Acme.Static.HelloWorld");
 
 router.get("/topsecret", "Acme.Static.TopSecret");


 /*
 * You can also use this tool can be used to provide value
 * directly. You could write your own business logic in Gateway
 * without requiring another service to proxy to.
 *
 * Useful? Probably most in conjunction with another service. But yes.
 */
router.post("/secret", "Acme.Proxy.Secret");
router.path("/{e|E}cho").name("Acme.Proxy.Echo");
router.get("/greetings", "Acme.Proxy.Greetings");
router.get("/counter", "Acme.Proxy.Counter");
router.get("/env", "Acme.Proxy.Environment");
// router.get("/widgets", "Acme.Proxy.Widgets");


/**
 * These endpoints build on the endpoints above to highlight some of
 * the things the proxy can do as a proxy.
 */
router.get("/proxy", "Acme.Proxy.Proxy");
router.path("/proxyEcho").name("Acme.Proxy.Proxy");
router.get("/error", "Acme.Proxy.ErrorHandler");
router.get("/composite", "Acme.Proxy.Composite");
router.get("/workflow", "Acme.Proxy.Workflow");


/**
 * This just shows some routing functionality.
 *
 * The following route matches only if:
 *   - path looks like /routed/1111 (any number of ones, nothing else)
 *   - HTTP method is 'CUSTOM'
 *   - query parameter x is present and a, b, or c
 *
 * $ curl -X CUSTOM localhost:5000/routed/1111?x=a
 * Hello, world!
 * $ curl -X CUSTOM localhost:5000/routed/1111?x=d
 * 404 page not found
 * $ curl localhost:5000/routed/1111?x=d
 * 404 page not found
 * $ curl localhost:5000/routed/1111?x=a
 * 404 page not found
 * $ curl -X CUSTOM localhost:5000/routed/12?x=a
 * 404 page not found
 */
router.path("/routed/{id:1+}").name("Hello World").methods("CUSTOM").queries({"x": "{x:[a-c]}"});
