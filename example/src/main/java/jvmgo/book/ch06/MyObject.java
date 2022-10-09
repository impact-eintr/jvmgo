package jvmgo.book.ch06;

public class MyObject {

    public int instanceVar;

    public static void main(String[] args) {
        int x = 32768; // ldc
        MyObject myObj = new SubObject(); // new
        //SubObject.staticVar = x; // putstatic
        //x = SubObject.staticVar; // getstatic
        //myObj.instanceVar = x; // putfield
        //x = myObj.instanceVar; // getfield
        //Object obj = myObj;
        //if (obj instanceof MyObject) { // instanceof
        //    myObj = (MyObject) obj; // checkcast
        //    System.out.println(myObj.instanceVar);
        //}
    }

}

class SubObject extends MyObject {
    public static int staticVar;
}
