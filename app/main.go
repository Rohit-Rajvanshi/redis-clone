package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type cfg struct {
		data map[string]string
		mu sync.RWMutex
	}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment the code below to pass the first stage
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
	fmt.Println("Failed to bind to port 6379")
	os.Exit(1)
	}

	config := cfg{
				data : make(map[string]string) ,
			}

	for{
	conn , err := l.Accept()
	if err != nil {
	fmt.Println("Error accepting connection: ", err.Error())
	os.Exit(1)
	}
	go handler(conn , &config)
	}
}


func handler(conn net.Conn , config *cfg){
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
			if len(tmp) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'echo' command\r\n"))
				continue
			}
			rsp := fmt.Sprintf("$%d\r\n%s\r\n" , len(tmp[1]) , tmp[1])
			conn.Write([]byte(rsp))
		case "set":

			if len(tmp) < 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'set' command\r\n"))
				continue
			}

			config.mu.Lock()
			config.data[tmp[1]] = tmp[2]
			config.mu.Unlock()


			conn.Write([]byte("+OK\r\n"))


		case "get":
			if len(tmp) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'get' command\r\n"))
				continue
			}

			config.mu.RLock()
			val, ok := config.data[tmp[1]]
			config.mu.RUnlock()

			if ok {
				fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(val), val)
			} else {
				conn.Write([]byte("$-1\r\n"))
			}


		default :
			conn.Write([]byte("+PONG\r\n"))
		}	
	}
}