package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)


func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment the code below to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
	fmt.Println("Failed to bind to port 6379")
	os.Exit(1)
	}

	for{
	conn , err := l.Accept()
	if err != nil {
	fmt.Println("Error accepting connection: ", err.Error())
	os.Exit(1)
	}
	go handler(conn)
	}
}


func handler(conn net.Conn){
	defer conn.Close()
	buff := make([]byte , 1024)
	for{
		byteArg , err := conn.Read(buff) 
		if err != nil {
			fmt.Println(err.Error())
			break
		}

		tmp , err := ParseInput(buff[:byteArg])
		if err != nil {
			fmt.Println("Error Parsing the Input" , err)
			break 
		}

		switch  strings.ToLower(tmp[0]){
		case "echo":
			rsp := fmt.Sprintf("$%d\r\n%s\r\n" , len(tmp[1]) , tmp[1])
			conn.Write([]byte(rsp))
		}	
	}
}