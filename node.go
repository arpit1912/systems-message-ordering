// socket-server project main.go
package main
import (
	"fmt"
	"net"
	"sync"
	"os"
	"time"
	"math/rand"
	"strconv"
)
const (
        SERVER_HOST = "localhost"
        SERVER_TYPE = "tcp"
)

type SafeConnections struct {
	mu sync.Mutex
	clock int
	all_conn map[string] net.Conn
}


var safeConnections SafeConnections

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
		buffer := make([]byte, 1024)
		mLen, err := client_conn.Read(buffer)
		if err != nil {
				fmt.Println("Error reading:", err.Error())
		}
		id:= string(buffer[:mLen])
		fmt.Println("Received Connection Request From ", id)
		safeConnections.mu.Lock()
		safeConnections.all_conn[id] = client_conn
		safeConnections.mu.Unlock()
		go listenClient(client_conn, id)

	}

}

func listenClient(connection net.Conn, id string) {
	for {
		buffer := make([]byte, 1024)
		mLen, err := connection.Read(buffer)
		if err != nil {
				fmt.Println("Error reading:", err.Error())
				delete(safeConnections.all_conn, id);
				break
		}
		fmt.Println("MSG RECEIVED:= ", string(buffer[:mLen]), " ", strconv.Itoa(safeConnections.clock))
		safeConnections.clock++;
	}
	
}

func establishConnections(wg *sync.WaitGroup, clients_port []string, my_port string) {
	defer wg.Done()
	fmt.Println("TRYING TO ESTABLISH CONNECTIONS")

	for {
		no_done:= 0
		for _, port := range(clients_port) {
			safeConnections.mu.Lock()
			_, ok := safeConnections.all_conn[port]
			safeConnections.mu.Unlock()
			if ok {
				fmt.Println("Already a connection is present to - ", port)
				no_done ++
			} else {
				conn, err := net.Dial(SERVER_TYPE, SERVER_HOST+":"+port)
				if err != nil {
					fmt.Println("Error occured in connection with port: ", port)
				} else {
					safeConnections.mu.Lock()
					safeConnections.all_conn[port] = conn
					safeConnections.mu.Unlock()
					_, err = conn.Write([]byte(my_port)) 
					go listenClient(conn, port)
					if err != nil {
						panic("Error sending message ;( ")
					}
				}
			}
			
			
		}
		if no_done == len(clients_port) {
			fmt.Println("All Connections Ready")
			break
		}	
	}
	
}

func BroadCastMessage(wg *sync.WaitGroup, my_port string) {
	// defer connection.Close()
	defer wg.Done()
	fmt.Println("TRYING TO BROADCAST")
	for {
		for v, conn := range safeConnections.all_conn {
			fmt.Println("Sending Message to -" , v, " ", strconv.Itoa(safeConnections.clock))
			safeConnections.clock++;
			msg := "MSG FROM - " + my_port
			_, err := conn.Write([]byte(msg))
			if err != nil {
				panic("Error sending message ;( ")
			}
		}
		r := rand.Intn(10) + 3
		time.Sleep(time.Duration(r) * time.Second)		
	}
	
}

func main() {
	var wg sync.WaitGroup
	wg.Add(3)
	safeConnections = SafeConnections{all_conn : make(map[string] net.Conn), clock : 0}
	
	go RecieveMessage(&wg, os.Args[1])
	time.Sleep(5*time.Second)
	go establishConnections(&wg, os.Args[2:], os.Args[1])
	time.Sleep(5*time.Second)
	go BroadCastMessage(&wg, os.Args[1])
	wg.Wait()
}