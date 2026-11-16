package main

import (
	"errors"
	"fmt"
	"strconv"
)





func ParseInput (data []byte)  ([]string , error)  {
	pos := 0 

	if len(data) == 0 {
		return nil , errors.New("empty input")
	}


	if data[pos] != '*' {
		return nil , fmt.Errorf("Expected * found %q" , data[pos])
	}

	pos += 1 

	strNum , pos , err := readLine(data , pos)
	if err != nil {
		return nil , err
	}

	num , err := strconv.Atoi(strNum)
	if err != nil {
		return nil , err 
	}

    tmp := []string{}

	for i := 0 ; i < num ; i++ {
		if pos >= len(data) {
			return nil , fmt.Errorf("expected %d elements, got %d" , num , i)
		}

		if data[pos] != '$' {
			return nil , fmt.Errorf("Expected $ found %q" , data[pos])
		}

		pos += 1 
		strNum , newPos , err := readLine(data , pos)
		if err != nil {
			return nil , err
		}
		
		num , err := strconv.Atoi(strNum)
		if err != nil {
			return nil , err 
		}

		if num < 0 || newPos+num+2 > len(data) {
			return nil , fmt.Errorf("bulk string length %d exceeds input" , num)
		}

		tmp = append(tmp, string(data[newPos:(newPos+num)]))
		pos = newPos + num + 2
	}

	return tmp , nil
}


func readLine (data []byte , pos int) (string , int , error){
	for i := pos ; i < len(data)-1 ; i++ {
	
		if data[i] == '\r' && data[i+1] == '\n' {
			return string(data[pos:i]) , i+2 , nil 
		}
	}

	return "" , 0 , fmt.Errorf("Expected \\r\\n , not found")
}