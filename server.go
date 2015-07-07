package gpoll

import 	(
	"fmt"
	"net/http"
	"strings"
	"sync"
	"net/url"
	"time"
	"strconv"
	)

// The Broadcaster struct contains configuration information to dictate
// long-polling behavior. The only parameters you have to assign values to are
// RoutineMaxClients, ClientBufferSize, and RoutineBufferSize
type Broadcaster struct {

	// Number of clients per routine
	RoutineMaxClients	int

	// Maximum number of messages per client channel
	ClientBufferSize	int

	// Maximum number of messages to hold in each routine's buffer
	RoutineBufferSize	int

	// Number of milliseconds to wait before broadcasting message in master buffer
	// default is 0, but can be set to compensate for the time it takes for
	// clients to reconnect to the polling service.
	SyncDelayMs	int

	// Routines holds routine references within the Broadcaster to allow proper
	// message handling and client maintenance.
	Routines	[]Routine

	// WG is a WaitGroup for ensuring synchronicity when necessary
	WG	sync.WaitGroup

	// Mux is a mutex for maintaining proper data control and preventing race
	// conditions during asynchronous broadcasting and client management.
	Mux	sync.Mutex

	// IsBenchmarking is used for controlling benchmark behavior
	IsBenchmarking	bool

	// BenchData is an array for holding data during benchmarking. This will
	// always be cleared after benchmarking to save memory.
	BenchData	[]TestStruct

	// Command salt. This will automatically be generated. You can generate one
	// if you like, but make sure it is not publicly known.
	CSalt	string
}

// The Routine struct holds information about routines, as well as the
// the references to clients.
type Routine struct{
	sync.Mutex

	// Id of the routine. This is generated automatically.
	ID 					int

	// Pipe holds the reference for the receiving channel of the routine.
	Pipe 				chan string

	// Clients holds a collection of Client objects for tracking and management.
	Clients 		[]Client
}

// TestStruct holds benchmarking information for tuning your server.
type TestStruct struct {

	// Ns is the time taken to complete a single broadcast in nanoseconds.
	Ns	int64

	// Clients represents the number of total clients being broadcasted to.
	Clients 	int

	// Routines represents the total number of routines being managed by the server.
	Routines 	int
}

// The Client struct holds the client id, and the channel reference for that client.
type Client struct {
	//ID is the UUID of the client.
	ID 		string

	//Pipe is the receiving channel of the client.
	Pipe 	chan string
}

// NewBroadcaster returns a pointer to a Broadcaster with the
// default settings of RoutineMaxClients: 1000, ClientBufferSize: 100,
// RoutineBufferSize: 200, and SyncDelayMs: 0.
func NewBroadcaster() *Broadcaster {
	return 	&Broadcaster 	{
				RoutineMaxClients: 1000,
				ClientBufferSize: 100,
				RoutineBufferSize: 200,
				SyncDelayMs: 0,
				}
}

var DefaultBroadcaster = NewBroadcaster()

// ListenAndBroadcast uses the DefaultBroadcaster to set up a handler for polling requests.
// If supplied with a ServeMux object, it will add the handler to it, otherwise
// it defaults to using the DefaultServeMux object.
func ListenAndBroadcast(path string, srv ...*http.ServeMux){
	if len(srv) > 0 {
		srv[0].HandleFunc(path,DefaultBroadcaster.HandleRequests)
	}	else {
	http.HandleFunc(path,DefaultBroadcaster.HandleRequests)
	}
}

// ListenAndBroadcast uses the associated Broadcaster object to set up a handler for polling requests.
// If supplied with a ServeMux object, it will add the handler to it, otherwise
// it defaults to using the DefaultServeMux object.
func (b *Broadcaster) ListenAndBroadcast(path string, srv ...*http.ServeMux){
	if len(srv) > 0 {
		srv[0].HandleFunc(path,b.HandleRequests)
	}	else {
	http.HandleFunc(path,b.HandleRequests)
	}
}

// HandleRequests takes a typical http request and returns a new client ID if the
// client wasn't previously registered, or a broadcasted message if the client is
// registered.
func (b *Broadcaster) HandleRequests(w http.ResponseWriter, r *http.Request) {
	u,_ := url.ParseQuery(r.URL.RawQuery);
	uid := u.Get("client-id")
	if len(uid) < 10 {
		//Generate new Client
		var tid int
		uid, tid = b.AddClient()
		ret := fmt.Sprintf("{\"client-id\":\"%d$%s\"}",tid,uid)
		w.Write([]byte(ret))
	} else {
		//retrieve client channel and wait for message
		info := strings.Split(uid,"$")
		if len(info) < 2 {
			//send error if bad uid
			w.WriteHeader(http.StatusNotFound)
			return
		}
		tid,_ := strconv.Atoi(info[0])
		q,err := b.GetChannel(info[1],tid)
		if err != nil{
			//send error if client not found
			w.WriteHeader(http.StatusNotFound)
			return
		}
		temp := <-q
		w.Write([]byte(temp))
	}
}

// HandleMessage is the basis of all coroutines within the gpoll package. It contains the code
// that automatically manages client removal and broadcasting. Look at the source to figure out more about this
// function. 
func (b *Broadcaster) HandleMessage(ch <-chan string,in int){
	var trackerIndex int = in-1
	for{
		var start time.Time
		message := <- ch
		if message =="quit"+b.CSalt {
			b.WG.Done()
			return
		}
		if message =="clean"+b.CSalt {
			b.WG.Done()
			command:=<-ch
			if command == "close"+b.CSalt {

				return
			} else{
				trackerIndex,_ = strconv.Atoi(command)

				continue
			}
		}
		var remove []int;
		if b.IsBenchmarking {
		start = time.Now()
		}
		for i, c := range b.Routines[trackerIndex].Clients {
			select{
			case c.Pipe <- message : //pass message through
			default: //nobody is listening, delete client from array to save space :)
				close(c.Pipe);
				remove = append(remove, i)
			}
		}
		//collect time data
		finish := time.Since(start).Nanoseconds()
		if b.IsBenchmarking {
			b.BenchData = append(b.BenchData, TestStruct{finish,len(b.Routines[trackerIndex].Clients)*len(b.Routines),len(b.Routines)})
		}
		//check for clients to remove from the list
		if len(remove) > 0 {
			for p,n := range remove {
				b.Routines[trackerIndex].Lock()
				if (n-p) == len(b.Routines[trackerIndex].Clients)-1 {
					b.Routines[trackerIndex].Clients = b.Routines[trackerIndex].Clients[:n-p]
				} else {
					b.Routines[trackerIndex].Clients =  append(b.Routines[trackerIndex].Clients[:n-p], b.Routines[trackerIndex].Clients[n-p+1:]...)
				}
				b.Routines[trackerIndex].Unlock();
			}
		}
		if b.IsBenchmarking {
		b.WG.Done()
		}
	}
}

// CleanUp is used to terminate routines by the IDs specified in the supplied toClean parameter.
// This can also be used directly to shut a routine down if necessary, or trigger a full purge.
func (b *Broadcaster) CleanUp(toClean []int) {
	if b.CSalt == "" {
		b.CSalt = UUID()
	}
	b.WG.Add(len(b.Routines))
	for _, t := range b.Routines {
		t.Pipe <- "clean"+b.CSalt //send "clean" command
	}
	b.WG.Wait() //wait for the clean
	for p,q := range toClean {
		b.Routines[q-p].Pipe <- "close"+b.CSalt
		if q-p == len(b.Routines)-1 {
			b.Routines = b.Routines[:q-p]
		} else {
			b.Routines = append(b.Routines[:q-p], b.Routines[q-p+1:]...)
		}
	}
	for n,t := range b.Routines {
		t.Pipe <- strconv.Itoa(n)
	}
}
// Send uses the associated Broadcster object to send a message to all registered
// clients
func (b *Broadcaster) Send(msg string){
	var toClean []int
	for  n,t := range b.Routines {
		t.Pipe <- msg
		if len(t.Clients) == 0 { //If there are no more clients on the routine,
			toClean = append(toClean, n) //then mark them for cleanup
		}
	}
	if len(toClean) > 0 { //Begin cleanup of empty routines
			b.CleanUp(toClean) // <- make this synchronous to prevent messaging problems
	}
}

// Send uses the DefaultBroadcaster object to send a message to all registered
// clients.
func Send(msg string){
	DefaultBroadcaster.Send(msg)
}
