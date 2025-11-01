package main

import "fmt"

type Greeter struct {
	Name string
}

func (g Greeter) Greet() {
	fmt.Printf("Hello, %s!", g.Name)
}

func main() {
	g := Greeter{Name: "World"}
	g.Greet()
}
