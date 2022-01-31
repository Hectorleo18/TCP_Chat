package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"log"
	"bufio"
	"errors"
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

/*
	Escribir un comando/mensaje
*/
func writeInput(conn *net.TCPConn){
	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		command := strings.Split(text," ")
		//Revisar si el comando es /file para leer el archivo.
		if command[0] == "/file" {
			//Eliminar el primer elemento del sice
			copy(command[0:],command[1:])
			command[len(command)-1] = ""
			command = command[:len(command)-1]
			//Se convierte en cadena y se elimina el salto de línea
			completeCommand := strings.Join(command," ")
			completeCommand = strings.Trim(completeCommand, "\r\n")
			fileInfo, err := os.Stat(completeCommand)
			if err != nil {
				fmt.Println(err)
				return
			}
			//Se obtiene el nombre del archivo
			fileName := fileInfo.Name()
			fmt.Println("Nombre: "+fileName)
			//Se envía el commando y el nombre del archivo
			conn.Write([]byte("/file "+fileName+" "))
			SendFile(completeCommand, conn)
		}else{
			//Si el usuario manda otro comando que no sea file
			err = writeMsg(conn, text)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

/*
	Imprimir el mensaje recibido
*/
func printOutput(conn *net.TCPConn) {
	for {
		msg, err := readMsg(conn)
		// Si se recibe EOF la conección fue cerrada
		if err == io.EOF {
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

/*
	Envía un archivo al servidor
*/
func SendFile(filePath string, conn net.Conn) {
	//Se obtiene el contenido del archivo
	f, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	//Se convierte el contenido en texto
	text := string(f[:])
	//Se reemplaza el salto de línea por una cadena.
	text = strings.ReplaceAll(text,"\n","HectorLeoRodriguez")
	//Se manda el archivo al servidor
	conn.Write([]byte(text))
	conn.Write([]byte("\n"))
}

/*
	Escribe un mensaje de texto en el servidor
*/
func writeMsg(conn net.Conn, msg string) error {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		return err
	}
	return nil
}

/*
	Lee la información proveniente del servidor
*/
func readMsg(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	//Variables a ser retornadas
	var msg string
	var err error
	//Se lee la primer palabra enviada desde el servidor
	typeMsg, err := reader.ReadString(' ')
	if err != nil {
		fmt.Println(err)
	}
	//Si se envía un archivo
	if typeMsg == "file "{
		//Se lee hasta que se encuentra un salto de línea
		data, err := reader.ReadString('\n')
		data = strings.Trim(data, "\r\n")
		//Se convierte la cadena en []string
		args := strings.Split(data, " ")
		//Se crea el archivo en blanco
		f, err := os.Create(args[2])
		if err != nil {
			return "", err
		}
		//Se cierra el archivo en caso de que la aplicación crashee
		defer f.Close()
		//Se crea una cadena sin las primeras 3 palabras
		text := strings.Join(args[3:], " ")
		//Se reemplaza la cadena con salto de línea
		text = strings.ReplaceAll(text, "HectorLeoRodriguez", "\n")
		//Se escribe la información en el archivo
		srcFile := []byte(text)
		f.Write(srcFile)
		f.Close()
		msg = strings.Join(args[:3]," ")
	}else{
		//Si se envía un mensaje de texto
		if typeMsg == "msg "{
			//Se lee hasta el salto de línea
			data, err := reader.ReadString('\n')
			msg = data
			if err != nil {
				msg = ""
				fmt.Println(err)
			}
		}else{
			//Si se recibe un formato distinto
			msg = "Formato no definido."
			err = errors.New("not defined format")
		}
	}
	return msg, err
}
