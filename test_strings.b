/* Test string handling */

string_table[3] "Hello", "World", "Test";

main()
{
    extrn string_table;
    auto i;
    i = 0;
    while (i < 3) {
        write(string_table[i]);
        i++;
    }
}
