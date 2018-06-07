package main

import (
	//"bufio"
	"github.com/kataras/iris"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	//"time"
)

var (
	RIGHT_M = 0
	LEFT_M  = 0
)

func SocketServer(port int) {

	listen, err := net.Listen("tcp4", ":"+strconv.Itoa(port))
	defer listen.Close()
	if err != nil {
		log.Fatalf("Socket listen port %d failed,%s", port, err)
		os.Exit(1)
	}
	log.Printf("Begin listen port: %d", port)

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatalln(err)
			continue
		}
		go handler(conn)
	}

}

func handler(conn net.Conn) {

	defer conn.Close()

	var (
	//r   = bufio.NewReader(conn)
	//w = bufio.NewWriter(conn)
	)

	tmp := make([]byte, 1024)
	// SCHEMA 0,0
	copy(tmp[:], []byte(strconv.Itoa(LEFT_M)+","+strconv.Itoa(RIGHT_M)))
	log.Println(len(tmp))
	conn.Write([]byte(tmp))
	log.Println("SENT: " + string(tmp))
	buffer := make([]byte, 1024)
	conn.Read(buffer)
	log.Println("RECEIVED: " + string(buffer))
}

func isTransportOver(data string) (over bool) {
	over = strings.HasSuffix(data, "\r\n\r\n")
	return
}

func main() {

	port := 3324

	app := iris.New()
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
            </body>
          </html>`, LEFT_M, RIGHT_M, LEFT_M, RIGHT_M)
	})

	app.Post("/", func(ctx iris.Context) {

		LEFT_M, _ = ctx.PostValueInt("LEFT")
		RIGHT_M, _ = ctx.PostValueInt("RIGHT")
		ctx.Redirect("/", 302)
	})
	go app.Run(iris.Addr(":25565"))

	SocketServer(port)
}
