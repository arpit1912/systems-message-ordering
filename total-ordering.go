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
	msg_id int
	seq_number int
	all_conn map[string] net.Conn
	message_queue map[string] string
	leader_message_queue map[int] string
	leader_port string
}

func (node *Node) ClockIndex(port string) int {
	val, _ := strconv.Atoi(port)
	return val - 80
}

func parseMessage(message string) (string, string, int) {
	result := strings.Split(message, ":")
	seq_number, _ := strconv.Atoi(strings.Trim(result[1], " "))
	port := strings.Trim(result[3], " ")
	msg_id := strings.Trim(result[5], " ")
	return port+"-"+ msg_id, port, seq_number
}

func generate_hash(message string) (string) {
	result := strings.Split(message, ":")
	port := strings.Trim(result[1], " ")
	msg_id := strings.Trim(result[3], " ")
	return port+"-"+msg_id
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
			if id == node.leader_port {
				fmt.Println("Message received from leader: ", message, ". Local Seq Number: ", node.seq_number)
				hash_key, port, seq_number := parseMessage(message)
				if( node.seq_number + 1 == seq_number){
					if port != node.server_port {
						fmt.Println("Message Delivered:= ", node.message_queue[hash_key])
						delete(node.message_queue, hash_key)
					}
					node.seq_number++;
					for {
						if hash_key, found := node.leader_message_queue[node.seq_number + 1]; found {
							fmt.Println("Message Delivered := ", node.message_queue[hash_key])
							delete(node.message_queue, hash_key)
							delete(node.leader_message_queue,node.seq_number + 1)
							node.seq_number++;
						} else {
							break
						}
					}
				} else {
					fmt.Println("Storing in local leader message queue")
					node.leader_message_queue[seq_number] = hash_key
				}
	
			} else {
				fmt.Println("Storing in local queue:= ", message, " Local Seq Number: ", node.seq_number)
				hash_key := generate_hash(message)
				node.message_queue[hash_key] =  message
			}
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
	defer wg.Done()
	fmt.Println("TRYING TO BROADCAST")
	for i:=0;i<10;i++ {
		node.msg_id++;
		for port, conn := range node.all_conn {
			fmt.Println("Sending Message to - " , port, " Msg_ID: ", node.msg_id)
			msg := "MSG FROM : " + my_port + ": Message ID : " + strconv.Itoa(node.msg_id) + ";"
			go node.SendMessage(conn, msg, port)
		}
		delayAgent(5,10)		
	}
	
}

func (node *Node) SendMessage(conn net.Conn, message string, port string) {
	delayAgent(3,10)	
	_, err := conn.Write([]byte(message))
	if err != nil {
		panic("Error sending message ;( ")
	}
}
func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	node := Node{all_conn : make(map[string] net.Conn),message_queue : make(map[string] string),leader_message_queue : make(map[int] string), server_port : os.Args[1], leader_port : os.Args[2]}
	go node.RecieveMessage(&wg, os.Args[1])
	node.establishConnections(&wg, os.Args[2:], os.Args[1])
	go node.BroadCastMessage(&wg, os.Args[1])

	wg.Wait()
}