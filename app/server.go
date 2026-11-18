package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

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
						conn.Write([]byte("-ERR error parsing the duration for expiry given\r\n"))
						continue 
					}
				}else if strings.ToLower(expiryFormat) == "px" {
					expiryDuration , err = time.ParseDuration(fmt.Sprintf("%sms" , tmp[4]))
					if err != nil {
						conn.Write([]byte("-ERR error parsing the duration for expiry given\r\n"))
						continue 
					}
				}else{
					conn.Write([]byte("-ERR invalid format to Parse , choose either PX or EX\r\n"))
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


			args := []string{}
			for i := 2 ; i < len(tmp) ; i++ {
				args = append(args , tmp[i])
			}

			length := 0
			wrongType := false

			config.mu.Lock()
			x , ok := config.data[tmp[1]]
			if ok{
				v , ok := x.val.([]string)
				if ok{	
						v = append(v , args...)
						x.val = v
						length  = len(v)
					}else{
						wrongType = true 
					}
			}else{	
				config.data[tmp[1]] = &valueBody{
					val: args,
				}
				length = len(args)
			}
			config.mu.Unlock()

			if !wrongType{
				fmt.Fprintf(conn , "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n")
				continue 
			}

			str := fmt.Sprintf(":%d\r\n" , length)
			conn.Write([]byte(str))

		default :
			conn.Write([]byte("+PONG\r\n"))
		}	
	}
}