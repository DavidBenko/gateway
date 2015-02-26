(function() {
    if (request.params.clear) {
        session.delete("num");
        response.body = "Cleared!\n";
        return;
    }

    var num = 0;
    if (session.isSet("num")) {
        num = session.get("num");
    }
    num += 1;
    session.set("num", num);

    response.body = "You have called this endpoint " + num + " times.\n";
})();
