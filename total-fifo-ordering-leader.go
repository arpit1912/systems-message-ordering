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
	seq_number int
	clock []int
	message_queue map[string] string
	all_conn map[string] net.Conn
}
func (node *Node) ClockIndex(port string) int {
	val, _ := strconv.Atoi(port)
	return val - 8081
}

func generate_hash(message string) (string) {
	result := strings.Split(message[:(len(message)-1)], ":")
	port := strings.Trim(result[1], " ")
	msg_id := strings.Trim(result[3], " ")
	return port+"-"+msg_id
}

func parseMessage(message string) (string, int) {
	result := strings.Split(message, ":")
	port := strings.Trim(result[1], " ")
	msg_id,_ := strconv.Atoi(strings.Trim(result[3], " "))
	return port,msg_id
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
		messages := strings.Split(string(buffer[:mLen]),";")
		fmt.Println("messages in buffer: ",len(messages))
		for _, message := range(messages) {
			if message == ""{
				continue
			}
			_, timestamp := parseMessage(message)
			message = message + ";"
			server_timestamp := node.clock[node.ClockIndex(id)]
			if(server_timestamp+1 != timestamp){
				node.message_queue[generate_hash(message)] = message
				fmt.Println("Storing in local queue:= ", message, " Server Clock: ", node.clock)
			} else {
				node.mu.Lock()
				node.clock[node.ClockIndex(id)]++;
				node.seq_number++;
				msg:= "Seq number: " + strconv.Itoa(node.seq_number) + " : " + string(message)
				fmt.Println("Broadcasting Messsage : ", message, " Server Clock: ", node.clock, " Global Seq Number: ", node.seq_number)
				node.mu.Unlock()
				go node.BroadCastMessage(msg)
				for {
					hash := id + "-" + strconv.Itoa(node.clock[node.ClockIndex(id)]+1)
					if local_message, found := node.message_queue[hash]; found {
						delete(node.message_queue, hash)

						node.mu.Lock()
						node.clock[node.ClockIndex(id)]++;
						node.seq_number++;
						msg:= "Seq number: " + strconv.Itoa(node.seq_number) + " : " + string(local_message)
						fmt.Println("Broadcasting Messsage from queue: ", local_message, " Server Clock: ", node.clock, " Global Seq Number: ", node.seq_number)
						node.mu.Unlock()
						go node.BroadCastMessage(msg)

					} else {
						break
					}
				}
			}
		}
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
	delayAgent(11,12)
	_, err := conn.Write([]byte(message))
	if err != nil {
		panic("Error sending message ;( ")
	}
}
func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	number_of_clients,_ := strconv.Atoi(os.Args[2])
	node := Node{all_conn : make(map[string] net.Conn),  server_port : os.Args[1], clock: make([]int, number_of_clients), message_queue: make(map[string] string)}
	go node.RecieveMessage(&wg, os.Args[1])
	wg.Wait()
}