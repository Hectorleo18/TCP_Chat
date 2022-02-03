package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

var (
	errNotDefinedFormat = errors.New("not defined format")

	msgNotDefinedFormat = "Formato no definido."
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost:8888")
	if err != nil {
		log.Fatal(err)
	}

	// Connect to the TCP server
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	go printOutput(conn)
	writeInput(conn)
}

// writeInput write a command or message to the server
func writeInput(conn *net.TCPConn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		command := strings.Split(text, " ")
		//Check if the command is "/file"
		if command[0] == "/file" {
			//Delete the first element of the slice
			copy(command[0:], command[1:])
			command[len(command)-1] = ""
			command = command[:len(command)-1]
			//Convert to string and delete the line break
			completeCommand := strings.Join(command, " ")
			completeCommand = strings.Trim(completeCommand, "\r\n")
			fileInfo, err := os.Stat(completeCommand)
			if err != nil {
				fmt.Println(err)
				continue
			}
			//get the name of the file
			fileName := fileInfo.Name()
			fmt.Println("Nombre: " + fileName)
			//Send the command and the file name
			conn.Write([]byte("/file " + fileName + " "))
			SendFile(completeCommand, conn)
		} else {
			//If the command isn't "/file"
			err = writeMsg(conn, text)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

// printOutput prints the message received from the server
func printOutput(conn *net.TCPConn) {
	for {
		msg, err := readMsg(conn)
		// if the error is EOF the connection is over
		if err == io.EOF {
			conn.Close()
			fmt.Println("Connection Closed. Bye bye.")
			os.Exit(0)
		}
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(msg)
	}
}

// SendFile sends file to the server
func SendFile(filePath string, conn net.Conn) {
	//gets the content of the file
	f, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	//Turn the file to string
	text := string(f[:])
	//Replace "\n" with another text
	text = strings.ReplaceAll(text, "\n", "HectorLeoRodriguez")
	//Send the file to the server
	conn.Write([]byte(text))
	conn.Write([]byte("\n"))
}

//writeMsg writes a text message to the server
func writeMsg(conn net.Conn, msg string) error {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		return err
	}
	return nil
}

//readMsg reads the message from the server
func readMsg(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	//Variables to return
	var msg string
	var err error
	//Read the first word from the server
	typeMsg, err := reader.ReadString(' ')
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	//if the word is "file"
	if typeMsg == "file " {
		//Read the server until finds a line break
		data, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		data = strings.Trim(data, "\r\n")
		//Convert the string in []string
		args := strings.Split(data, " ")
		//Create an empty file
		f, err := os.Create(args[2])
		if err != nil {
			return "", err
		}
		defer f.Close()
		//Create a string without the first 3 words
		text := strings.Join(args[3:], " ")
		//Replace the string with line break
		text = strings.ReplaceAll(text, "HectorLeoRodriguez", "\n")
		//Write the information in the file
		srcFile := []byte(text)
		f.Write(srcFile)
		f.Close()
		msg = strings.Join(args[:3], " ")
	} else {
		//If the first word is "msg"
		if typeMsg == "msg " {
			//Read until the line break
			data, err := reader.ReadString('\n')
			msg = data
			if err != nil {
				msg = ""
				fmt.Println(err)
			}
		} else {
			//If the first word is anther one
			msg = msgNotDefinedFormat
			err = errNotDefinedFormat
		}
	}
	return msg, err
}
