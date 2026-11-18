package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type cfg struct {
		data map[string]*valueBody
		mu sync.RWMutex
	}

type valueBody struct {
	val any
	expiryTime time.Time
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
				data : make(map[string]*valueBody) ,
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

			if len(tmp) == 4 {
				conn.Write([]byte("-ERR wrong number of arguments for 'set' command , provide a duration for expiry\r\n"))
				continue
			}

			expiryTime := time.Time{}

			if len(tmp) > 3 {
				expiryFormat := tmp[3]

				var expiryDuration time.Duration

				if strings.ToLower(expiryFormat)== "ex" {	
					expiryDuration , err = time.ParseDuration(fmt.Sprintf("%ss" , tmp[4]))
					if err != nil {
						conn.Write([]byte("-ERR error parsing the duration for expiry given"))
						continue 
					}
				}else if strings.ToLower(expiryFormat) == "px" {
					expiryDuration , err = time.ParseDuration(fmt.Sprintf("%sms" , tmp[4]))
					if err != nil {
						conn.Write([]byte("-ERR error parsing the duration for expiry given"))
						continue 
					}
				}else{
					conn.Write([]byte("-ERR invalid format to Parse , choose either PX or EX"))
					continue 
				}

				expiryTime = time.Now().Add(expiryDuration)	
			}

			config.mu.Lock()
			config.data[tmp[1]] = &valueBody{
				val: tmp[2],
				expiryTime: expiryTime,
			}
			config.mu.Unlock()


			conn.Write([]byte("+OK\r\n"))


		case "get":
			if len(tmp) < 2 {
				conn.Write([]byte("-ERR wrong number of arguments for 'get' command\r\n"))
				continue
			}

			currentTime := time.Now()

			config.mu.RLock()
			val, ok := config.data[tmp[1]]
			config.mu.RUnlock()

			if ok {
				if val.expiryTime.IsZero() || !currentTime.After(val.expiryTime){
					v , ok := val.val.(string)
					if ok{	
						fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(v), v)
					}else{
						fmt.Fprintf(conn , "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n")
					}
				}else{
					config.mu.Lock()
					cur := config.data[tmp[1]].expiryTime
					if currentTime.After(cur) && !cur.IsZero(){
						delete(config.data , tmp[1])
					}
					config.mu.Unlock()
					conn.Write([]byte("$-1\r\n"))
				}
			}else {	
				conn.Write([]byte("$-1\r\n"))
			}
			
		case "rpush": 
			if len(tmp) < 3 {
				conn.Write([]byte("-ERR wrong number of arguments for 'rpush' command\r\n"))
				continue
			}

			
			expiryTime := time.Time{}


			config.mu.Lock()	
			_ , ok := config.data[tmp[1]]
			if ok{

				v , ok := config.data[tmp[1]].val.([]string)
				if ok{	
						v = append(v , tmp[2])
						config.data[tmp[1]].val = v
						length  := len(v)
						str := fmt.Sprintf(":%d\r\n" , length)
						conn.Write([]byte(str))
					}else{
						fmt.Fprintf(conn , "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n")
					}
			}else{	
				config.data[tmp[1]] = &valueBody{
					val: []string{tmp[2]},
					expiryTime: expiryTime,
				}

				conn.Write([]byte(":1\r\n"))
			}
			config.mu.Unlock()

		default :
			conn.Write([]byte("+PONG\r\n"))
		}	
	}
}