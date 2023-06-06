/* File: Servidor IRC
 * Autor: Luis Jesus Morales Juarez - A01703455
 *
 * Descripcion: Servidor IRC que gestiona la comunicacion entre los clientes
 * Argumentos:
 *		-host: Direccion IP del servidor
 *		-port: Puerto del servidor
 *
 * Uso: go run server_irc.go -host=<direccion ip> -port=<puerto>
 */

package main

// Importación de paquetesrv.
import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

// - / - / - / - / - / - / - / - / - / - / - / - / - / - / - / - / - / - / -
// Estructuras de datos.

// Client: representa a un cliente conectado al servidor.
type Client struct {
	conn     net.Conn // Conexión del cliente.
	username string   // Nimbre de usuario.
	isAdmin  bool     // Indica si el cliente es el administrador del chat.
}

// Server: representa al servidor.
type Server struct {
	clients map[net.Conn]Client // Indexa a los clientes conectados al servidor.
	join    chan net.Conn       // Canal de 'Registro'
	leave   chan net.Conn       // Canal de 'Desconexion'
	msg     chan string         // Canal de 'Mensajes'
}

// - / - / - / - / - / - / - / - / - / - / - / - / - / - / - / - / - / - / -
// Funciones [Servidor]

// start_server: inicializa el servidor.
func start_server() *Server {
	return &Server{
		clients: make(map[net.Conn]Client),
		join:    make(chan net.Conn),
		leave:   make(chan net.Conn),
		msg:     make(chan string),
	}
}

// run_server: ejecuta el servidor
func (srv *Server) run_server() {

	log.Println("irc-server > Ready for receiving new clients")

	for {
		// Para cada conexion entrante
		select {
		// Selecciona el respectivo canal, y ejecuta el gestor de eventos correspondiente
		case conn := <-srv.join:
			srv.handle_conection(conn) // Registro
		case conn := <-srv.leave:
			srv.close_conection(conn) // Desconexion
		case msg := <-srv.msg:
			srv.handle_message(msg) // Mensajes
		}
	}
}

// - - - - - - - - - - -
// Funciones [Gestion de conexiones]

// handle_conection: gestiona las conexiones entrantes al servidor.
func (srv *Server) handle_conection(conn net.Conn) {

	// scanner: incializa un 'lector' de mensajes
	scanner := bufio.NewScanner(conn)
	// Lee el nombre de usuario del cliente
	scanner.Scan()

	// username: almacena el nombre de usuario del cliente
	username := scanner.Text()

	// client: representa una instancia del cliente
	client := Client{
		conn:     conn,
		username: username,
		isAdmin:  username == "admin",
	}

	// Se agrega el cliente a la lista de conexiones activas
	srv.clients[conn] = client

	// El servidor registra la conexion del cliente
	log.Printf("irc-server > New connected user [%s]\n", username)

	// Si un cliente se regitra con el nombre de usuario 'admin'
	// se le otorga el rol de administrador, se registra en el servidor
	if client.isAdmin {
		log.Printf("irc-server > [%s] was promoted as the channel ADMIN\n", username)
	}

	// El servidor incia una gorutine para gestionar al cliente recien conectado
	go srv.handle_client(client)
}

// close_conection: gestiona el cierre de conexiones al servidor.
func (srv *Server) close_conection(conn net.Conn) {

	// si el cliente se encuetra en la lista de clientes conectados
	client, ok := srv.clients[conn]

	if ok {
		// El servidor registra la desconexion del cliente
		log.Printf("irc-server > [%s] left\n", client.username)
		// Elimina la instacia del cliente
		delete(srv.clients, conn)
	}
	// Y cierra la conexion
	conn.Close()
}

// - - - - - - - - - - -
// Funciones [Gestion de canales]

// handle_message: gestiona los mensajes enviados por los clientes.
func (srv *Server) handle_message(msg string) {

	// para cada cliente conectado al servidor
	for _, client := range srv.clients {

		// El servidor envia el mensaje a cada cliente
		_, err := client.conn.Write([]byte(msg + "\n"))

		// Registra el error en caso de que ocurra
		if err != nil {
			log.Printf("irc-server > Message sending failed ❌: %v\n", err)
		}

	}
}

// handle_client: gestiona los mensajes enviados por los clientes.
func (srv *Server) handle_client(client Client) {

	//Se define un lector de mensajes
	reader := bufio.NewReader(client.conn)

	for {
		// Se lee el mensaje del cliente
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

		msg = strings.TrimSpace(msg) //Formato del mensaje

		if len(msg) == 0 {
			continue // Si el mensaje está vacío se ignora
		}

		// Si el mensaje es un comando
		if strings.HasPrefix(msg, "/") {

			// Si el cliente es administrador
			if client.isAdmin {
				tokens := strings.Split(msg, " ")
				cmd := tokens[0]

				//Lista de comandos completa
				switch cmd {
				case "/users":
					srv.list_users(client)
				case "/msg":
					srv.direct_message(client, tokens)
				case "/time":
					srv.current_time(client)
				case "/user":
					srv.user_info(client, tokens)
				case "/kick":
					srv.kick_user(client, tokens)
				default:
					_, _ = client.conn.Write([]byte("Invalid command ⁉️\n"))
				}
			} else {
				// Si esl cliente es un usuario normal
				tokens := strings.Split(msg, " ")
				cmd := tokens[0]

				switch cmd {
				case "/users":
					srv.list_users(client)
				case "/msg":
					srv.direct_message(client, tokens)
				case "/time":
					srv.current_time(client)
				case "/user":
					srv.user_info(client, tokens)
				default:
					_, _ = client.conn.Write([]byte("Not allowed command, must be admin 👮‍♂️\n"))
				}
			}
		} else {
			// Mensaje global
			srv.msg <- fmt.Sprintf("%s> %s", client.username, msg)

		}
	}
	// Se cierra la conexion
	srv.leave <- client.conn
}

// - - - - - - - - - - - - - - - -
//Funciones [Comandos]

// list_users: lista los usuarios conectados al servidor [comando /users]
func (srv *Server) list_users(client Client) {
	for _, c := range srv.clients {
		_, _ = client.conn.Write([]byte("\t" + c.username + " 👤\n"))
	}
}

// direct_message: envía un mensaje directo a un usuario específico [comando /msg]
func (srv *Server) direct_message(client Client, tokens []string) {

	// /msg <user> <message>
	if len(tokens) < 3 {
		_, _ = client.conn.Write([]byte("irc-server > Please provide a user 👤  and a message 💬\n"))
		return
	}

	// user: nombre de usuario
	username := tokens[1]

	// msg: Mensaje
	msg := strings.Join(tokens[2:], " ")

	// Busqueda del usuario
	for _, c := range srv.clients {
		// Si el usuario existe, se le envía el mensaje privado
		if c.username == username {
			_, _ = c.conn.Write([]byte(fmt.Sprintf("DM from <%s>: \"%s\" 😶‍🌫️\n", client.username, msg)))
			return
		}
	}

	_, _ = client.conn.Write([]byte(fmt.Sprintf("irc-server> The user '%s' does not exist 🫥\n", username)))
}

// current_time: envía al cliente la hora actual del servidor [comando /time]
func (srv *Server) current_time(client Client) {
	_, _ = client.conn.Write([]byte("irc-server> Server time: " + time.Now().String() + "⏰ \n"))
}

// user_info: envía al cliente la información de un usuario específico [comando /user]
func (srv *Server) user_info(client Client, tokens []string) {

	// Comando inválido
	if len(tokens) < 2 {
		_, _ = client.conn.Write([]byte("Please provide a <user> to view 😓\n"))
		return
	}

	// username: Nombre de usuario
	username := tokens[1]
	for _, c := range srv.clients {
		// Si el usuario existe, se despliega su información
		if c.username == username {
			_, _ = client.conn.Write([]byte(fmt.Sprintf("\tUsername: %s 🥸, \tIP: %s 🛜\n", c.username, c.conn.RemoteAddr().String())))
			return
		}
	}

	_, _ = client.conn.Write([]byte(fmt.Sprintf("The <user> %s does not exist 😐\n", username)))
}

// kick_user: expulsa a un usuario del servidor [comando /kick]
func (srv *Server) kick_user(client Client, tokens []string) {
	// Comando inválido
	if len(tokens) < 2 {
		_, _ = client.conn.Write([]byte("Please provide a user to kick 🩴\n"))
		return
	}
	// username: Nombre de usuario
	username := tokens[1]

	// Busqueda del usuario a expulsar
	for conn, c := range srv.clients {

		// Si el usuario existe, se le envía el mensaje y se le expulsa
		if c.username == username {
			_, _ = c.conn.Write([]byte("🚩 You have been kicked from the server 🚩\n"))
			log.Printf("irc-server > [%s] was kicked\n", username)
			srv.leave <- conn
			return
		}
	}

	_, _ = client.conn.Write([]byte(fmt.Sprintf("The <user> %s does not exist 🤠\n", username)))
}

// - / - / - / - / - / - / - / - / - / - / - / - / - / - / - / - / - / - / - / - / -
// Funcion principal

// main: Función principal del servidor
func main() {

	// Manejo de argumentos
	host := flag.String("host", "localhost", "define host for the server")
	port := flag.String("port", "9000", "define port for the server")
	flag.Parse()

	// Inicializacion del servidor
	server := start_server()

	// Ejecucion del sevidor
	go server.run_server()

	// Configuracion del log
	log.SetFlags(log.LstdFlags)
	log.Printf("irc-server > Simple IRC Server started at %s:%s\n", *host, *port)

	// Inicializacion del listener
	listener, _ := net.Listen("tcp", fmt.Sprintf("%s:%s", *host, *port))

	// Cierre del listener
	defer listener.Close()

	// Ciclo infinito para aceptar conexiones
	for {
		conn, _ := listener.Accept()
		server.join <- conn
	}
}
