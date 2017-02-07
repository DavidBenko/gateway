using System;
using System.IO;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;

public class main {
  static private JObject Input() {
    MemoryStream ms = new MemoryStream();
    Console.OpenStandardInput().CopyTo(ms);
    byte[] buffer = ms.ToArray();
    string j = System.Text.Encoding.UTF8.GetString(buffer, 0, buffer.Length);
    return JObject.Parse(j);
  }

  static private void Output(JObject output) {
    Console.Write("\x00\x00\x00\x00\x00\x00\x00\x00");
    Console.Write(output.ToString());
  }

  static public void Main() {
    JObject json = Input();
    Console.Write("test message");
    Output(json);
  }
}
