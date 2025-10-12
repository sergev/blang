/* Test string literals */

messages[3] "Hello", "World", "Test";

main() {
    extrn messages;
    printf("%s*n", messages[0]);
    printf("%s*n", messages[1]);
    printf("%s*n", messages[2]);
}
