/*
Aquest codi s'encarrega de administar el servidor i la seva interfície
*/
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	//Imports dels sockets de la llibreria websocket de gorilla
	"github.com/gorilla/websocket"
)

// Tipus de missatges tant de control com de text
const MESSAGE_NEW_USER = "Nuevo usuario"
const MESSAGE_CHAT = "Nuevo mensaje"
const MESSAGE_LEAVE = "Mensaje salir"
const MESSAGE_USER_LIST = "Llista usaris"
const MESSAGE_USER_ALREADY_EXIST = "Usuari ja existeix"
const MESSAGE_NOTFY_USER_CHANGE = "Canvi de usuari"

// Colors del servidor per fer distingible la info que proporciona
var control = "\033[33m"
var chat_message = "\033[97m"
var control_leave = "\033[31m"
var usuari = "\033[34m"
var opertura_servidor = "\033[32m"

//Llistat de conexions al servidor, websockets connectats
var connections = make([]*WebSocketConnection, 0)

//Declaracions de la estructura del SocketPayload
type SocketPayload struct {
	Message string
}

/*
Estructura del Socket de resposta format per:
	From 	---> procedencia del missatge, persona que l'ha enviat
	Type 	---> tipus de missatge, control i text
	Message ---> missatge que es vol enviar
*/
type SocketResponse struct {
	From    string
	Type    string
	Message string
}

//Composició del socket formada per el socket i un ID que es el nom del usuari
type WebSocketConnection struct {
	*websocket.Conn
	Username string
}

/*
Funció encarregada de començar el programa executar tots els HTTP.handle(net/http library)
encarregada de administrar les connexions client-servidor.
Cada petició de connexió localhost:8080 la administra la vincula al socket i
executa el seu fil handleIO
*/
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		content, err := ioutil.ReadFile("index.html")
		if err != nil {
			http.Error(w, "No s'ha pogut obrir el fitxer html", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%s", content)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		currentConnUpgraded, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
		if err != nil {
			http.Error(w, "No s'ha pogut establir connexió", http.StatusBadRequest)
		}

		username := r.URL.Query().Get("username")
		currentConn := WebSocketConnection{Conn: currentConnUpgraded, Username: username}
		connections = append(connections, &currentConn)

		go handleIO(&currentConn, connections)
	})

	fmt.Println(opertura_servidor + "****************Servidor començant al port: 8080****************")

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))

	http.ListenAndServe(":8080", nil)

}

/*
Fil encarregat de administrar la connexió de cada usuari, on s'escolta constantment el socket
en busca de nous missatges.
*/
func handleIO(currentConn *WebSocketConnection, connections []*WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("ERROR", fmt.Sprintf("%v", r))
		}
	}()
	broadcastMessage(currentConn, MESSAGE_NEW_USER, "")
	//enviem llista de usuaris connectats
	enviarllistaUsuaris(currentConn)

	comprovarusuariexistent(currentConn)

	//bucle constant de escolta al socket
	for {
		payload := SocketPayload{}
		err := currentConn.ReadJSON(&payload)
		if err != nil {
			if strings.Contains(err.Error(), "websocket: close") {
				broadcastMessage(currentConn, MESSAGE_LEAVE, "")
				//expulsió de la llista de usuaris connectats
				ejectConnection(currentConn)
				return
			}
			log.Println("ERROR", err.Error())
			continue
		}

		if strings.Contains(payload.Message, "/)(") {
			var nom_vell string = currentConn.Username
			currentConn.Username = strings.TrimPrefix(payload.Message, "/)(")
			broadcastMessage(currentConn, MESSAGE_NOTFY_USER_CHANGE, nom_vell)
			comprovarusuariexistent(currentConn)
		} else {
			broadcastMessage(currentConn, MESSAGE_CHAT, payload.Message)
		}
	}
}

/*
Aquesta funció s'encarrega de comprovar si el codi existeix o no, en cas
de que ja existeixi el nom envia un missatge al client diguent que ha de canviar de nom
*/
func comprovarusuariexistent(currentConn *WebSocketConnection) {
	var i int = 0
	for _, eachConn := range connections {
		if eachConn.Username == currentConn.Username {
			i++
			if i == 2 {
				currentConn.WriteJSON(SocketResponse{
					From:    currentConn.Username,
					Type:    MESSAGE_USER_ALREADY_EXIST,
					Message: " ",
				})
				MissatgeServidor(currentConn, MESSAGE_USER_ALREADY_EXIST, "")
			}

		}
	}
}

/*
Enviem la llista de usuaris al nou usuari connectat, i passem la llista com a missatge
*/
func enviarllistaUsuaris(currentConn *WebSocketConnection) {
	currentConn.WriteJSON(SocketResponse{
		From:    currentConn.Username,
		Type:    MESSAGE_USER_LIST,
		Message: llistaUsuaris(),
	})
}

/*
Creem un string de usuaris separats per ";" amb format: ";Josep;Pep;Roger"
*/
func llistaUsuaris() string {
	var str string
	for _, eachConn := range connections {
		str += ";"
		str += eachConn.Username
	}
	return str
}

/*
Treiem persona que ha sortit de la llista de usuaris, per tal de quan es connecti una
persona nova no se li envii la llista de usuaris actualitzada
*/
func ejectConnection(currentConn *WebSocketConnection) {
	var i int = 0
	var connections_new = make([]*WebSocketConnection, 0)
	for _, eachConn := range connections {
		if eachConn == currentConn {
			continue
		}
		connections_new = append(connections_new, eachConn)
		i++
	}
	connections = connections_new
}

/*
La funció envia el missatge a tots els sockets de la llista que te el servidor,
aquesta funció crea el fromat del missatge avans de enviar ja que varia el tipus de missatge
i el camp de dades en sí
*/
func broadcastMessage(currentConn *WebSocketConnection, kind, message string) {
	for _, eachConn := range connections {
		if eachConn == currentConn {
			MissatgeServidor(currentConn, kind, message)
			continue
		}

		eachConn.WriteJSON(SocketResponse{
			From:    currentConn.Username,
			Type:    kind,
			Message: message,
		})
	}
}

/*
Aquesta funció s'encarrega de imprimir missatges de control a la consola del servidor, escriu
tres missatges diferents que son:
	Missatge de dades 			---> Missatge de les dades i que les ha enviat
	Missatge nou usuari 		---> Missatge de INFO de un nou usuari ha entrat a la conversa
	Missatge usuari ha marchat	---> Missatge de INFO que un usuari ha marchat de la conversa
Aquests es distingeixen segons el tipus que s'envia a la funció
*/
func MissatgeServidor(currentConn *WebSocketConnection, kind, message string) {
	if kind == MESSAGE_CHAT {

		fmt.Println(chat_message + "Missatge ' " + message + " ' enviat per el usuari: " + usuari + currentConn.Username)

	} else if kind == MESSAGE_LEAVE {

		fmt.Println(control_leave + "-----------El Usuari ' " + currentConn.Username + " ' ha sortit de la conversa-----------")

	} else if kind == MESSAGE_NEW_USER {

		fmt.Println(control + "-----------El Usuari ' " + currentConn.Username + " ' ha entrat a la conversa-----------")

	} else if kind == MESSAGE_USER_ALREADY_EXIST {

		fmt.Println(control + "-----------Un usuari ha canviat el seu nom-----------")

	}
}
