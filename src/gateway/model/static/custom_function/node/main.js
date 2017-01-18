var stdin = process.stdin,
    stdout = process.stdout,
    inputChunks = [];

stdin.resume();
stdin.setEncoding('utf8');

stdin.on('data', function (chunk) {
    inputChunks.push(chunk);
});

function output(data) {
  stdout.write("\x00\x00\x00\x00\x00\x00\x00\x00");
  stdout.write(data);
}

console.log("test message");

stdin.on('end', function () {
    var inputJSON = inputChunks.join(""),
        parsedData = JSON.parse(inputJSON),
        outputJSON = JSON.stringify(parsedData, null, '    ');
    output(outputJSON)
});
