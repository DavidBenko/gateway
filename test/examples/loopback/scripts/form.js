debug.request = new AP.HTTP.Request();
debug.request.setFormBody({
    "foo": "bar",
    "array": [1, 2, 3],
    "object": {
        "a": "b",
        "c": "d"
    },
    "nested": {
        "array": ["a", "b", "c"]
    }
});
