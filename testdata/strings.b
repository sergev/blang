/* Test string literals */

messages[3] "Hello", "World", "Test";

main() {
    extrn messages;
    write(messages[0]);
    write(messages[1]);
    write(messages[2]);
}

