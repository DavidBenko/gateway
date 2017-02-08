import json
import sys

def Input():
  return json.loads(sys.stdin.read())
def Output(output):
  print("\x00\x00\x00\x00\x00\x00\x00\x00")
  print(output)

j = Input()
print("test message")
Output(j)
