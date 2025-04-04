package service

import "fmt"

func ExampleGetRandString() {
	testString1 := "testString1"
	out1 := GetRandString(testString1)
	fmt.Println(out1)

	testString2 := "testString2"
	out2 := GetRandString(testString2)
	fmt.Println(out2)

	testString3 := ""
	out3 := GetRandString(testString3)
	fmt.Println(out3)

	// Output:
	//d82007aa
	//fc1084e4
	//da39a3ee
}
