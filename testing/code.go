package testing

import "fmt"

// pointer to function and can be overridden during run time from testing (can be used to stubb data base connections and other things)
var bringValue = getValue

func checking(a, b int) int {
	return print(a, b, bringValue())
}

func getValue() int {
	return 2;
}
func print(a, b, c int) int {
	fmt.Println("value is ", a+b+c)
	return a + b + c
}
