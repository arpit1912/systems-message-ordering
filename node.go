// socket-server project main.go
package main
import (
	"fmt"
	"net"
	"sync"
	"os"
	"time"
	"math/rand"
)
const (
        SERVER_HOST = "localhost"
        SERVER_TYPE = "tcp"
)

func RecieveMessage (wg *sync.WaitGroup, port string) {
	defer wg.Done()
	fmt.Println("Searching for available port...")
	conn, err := net.Listen(SERVER_TYPE, SERVER_HOST + ":" + port)

	if err != nil {
		fmt.Println(port, " is not available to listen ")
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("Listening on " + SERVER_HOST + ":" + port)

	for {
		client_conn, err := conn.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
        fmt.Println("Client connected")
		go processClient(client_conn)

	}

}

func processClient(connection net.Conn) {
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
			fmt.Println("Error reading:", err.Error())
	}
	fmt.Println("Received: ", string(buffer[:mLen]))
	connection.Close()
}



func SendMessage(wg *sync.WaitGroup, clients_port []string) {
	defer wg.Done()
	for {
			fmt.Println("Sending message to the other nodes")
			for _, port := range(clients_port) {
				conn, err := net.Dial(SERVER_TYPE, SERVER_HOST+":"+port)
				if err != nil {
						fmt.Println("Error occured in connection: ")
				} else {
					go BroadCastMessage(conn)
				}
				
			}
		r := rand.Intn(5)
		time.Sleep(time.Duration(r) * time.Second)		
	}
}

func BroadCastMessage(connection net.Conn) {
	defer connection.Close()
	_, err := connection.Write([]byte("Second node!!"))
	if err != nil {
		panic("Error sending message ;( ")
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	go RecieveMessage(&wg, os.Args[1])
	time.Sleep(5*time.Second)
	clients_port := os.Args[2:]
	go SendMessage(&wg, clients_port)
	wg.Wait()
}

// // socket-client project main.go
// package main
// import (
//         "fmt"
//         "net"
// )
// const (
//         SERVER_HOST = "localhost"
//         SERVER_PORT = "9988"
//         SERVER_TYPE = "tcp"
// )
// func main() {

// }