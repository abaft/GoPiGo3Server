package main

import (
	//"bufio"
	"github.com/kataras/iris"
	"github.com/kataras/iris/websocket"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	//"time"
)

var (
	RIGHT_M atomic.Value
	LEFT_M  atomic.Value
	LATANCY  atomic.Value
)

func SocketServer(port int) {

	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	defer ServerConn.Close()
	if err != nil {
		log.Fatalf("Socket listen port %d failed,%s", port, err)
		os.Exit(1)
	}
	log.Printf("Begin listen port: %d", port)

	for {
		buffer := make([]byte, 32)
		_, clientAddr, err := ServerConn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatalln(err)
			continue
		}
		go handler(ServerConn, clientAddr, buffer)
	}

}

func handler(conn *net.UDPConn, addr *net.UDPAddr, r []byte) {

	tmp := make([]byte, 32)
	// SCHEMA 0,0
	copy(tmp[:], []byte(strconv.Itoa(LEFT_M.Load().(int))+","+strconv.Itoa(RIGHT_M.Load().(int))))
	conn.WriteToUDP(tmp, addr)
	spStr := strings.Split(string(r), ",")
        //latancy, _ := strconv.ParseFloat(spStr[len(spStr)-1], 64)
        LATANCY.Store(spStr[len(spStr)-1])

	log.Println("RECEIVED: " + string(r))
	log.Println("SENT: " + string(tmp))
}

func isTransportOver(data string) (over bool) {
	over = strings.HasSuffix(data, "\r\n\r\n")
	return
}

func setupWebsocket(app *iris.Application) {
	ws := websocket.New(websocket.Config{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	})

	app.Get("/echo", ws.Handler())

	app.Any("/iris-ws.js", func(ctx iris.Context) {
		ctx.Write(websocket.ClientSource)
	})
}

func main() {

	port := 3322
	LEFT_M.Store(0)
	RIGHT_M.Store(0)

	app := iris.New()
	setupWebsocket(app)

	app.Get("/", func(ctx iris.Context) {
		ctx.Writef(`
          <html> 
            <body>
            <form method='post'>
            <table>
            <tr>
            <td>LEFT</td>
            <td>RIGHT</td>
            </tr>
            <tr>
            <td>%d Power</td>
            <td>%d Power</td>
            </tr>
            <tr>
            <td><input type='text' name='LEFT' value='%d'></td>
            <td><input type='text' name='RIGHT' value='%d'></td>
            </tr>
            </table>
            <button type='submit'>Send Command</button>
            </form>
            Round Trip Latancy: <pre id="output">%s</pre>
            <script src="/iris-ws.js"></script>
            <script>
    var scheme = document.location.protocol == "https:" ? "wss" : "ws";
    var port = document.location.port ? (":" + document.location.port) : "";
    // see app.Get("/echo", ws.Handler()) on main.go
    var wsURL = scheme + "://" + document.location.hostname + port+"/echo";
var socket = new Ws(wsURL)

    socket.On("latancy", function (msg) {
        document.getElementById("output").innerHTML = msg;
    });
            </script>
            </body>
          </html>`, LEFT_M.Load().(int), RIGHT_M.Load().(int), LEFT_M.Load().(int), RIGHT_M.Load().(int), LATANCY.Load().(string))
	})

	app.Post("/", func(ctx iris.Context) {

		LEFT_M_tmp, _ := ctx.PostValueInt("LEFT")
		LEFT_M.Store(LEFT_M_tmp)

		RIGHT_M_tmp, _ := ctx.PostValueInt("RIGHT")
		RIGHT_M.Store(RIGHT_M_tmp)
		ctx.Redirect("/", 302)
	})
	go app.Run(iris.Addr(":25565"))

	SocketServer(port)
}
