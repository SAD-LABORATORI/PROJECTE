/*
Aquest codi s'encarrega de administrar el client i la seva interficie
*/

//
var app = {}
app.ws = undefined
app.container = undefined
 

//Constants dels diferents tipus de missatges
const MESSAGE_NEW_USER = "Nuevo usuario"
const MESSAGE_CHAT = "Nuevo mensaje"
const MESSAGE_LEAVE = "Mensaje salir"
const MESSAGE_USER_LIST = "Llista usaris"
const MESSAGE_USER_ALREADY_EXIST = "Usuari ja existeix"
const MESSAGE_NOTFY_USER_CHANGE = "Canvi de usuari"
/*
Aquesta funció s'encarrega de administrar els missatges afegintlos al codi HTML 
per tal de que siguin visibles a la pàgina web.
Quan es executada crea un nou element HTML de tipus "p" i dintre de aquest afegeix
el missatge, posteriorment s'afegeix al container on hi han tots el missatges.
*/
app.print = function (message) {
    var el = document.createElement("p")
    el.innerHTML = message
    app.container.append(el)
}

/*
Aquesta funció administra els missatges que nosaltres enviem.
Llegeix el valor del missatge que hem escrit, només agafarà el missatge en cas
de que no estigui vuit, despres crea el format del missatge i crida la funció
app.print afegint el teu missatge al container.
Despres de cada lectura buidem el missatge per no haver de borrar el anterior cada vegada que escrivim
*/
app.doSendMessage = function () {
    var messageRaw = document.querySelector('.input-message').value
    if(document.querySelector('.input-message').value!='') { 
        app.ws.send(JSON.stringify({
            Message: messageRaw
        }))

        var message = '<div class= "propi"><b>me</b>: ' + messageRaw +'</div>'
        app.print(message)
    }
    document.querySelector('.input-message').value = ''
}

/*
Aquesta funció inicialitza el programa i administra el mateix. Al començar
executa el promt que et pregunta el nom, fins que no posis un nom no he deixa continuar.
A continuació crida la funció onmessage que s'encarrega de llegir els diferents missatges
enviats per el servidor distingir-los i fer els canvis corresponents.
*/
app.init = function () {
    if (!(window.WebSocket)) {
        alert('Your browser does not support WebSocket')
        return
    }

    document.querySelector('.input-message').value = ''
    var name = prompt('Introdueix el teu nom Usuari') || "---"

    while(name=="---"){
        name = prompt('Introdueix el teu nom Usuari') || "---"
    }

    app.container = document.querySelector('.container')

    app.ws = new WebSocket("ws://localhost:8080/ws?username=" + name)
    
    app.ws.onopen = function() {
        var message = '<div class="msg_control">Has entrat a la sala</div>'
        app.print(message)
    }

    /*
    Aquesta funció distingeix els tipus de missatge ha rebut, en cas de ser un nou missatge de control de
    tipus afegir o treure usuari s'encarrefa de afegir i treure elements de la llista del menú pushbar.
    En cas de ser un missatge de llista de usuaris creem un bucle separant els ";" de string i creant
    un element de tipus llista "li" per cada un. Si és un missatge normal l'afegeix al contenidor.
    Els missatges de control també els afegeix al contenidor peró aquest amb un classe diferent per 
    donarlis una estètica diferent.
    Tambè implementa el codi capaç de manejar un canvi de nom de usuari.
    */
    app.ws.onmessage = function (event) {
        var res = JSON.parse(event.data)

        var message
        var str
        if (res.Type == MESSAGE_NEW_USER) {
            message = '<div class="msg_control">El usuari <b>' + res.From + '</b> ha entrat a la conversa </div>'
            str = '<li id='+res.From+'>'+res.From+'</li>'
            document.getElementById("menu").insertAdjacentHTML("beforeend", str)
            app.print(message)

        } else if (res.Type == MESSAGE_LEAVE) {
            message = '<div class="msg_control">El usuari <b>' + res.From + '</b> ha marxat de la conversa </div>'
            let li = document.getElementById(res.From)
            document.getElementById("menu").removeChild(li)
            app.print(message)

        }else if (res.Type == MESSAGE_USER_LIST){
            var parts = res.Message.split(";")
            for(var i = 1; i<parts.length; i++){
                str = '<li id='+parts[i]+'>'+parts[i]+'</li>'
                document.getElementById("menu").insertAdjacentHTML("beforeend", str)
            }
            console.log(res.Message)
        }else if (res.Type == MESSAGE_USER_ALREADY_EXIST){
            let li = document.getElementById(res.From)
            document.getElementById("menu").removeChild(li)
            
            var name_nou = prompt('El nom de usuari ja existeix') || "---"

            while(name=="---"){
                name_nou = prompt('El nom de usuari ja existeix') || "---"
            }
            
            str = '<li id='+name_nou+'>'+name_nou+'</li>'
            document.getElementById("menu").insertAdjacentHTML("beforeend", str)

            app.ws.send(JSON.stringify({Message: '/)('+name_nou}))

        }else if (res.Type == MESSAGE_NOTFY_USER_CHANGE){
            let li = document.getElementById(res.Message)
            document.getElementById("menu").removeChild(li)
            str = '<li id='+res.From+'>'+res.From+'</li>'
            document.getElementById("menu").insertAdjacentHTML("beforeend", str)

            message='<div class="msg_control">El usuari <b>' + res.Message + '</b> ha canviat el nom a <b>' + res.From + '</b></div>'
            
            app.print(message)

        }else {
            message = '<div class= "altre_usuari"><b>' + res.From + '</b>: ' + res.Message + '</div>'
            app.print(message)
        }

    }

    app.ws.onclose = function () {
        var message = '<b>me</b>: disconnected'
        app.print(message)
    }
}

window.onload = app.init
