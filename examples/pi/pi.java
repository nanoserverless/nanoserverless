public class app {
  public static void main(String[] args) {
    int c = 100000000;
    double pi = 0;
    int n = 1;
    for (int i = 0; i < c; i++) {
      pi += (4/n)-(4/(n+2));
      n += 4;
    }
    System.out.println(pi);
  }
}
