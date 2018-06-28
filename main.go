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

func formatBool(b bool) string {
	if b {
		return "1"
	} else {
		return "0"
	}
}

type car struct {
	RIGHT_M  atomic.Value
	LEFT_M   atomic.Value
	LATANCY  atomic.Value
	OVERRIDE atomic.Value
}

func (c *car) init() {
	c.LATANCY.Store(string("N/A"))
	c.LEFT_M.Store(0)
	c.RIGHT_M.Store(0)
	c.OVERRIDE.Store(false)
}

var (
	car0_direct car
	car0        car
	car1_direct car
	car1        car

	direct bool
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

	spStr := strings.Split(string(r), ",")

	tmp := make([]byte, 32)

	if spStr[0] == "C0" {
		if direct {
			copy(tmp[:], []byte(strconv.Itoa(car0_direct.LEFT_M.Load().(int))+","+strconv.Itoa(car0_direct.RIGHT_M.Load().(int))+","+formatBool(car0_direct.OVERRIDE.Load().(bool))))
		} else {
			copy(tmp[:], []byte(strconv.Itoa(car0.LEFT_M.Load().(int))+","+strconv.Itoa(car0.RIGHT_M.Load().(int))+","+formatBool(car0.OVERRIDE.Load().(bool))))
		}
	}

	if spStr[0] == "C1" {
		if direct {
			copy(tmp[:], []byte(strconv.Itoa(car1_direct.LEFT_M.Load().(int))+","+strconv.Itoa(car1_direct.RIGHT_M.Load().(int))+","+formatBool(car1_direct.OVERRIDE.Load().(bool))))
		} else {
			copy(tmp[:], []byte(strconv.Itoa(car1.LEFT_M.Load().(int))+","+strconv.Itoa(car1.RIGHT_M.Load().(int))+","+formatBool(car0.OVERRIDE.Load().(bool))))
		}
	}
	conn.WriteToUDP(tmp, addr)

	car0_direct.LATANCY.Store(spStr[len(spStr)-1])

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

	car0_direct.init()
	car0.init()
	car1_direct.init()
	car1.init()

	direct = true

	port := 3322

	app := iris.New()
	setupWebsocket(app)

	app.Get("/", func(ctx iris.Context) {
		ctx.Writef(`
          <html> 
            <body>
            <form method='post'>
            <table>
            <tr>
            <td>LEFT0</td>
            <td>RIGHT0</td>
            <td>LEFT1</td>
            <td>RIGHT1</td>
            </tr>
            <tr>
            <td>%d Power</td>
            <td>%d Power</td>
            <td>%d Power</td>
            <td>%d Power</td>
            </tr>
            <tr>
            <td><input type='text' name='LEFT0' value='%d'></td>
            <td><input type='text' name='RIGHT0' value='%d'></td>
            <td><input type='text' name='LEFT1' value='%d'></td>
            <td><input type='text' name='RIGHT1' value='%d'></td>
            </tr>
            </table>
            <button type='submit' name="sub" value="0">Send Command</button>
            <button type='submit' name="sub" value="1">Send global STOP</button>
            </form>
            Round Trip Latancy: <pre id="output">%ss</pre>
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
          </html>`, car0_direct.LEFT_M.Load().(int), car0_direct.RIGHT_M.Load().(int),
			car1_direct.LEFT_M.Load().(int), car1_direct.RIGHT_M.Load().(int),
			car0_direct.LEFT_M.Load().(int), car0_direct.RIGHT_M.Load().(int),
			car1_direct.LEFT_M.Load().(int), car1_direct.RIGHT_M.Load().(int), car0_direct.LATANCY.Load().(string))
	})

	app.Post("/", func(ctx iris.Context) {

		if sub, _ := ctx.PostValueInt("sub"); sub == 0 {
			LEFT_M_tmp, _ := ctx.PostValueInt("LEFT0")
			car0_direct.LEFT_M.Store(LEFT_M_tmp)
			LEFT_M_tmp, _ = ctx.PostValueInt("LEFT1")
			car1_direct.LEFT_M.Store(LEFT_M_tmp)

			RIGHT_M_tmp, _ := ctx.PostValueInt("RIGHT0")
			car0_direct.RIGHT_M.Store(RIGHT_M_tmp)
			RIGHT_M_tmp, _ = ctx.PostValueInt("RIGHT1")
			car1_direct.RIGHT_M.Store(RIGHT_M_tmp)

		} else if sub, _ := ctx.PostValueInt("sub"); sub == 1 {

			car0_direct.LEFT_M.Store(0)
			car1_direct.LEFT_M.Store(0)

			car0_direct.RIGHT_M.Store(0)
			car1_direct.RIGHT_M.Store(0)

		}

		ctx.Redirect("/", 302)
	})
	go app.Run(iris.Addr(":25565"))

	SocketServer(port)
}
