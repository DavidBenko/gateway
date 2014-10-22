router.path("/helloworld").methods("GET", "POST").name("hi");
router.path("/helloworld/{id}").methods("GET", "POST").name("hi");