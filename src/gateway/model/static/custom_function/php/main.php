<?php
$json = json_decode(file_get_contents("php://stdin"), true);
print("test message");
print("\x00\x00\x00\x00\x00\x00\x00\x00");
print(json_encode($json));
?>
