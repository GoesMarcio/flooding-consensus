// Em caso de erro utilizar o comando GO111MODULE=off no inicio da execução no terminal
// EX: GO111MODULE=off go run main.go 127.0.0.1:5001  127.0.0.1:6001  127.0.0.1:7001

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	Flooding "./Flooding"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please specify at least one address:port!")
		fmt.Println("go run main.go 127.0.0.1:5001  127.0.0.1:6001  127.0.0.1:7001")
		fmt.Println("go run main.go 127.0.0.1:6001  127.0.0.1:5001  127.0.0.1:7001")
		fmt.Println("go run main.go 127.0.0.1:7001  127.0.0.1:6001  127.0.0.1:5001")
		return
	}


	args := os.Args[1:] // lista de ip:port dos outros processos
	addresses := []string{}
	round := -1

	for i, arg := range args {
		if arg == "-c" {
			round, _ = strconv.Atoi(args[i+1])
		}
		if strings.Contains(arg, "."){
			addresses = append(addresses, arg)
		}
	}


	module := Flooding.Flooding_Module{}

	module.Init(addresses, round)

	blq := make(chan int)
	<-blq
}
