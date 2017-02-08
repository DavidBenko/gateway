import java.util.Arrays;
import org.json.JSONObject;

public class Main {
  public static void main(String[] args) throws Exception {
    JSONObject json = Input();
    System.out.println("test message");
    Output(json);
  }

  private static JSONObject Input() throws Exception {
    byte[] buffer = new byte[256];
    int b;
    int i = 0;

    while (true) {
      b = System.in.read();
      if (b == -1)
          break;
      if (i >= buffer.length) {
        buffer = Arrays.copyOf(buffer, 2 * buffer.length);
      }
      buffer[i++] = (byte) b;
    }

    return new JSONObject(new String(Arrays.copyOfRange(buffer, 0, i), "UTF-8"));
  }

  private static void Output(JSONObject json) {
    System.out.print("\000\000\000\000\000\000\000\000");
    System.out.println(json);
  }
}
