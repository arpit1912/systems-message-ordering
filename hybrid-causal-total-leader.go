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
	"encoding/json"
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
	all_conn map[string] net.Conn
	message_queue []string
}
func (node *Node) ClockIndex(port string) int {
	val, _ := strconv.Atoi(port)
	return val - 81
}

func parseMessage(message string) (int, []int) {
	result := strings.Split(message, ":")
	port, _ := strconv.Atoi(strings.Trim(result[1], " "))
	var vec_clock []int
	fmt.Println("result5",result[5])
	err:= json.Unmarshal([]byte(strings.Trim(result[5], " ")), &vec_clock)
	if err != nil {
        fmt.Println("ERROR PARSING")
    }
	return port, vec_clock
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
		for _, message := range(messages) {
			if message == ""{
				continue
			}
		
			_, vec_clock := parseMessage(message)
			server_timestamp := node.clock[node.ClockIndex(id)]
			is_order_correct := false

			fmt.Println("Vec Clock", vec_clock, " Server Clock: ", node.clock)



			if( server_timestamp + 1 == vec_clock[node.ClockIndex(id)]) {
				// Message is in FIFO Order
				// Checking for Causal Ordering
				temp:= true
				for i, val := range node.clock {
					if i == node.ClockIndex(id) {
						continue
					}
					if val < vec_clock[i] {
						temp =false
						break
					}

				}
				if temp {
					is_order_correct = true
				}
			}

			if is_order_correct {
				node.mu.Lock()
				node.seq_number++;
				msg:= "Seq number: " + strconv.Itoa(node.seq_number) + " : " + string(buffer[:mLen])
				fmt.Println("Message Received : ", message, " Global Seq Number: ", node.seq_number)
				node.mu.Unlock()
				go node.BroadCastMessage(msg)
				fmt.Println("Message Delivered:= ", message, " Server Clock: ", node.clock)
				node.clock[node.ClockIndex(id)]++;
				var temp_queue []string
				for {
					found_once := false
					for _, message := range node.message_queue {
						client_port, vec_clock := parseMessage(message)
						is_order_correct1 := false
						
						if( node.clock[client_port-80] + 1 == vec_clock[node.ClockIndex(id)]) {
							// Message is in FIFO Order
							// Checking for Causal Ordering
							temp:= true
							for i, val := range node.clock {
								if i == node.ClockIndex(id) {
									continue
								}
								if val < vec_clock[i] {
									temp =false
									break
								}
				
							}
							if temp {
								is_order_correct1 = true
							}
						}
		
		
						if is_order_correct1 {
							found_once = true
							node.mu.Lock()
							node.seq_number++;
							msg:= "Seq number: " + strconv.Itoa(node.seq_number) + " : " + message
							fmt.Println("Message Received : ", message, " Global Seq Number: ", node.seq_number)
							node.mu.Unlock()
							go node.BroadCastMessage(msg)
							fmt.Println("Removing from queue. Message Delivered : ", message, " Server Clock: ", node.clock)
							node.clock[client_port-80]++;
						} else {
							temp_queue = append(temp_queue, message)
							//fmt.Println("Again storing in queue : ", message, " Server Clock: ", node.clock)
						}
					}
					node.message_queue = temp_queue
					temp_queue = nil
					if !found_once {
						break
					}
				}
				

			} else {
				node.message_queue = append(node.message_queue, string(buffer[:mLen]))
				fmt.Println("Storing in local queue:= ", string(buffer[:mLen]), " Server Clock: ", node.clock)
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
	delayAgent(10,15)
	_, err := conn.Write([]byte(message))
	if err != nil {
		panic("Error sending message ;( ")
	}
}
func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	n, _ := strconv.Atoi(os.Args[2])
	node := Node{all_conn : make(map[string] net.Conn),clock: make([]int, n),  server_port : os.Args[1]}
	go node.RecieveMessage(&wg, os.Args[1])
	wg.Wait()
}