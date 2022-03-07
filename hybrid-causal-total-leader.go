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

type Node struct {
	server_port string
	mu sync.Mutex
	seq_number int
	all_conn map[string] net.Conn
}
func (node *Node) ClockIndex(port string) int {
	val, _ := strconv.Atoi(port)
	return val - 80
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

func delayAgent(min int, max int) {
	r := rand.Intn(max-min) + min
	time.Sleep(time.Duration(r) * time.Second)
}

func (node *Node) listenClient(connection net.Conn, id string) {
	for {
		buffer := make([]byte, 1024)
		mLen, err := connection.Read(buffer)
		if err != nil {
				fmt.Println("Error reading:", err.Error())
				delete(node.all_conn, id);
				break
		}
		node.mu.Lock()
		node.seq_number++;
		msg:= "Seq number: " + strconv.Itoa(node.seq_number) + " : " + string(buffer[:mLen])
		fmt.Println("Message Received : ", string(buffer[:mLen]), " Global Seq Number: ", node.seq_number)
		node.mu.Unlock()
		go node.BroadCastMessage(msg)
	}
	
}

func (node *Node) BroadCastMessage(message string) {
	fmt.Println("TRYING TO BROADCAST")
	for port, conn := range node.all_conn {
		fmt.Println("Sending Message to - " , port, " Seq_Number: ", node.seq_number)
		go node.SendMessage(conn, message)
	}	
	
}

func (node *Node) SendMessage(conn net.Conn, message string) {
	delayAgent(10,15)
	_, err := conn.Write([]byte(message))
	if err != nil {
		panic("Error sending message ;( ")
	}
}
func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	node := Node{all_conn : make(map[string] net.Conn),  server_port : os.Args[1]}
	go node.RecieveMessage(&wg, os.Args[1])
	wg.Wait()
}