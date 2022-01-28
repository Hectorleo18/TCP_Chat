package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"log"
	"bufio"
	"bytes"
	"encoding/binary"
)

func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost:8888")
	if err != nil {
		log.Fatal(err)
	}

	// Conectar al servidor TCP
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	go printOutput(conn)
	writeInput(conn)
}

func writeInput(conn *net.TCPConn){
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		command := strings.Split(text," ")
		//Revisar si el comando es /msg para leer el archivo.
		if command[0] == "/file" {
			fileInfo, err := os.Stat(command[1])
			if err != nil {
				fmt.Println(err)
				return
			}
			fileName := fileInfo.Name()
			fileSize := fileInfo.Size()
			fmt.Println("Nombre: "+fileName)
			conn.Write([]byte("/file "+fileName+" "))//<----Ojo aquí
			// buf := make([]byte, 2048)
			// n, err := conn.Read(buf)
			// if err != nil {
			// 	fmt.Println(err)
			// 	return
			// }
			// revData := string(buf[:n])
			// if revData == "ok" {
				//Send file data
			SendFile(command[1], fileSize, conn)
			// }
		}else{
			//Si el usuario manda otro comando que no sea file
			err = writeMsg(conn, text)
			if err != nil {
				log.Println(err)
			}
		}

		//https://github.com/pplam/tcp-file-transfer
	}
}

func printOutput(conn *net.TCPConn) {
	for {
		msg, err := readMsg(conn)
		// Receiving EOF means that the connection has been closed
		if err == io.EOF {
			// Close conn and exit
			conn.Close()
			fmt.Println("Connection Closed. Bye bye.")
			os.Exit(0)
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(msg)
	}
}

//Send a file to the server
func SendFile(filePath string, fileSize int64, conn net.Conn) {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	var count int64
	for {
		buf := make([]byte, 2048)
		//Read file content
		n, err := f.Read(buf)
		if err != nil && io.EOF == err {
			fmt.Println("File Transfer")
			//Tell the server end file reception
			// conn.Write([]byte("finish \n"))
			return
		}
		//Send to the server
		conn.Write(buf[:n])

		count += int64(n)
		sendPercent := float64(count) / float64(fileSize) * 100
		value := fmt.Sprintf("%.2f", sendPercent)
		//Print upload progress
		fmt.Println("Upload:" + value + "%")
	}
}

func writeMsg(conn net.Conn, msg string) error {
	// Send the message
	_, err := conn.Write([]byte(msg))
	if err != nil {
		return err
	}
	return nil
}

func readMsg(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	var msg string
	var err error
	typeMsg, err := reader.ReadString(' ')
	if err != nil {
		fmt.Println(err)
	}
	if typeMsg == "file"{

	}else{
		data, err := reader.ReadString('\n')
		msg = data
		if err != nil {
			msg = ""
			fmt.Println(err)
		}
	}
	return msg, err
}

func fromBytes(b []byte) (int32, error) {
	buf := bytes.NewReader(b)
	var result int32
	err := binary.Read(buf, binary.BigEndian, &result)
	return result, err
}

// To convert an int32 to a 4 byte Big Endian binary format
func toBytes(i int32) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, i)
	return buf.Bytes(), err
}