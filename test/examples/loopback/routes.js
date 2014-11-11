/**
 * Primary Source
 * 
 * These routes mock out a back end, so that you don't have to
 * run another server for little experimentation.
 * 
 * It also shows how this tool can be used to provide value
 * directly. You could write your own business logic in Gateway
 * without requiring another service to proxy to.
 * 
 * Useful? Probably most in conjunction with another service. But yes.
 */
router.get("/", "Hello World");
router.path("/echo").methods("GET", "POST").name("Echo");
router.get("/foo", "Foo");
router.get("/bar", "Bar");
router.post("/secret", "Secret");
router.get("/topsecret", "Top Secret");
router.get("/greetings", "Greetings");
router.get("/counter", "Counter");
router.get("/env", "Environment");

/**
 * Proxy Endpoints
 * 
 * These endpoints build on the endpoints above to highlight some of
 * the things the proxy can do.
 */
router.get("/proxy", "Proxy");
router.get("/error", "Error Handling");
router.get("/composite", "Composite");
router.get("/workflow", "Workflow");

/**
 * Router Examples
 *
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
