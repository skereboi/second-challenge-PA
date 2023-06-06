/* File: Cliente IRC
 * Autor: Luis Jesus Morales Juarez - A01703455
 *
 * Descripcion: Cliente IRC, gestiona la comunicacion con el servidor
 * Argumentos:
 *		-host: Direccion IP del servidor
 *		-port: Puerto del servidor
 *		-user: Nombre de usuario
 *
 * Uso: go run client_irc.go -host localhost -port 9001 -user user2
 */

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

// Definición de las los argumentos de la línea de comandos.
var (
	host = flag.String("host", "localhost", "server host")
	port = flag.String("port", "9000", "server port")
	user = flag.String("user", "", "username for the client")
)

func main() {

	// Extracción de los argumentos.
	flag.Parse()

	// Creación de la dirección del servidor.
	serverAddress := *host + ":" + *port

	// Intento de conexión al servidor.
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		log.Fatalf("Failed to connect to the server 😓: %v", err)
	}

	fmt.Println("irc-server > Welcome to the Simple IRC Server")

	// Si no se proporcionó un nombre de usuario, se solicita al usuario que ingrese uno.
	if *user == "" {
		fmt.Print("irc-server > Enter your username: ")
		*user, err = bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Fatalf("irc-server > Failed to read username: %v", err)
		}
		*user = strings.TrimSpace(*user)
	}

	// Envío del nombre de usuario al servidor.
	_, err = fmt.Fprintln(conn, *user)
	if err != nil {
		log.Fatalf("irc-server > Failed to send username to the server: %v", err)
	}

	ch := make(chan struct{})

	// Goroutine para leer los mensajes del servidor.
	go func() {
		reader := bufio.NewReader(conn)
		for {
			msg, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					fmt.Println("irc-server > The connection is closed.")
					break
				} else {
					log.Printf("irc-server > Failed to read from the server: %v", err)
					break
				}
			}
			// Se imprime el mensaje recibido.
			fmt.Print(msg)
		}

		ch <- struct{}{}
	}()

	// Bucle principal para leer la entrada del usuario y enviarla al servidor.
	for {
		// Se lee la entrada del usuario.
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')

		if err != nil {
			log.Printf("irc-server > Failed to read input: %v", err)
			break
		}

		// Se escrible la salida al servidor.
		_, err = conn.Write([]byte(input + "\n"))
		if err != nil {
			log.Printf("irc-server > Failed to send input to the server: %v", err)
			break
		}
	}

	// Se espera a que la goroutine termine.
	<-ch
	// Se cierra la conexión.
	os.Exit(0)

}
