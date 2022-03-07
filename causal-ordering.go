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
	clock []int
	all_conn map[string] net.Conn
	message_queue []string
	delivery_queue []string
}
// MSG RECEIVED:=  MSG FROM - 81 : Seq Number : 1
func (node *Node) ClockIndex(port string) int {
	val, _ := strconv.Atoi(port)
	return val - 8080
}

func parseMessage(message string) (int, []int) {
	result := strings.Split(message, ":")
	port, _ := strconv.Atoi(strings.Trim(result[1], " "))
	var vec_clock []int
	err:= json.Unmarshal([]byte(strings.Trim(result[3], " ")), &vec_clock)
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

func delayTime(min int, max int) {
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
		
		_, vec_clock := parseMessage(string(buffer[:mLen]))
		server_timestamp := node.clock[node.ClockIndex(id)]

		is_order_correct := false

		
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
			fmt.Println("Message Delivered:= ", string(buffer[:mLen]), " Server Clock: ", node.clock)
			node.delivery_queue = append(node.delivery_queue, string(buffer[:mLen]))
			node.clock[node.ClockIndex(id)]++;
			var temp_queue []string
			for {
				found_once := false
				for _, message := range node.message_queue {
					client_port, vec_clock := parseMessage(message)
					is_order_correct1 := false
	
					if( node.clock[client_port-8080] + 1 == vec_clock[node.ClockIndex(id)]) {
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
						fmt.Println("Removing from queue. Message Delivered : ", message, " Server Clock: ", node.clock)
						node.delivery_queue = append(node.delivery_queue, message)
						node.clock[client_port-8080]++;
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
		vec_clock, _ := json.Marshal(node.clock)
		for v, conn := range node.all_conn {
			fmt.Println("Sending Message to - " , v, " Server Clock: ", node.clock)
			msg := "MSG FROM : " + my_port + ": Vec Clock : " + string(vec_clock)
			go node.SendMessage(conn, msg)
		}
		delayTime(0,10)		
	}
	delayTime(20,21)
	fmt.Println("PRINTING DELIVERY QUEUE.")
	for i,v := range(node.delivery_queue) {
		fmt.Println(i," - ",v)
	}
}

func (node *Node) SendMessage(conn net.Conn, message string) {
	delayTime(0,10)
	_, err := conn.Write([]byte(message))
	if err != nil {
		panic("Error sending message ;( ")
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