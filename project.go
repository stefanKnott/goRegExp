package main

import (
	"fmt"
	"os"
	"sync"
	"bufio"
	"net"
	"regexp"
	)

//  The Queue library was created by Hicham Bouabdallah
// (Minus the Populate(fileName string) and Resolve() functions which were written by Stefan Knott
//  Copyright (c) 2012 SimpleRocket LLC
//
//  Permission is hereby granted, free of charge, to any person
//  obtaining a copy of this software and associated documentation
//  files (the "Software"), to deal in the Software without
//  restriction, including without limitation the rights to use,
//  copy, modify, merge, publish, distribute, sublicense, and/or sell
//  copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following
//  conditions:
//
//  The above copyright notice and this permission notice shall be
//  included in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
//  EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
//  OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
//  HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
//  FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.
type queuenode struct {
	data interface{}
	next *queuenode
}

//	A go-routine safe FIFO (first in first out) data stucture.
type Queue struct {
	head  *queuenode
	tail  *queuenode
	count int
	lock  *sync.Mutex
}


//	Creates a new pointer to a new queue.
func NewQueue() *Queue {
	q := &Queue{}
	q.lock = &sync.Mutex{}
	return q
}

//	Returns the number of elements in the queue (i.e. size/length)
//	go-routine safe.
func (q *Queue) Len() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.count
}

//	Pushes/inserts a value at the end/tail of the queue.
//	Note: this function does mutate the queue.
//	go-routine safe.
func (q *Queue) Push(item interface{}) {
	q.lock.Lock()
	defer q.lock.Unlock()

	n := &queuenode{data: item}

	if q.tail == nil {
		q.tail = n
		q.head = n
	} else {
		q.tail.next = n
		q.tail = n
	}
	q.count++
}

//	Returns the value at the front of the queue.
//	i.e. the oldest value in the queue.
//	Note: this function does mutate the queue.
//	go-routine safe.
func (q *Queue) Poll() interface{} {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.head == nil {
		return nil
	}

	n := q.head
	q.head = n.next

	if q.head == nil {
		q.tail = nil
	}
	q.count--

	return n.data
}

//	Returns a read value at the front of the queue.
//	i.e. the oldest value in the queue.
//	Note: this function does NOT mutate the queue.
//	go-routine safe.
func (q *Queue) Peek() interface{} {
	q.lock.Lock()
	defer q.lock.Unlock()

	n := q.head
	if n == nil || n.data == nil {
		return nil
	}

	return n.data
}

//----------------------------------------------------------------------
//Written by Stefan Knott
//----------------------------------------------------------------------

var domainRegexp = regexp.MustCompile(`^(([a-zA-Z]{1})|([a-zA-Z]{1}[a-zA-Z]{1})|([a-zA-Z]{1}[0-9]{1})|([0-9]{1}[a-zA-Z]{1})|([a-zA-Z0-9][a-zA-Z0-9-_]{1,61}[a-zA-Z0-9]))\.([a-zA-Z]{2,6}|[a-zA-Z0-9-]{2,30}\.[a-zA-Z]{2,3})$`)
var emailRegexp = regexp.MustCompile(`^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$`)
var phoneRegexp = regexp.MustCompile(`^(\+0?1\s)?\(?\d{3}\)?[\s.-]\d{3}[\s.-]\d{4}$`)

//Populates queue with info contained in fileName
func(q *Queue) Populate(fileName string) {

	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("ERROR")
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan(){
		line := scanner.Text()
		q.Push(line)
	}
}

//Pop (Poll) from queue and perform necesarry op depending whether we are reading a domain name, email address, or phone number
func(q *Queue) Resolve() {
	dnsFile, err := os.Create("dnslookup.txt")
	if err != nil{
		fmt.Printf("ERROR")
	}
	defer dnsFile.Close()

	emailFile, err := os.Create("emailAddrs.txt")
		if err != nil{
			fmt.Printf("ERROR")
		}
	defer emailFile.Close()

	phoneFile, err := os.Create("phoneNumbers.txt")
		if err != nil{
			fmt.Printf("ERROR")
		}
	defer phoneFile.Close()

		for {
			item := q.Poll()
		//convert item of type interface{} to a string type and do regex on it
			if str, ok := item.(string); ok && len(str) > 0{	
			//string str matched to a domain name
				if retStr := domainRegexp.FindString(str); len(retStr) != 0{			
					//ip, err := net.ResolveIPAddr("ip4", retStr)
					ip, err := net.ResolveIPAddr("ip4", retStr)

					if err == nil{

						fmt.Printf("Domain: %s\t IP: %s\n", retStr, ip.String())
						toFile := []byte("Domain: " + retStr + "\t" + "IP: " + ip.String() + "\n")
						//fmt.Println(toFile)
						dnsFile.Write(toFile)
        			} else {
             		   fmt.Printf("%s error: %s\n", retStr, err)
        			}
        	//string str matched to an email address
    			} else if retStr = emailRegexp.FindString(str); len(retStr) != 0{
    				fmt.Printf("Email: %s\n", retStr)
    				toFile := []byte(retStr + "\n")
    				emailFile.Write(toFile)

    			} else if retStr = phoneRegexp.FindString(str); len(retStr) != 0{
    				fmt.Printf("US Phone: %s\n", retStr)
    				toFile := []byte(retStr + "\n")
    				phoneFile.Write(toFile)
    			}
			}
			if q.count == 0{ break}
		}
}

//Open a file, match on domain names, do dns lookup on said names
func main(){
	fileName := os.Args[1]
	queue := NewQueue()
	queue.Populate(fileName)
	queue.Resolve()
}