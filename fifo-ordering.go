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
	"strings"
)
const (
        SERVER_HOST = "localhost"
        SERVER_TYPE = "tcp"
)

type Node struct {
	server_port string
	mu sync.Mutex
	clock []int
	all_conn map[string] net.Conn
	message_queue []string
}
// MSG RECEIVED:=  MSG FROM - 81 : Seq Number : 1
func (node *Node) ClockIndex(port string) int {
	val, _ := strconv.Atoi(port)
	return val - 80
}

func parseMessage(message string) (int, int) {
	result := strings.Split(message, ":")
	port, _ := strconv.Atoi(strings.Trim(result[1], " "))
	timestamp, _ := strconv.Atoi(strings.Trim(result[3], " "))
	return port, timestamp
}

func (node *Node) RecieveMessage (wg *sync.WaitGroup, port string) {
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
		node.mu.Lock()
		node.all_conn[id] = client_conn
		node.mu.Unlock()
		go node.listenClient(client_conn, id)

	}

}

func delayTime(min int, max int) {
	r := rand.Intn(max-min) + min
	time.Sleep(time.Duration(r) * time.Second)
}

func (node *Node) listenClient(connection net.Conn, id string) {
	for {
		buffer := make([]byte, 1024)
		mLen, err := connection.Read(buffer)
		delayTime(0,10)

		if err != nil {
				fmt.Println("Error reading:", err.Error())
				delete(node.all_conn, id);
				break
		}
		
		_, timestamp := parseMessage(string(buffer[:mLen]))
		server_timestamp := node.clock[node.ClockIndex(id)]
		if( server_timestamp + 1 != timestamp) {
			node.message_queue = append(node.message_queue, string(buffer[:mLen]))
			fmt.Println("Storing in local queue:= ", string(buffer[:mLen]), " Server Clock: ", node.clock)
		} else {
			fmt.Println("Message recieved:= ", string(buffer[:mLen]), " Server Clock: ", node.clock)
			node.clock[node.ClockIndex(id)]++;
			var temp_queue []string
			for _, message := range node.message_queue {
				client_port, timestamp := parseMessage(message)
				if(node.clock[client_port-80]+1 != timestamp){
					temp_queue = append(temp_queue, message)
					fmt.Println("Again storing in queue : ", message, " Server Clock: ", node.clock)
				} else {
					fmt.Println("Removing from queue : ", message, " Server Clock: ", node.clock)
				}
			}
			copy(node.message_queue, temp_queue)
		}

	}
	
}

func (node *Node) establishConnections(wg *sync.WaitGroup, clients_port []string, my_port string) {
	fmt.Println("TRYING TO ESTABLISH CONNECTIONS")
	for {
		no_done:= 0
		for _, port := range(clients_port) {
			node.mu.Lock()
			_, ok := node.all_conn[port]
			node.mu.Unlock()
			if ok {
				fmt.Println("Already a connection is present to - ", port)
				no_done ++
			} else {
				conn, err := net.Dial(SERVER_TYPE, SERVER_HOST+":"+port)
				if err != nil {
					fmt.Println("Error occured in connection with port: ", port)
				} else {
					node.mu.Lock()
					node.all_conn[port] = conn
					node.mu.Unlock()
					_, err = conn.Write([]byte(my_port)) 
					go node.listenClient(conn, port)
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

func (node *Node) BroadCastMessage(wg *sync.WaitGroup, my_port string) {
	// defer connection.Close()
	defer wg.Done()
	fmt.Println("TRYING TO BROADCAST")
	for i:=0;i<10;i++ {
		node.clock[node.ClockIndex(my_port)]++;
		for v, conn := range node.all_conn {
			fmt.Println("Sending Message to - " , v, " Server Clock: ", node.clock)
			clock := node.clock[node.ClockIndex(my_port)]
			msg := "MSG FROM : " + my_port + ": Seq Number : " + strconv.Itoa(clock)
			_, err := conn.Write([]byte(msg))
			if err != nil {
				panic("Error sending message ;( ")
			}
		}
		delayTime(0,2)		
	}
	
}

func main() {
	
	var wg sync.WaitGroup
	wg.Add(2)
	node := Node{all_conn : make(map[string] net.Conn), clock : make([]int, len(os.Args[1:])), server_port : os.Args[1]}
	fmt.Println(node.clock)
	go node.RecieveMessage(&wg, os.Args[1])
	node.establishConnections(&wg, os.Args[2:], os.Args[1])
	go node.BroadCastMessage(&wg, os.Args[1])

	wg.Wait()
}