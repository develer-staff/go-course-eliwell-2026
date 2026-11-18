package main

import "fmt"

type User struct {
	ID   int
	Name string
}

// In C++, returning &u from the stack is a fatal dangling pointer bug!
// In Go, the compiler runs escape analysis: u escapes to heap automatically.
func CreateUser(id int, name string) *User {
	u := User{ID: id, Name: name}
	return &u
}

func main() {
	user := CreateUser(42, "Alice")
	fmt.Printf("User: %+v, Address: %p\n", user, user)
}
