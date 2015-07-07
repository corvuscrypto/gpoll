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
	
//###### DATA STRUCTS ######
type Broadcaster struct {
	
	// Number of clients per routine
	ROUTINE_MAX_CLIENTS 	int
	
	// Maximum number of messages per client channel
	CLIENT_BUFFER_SIZE 	int
	
	// Maximum number of messages to hold in each routine's buffer
	ROUTINE_BUFFER_SIZE	int
	
	// Number of milliseconds to wait before broadcasting message in master buffer
	// default is 0, but can be set to compensate for the time it takes for
	// clients to reconnect to the polling service.
	SYNC_DELAY_MS 		int
	
	// Hold routine references within the Broadcaster to allow multiple instances
	// of the Broadcaster to work side-by-side
	Routines		[]Routine
	
	// Sync WaitGroup for ensuring proper message delivery
	WG			sync.WaitGroup
	
	// Mutex for maintaining proper client data control
	Mux			sync.Mutex
	
	// bool used for controlling benchmark behavior
	IsBenchmarking		bool
	
	// array for holding data during benchmarking. This will
	// always be cleared after benchmarking to save memory.
	benchData		[]TestStruct
	
	// Command salt. This will automatically be generated. You can generate one
	// if you like, but make sure it is not publicly known.
	CSalt			string
}
type Routine struct{
	sync.Mutex
	ID 					int
	Pipe 				chan string
	Clients 		[]Client
}
type TestStruct struct {
	Ns 				int64
	Clients 	int
	Routines 	int
}
type Client struct {
	ID 		string
	Pipe 	chan string
}

//Broadcaster Default
func NewBroadcaster() *Broadcaster {
	return 	&Broadcaster 	{
				ROUTINE_MAX_CLIENTS: 1000,
				CLIENT_BUFFER_SIZE: 100,
				ROUTINE_BUFFER_SIZE: 200,
				SYNC_DELAY_MS: 0,
				}
}

var DefaultBroadcaster = NewBroadcaster()

//###### SERVER FUNCTIONS ######

func ListenAndBroadcast(path string, srv ...*http.ServeMux){
	if len(srv) > 0 {
		srv[0].HandleFunc(path,DefaultBroadcaster.HandleRequests)
	}	else {
	http.HandleFunc(path,DefaultBroadcaster.HandleRequests)
	}
}
func (b *Broadcaster) ListenAndBroadcast(path string, srv ...*http.ServeMux){
	if len(srv) > 0 {
		srv[0].HandleFunc(path,b.HandleRequests)
	}	else {
	http.HandleFunc(path,b.HandleRequests)
	}
}
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
			w.WriteHeader(http.StatusExpectationFailed)
			return
		}
		tid,_ := strconv.Atoi(info[0])
		q,err := b.GetChannel(info[1],tid)
		if err != nil{
			//send error if client not found
			w.WriteHeader(http.StatusExpectationFailed)
			return
		}
		temp := <-q
		w.Write([]byte(temp))
	}
}
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

//###### POLL MESSAGING FUNCTIONs ######
func (b *Broadcaster) CleanUp(toClean []int) {
	if b.CSalt == "" {
		b.CSalt = UUID()
	}
	b.WG.Add(len(b.routines))
	for _, t := range b.routines {
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
func Send(msg string){
	DefaultBroadcaster.Send(msg)
}
