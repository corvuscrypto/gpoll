# gpoll
Long-polling server package for Go. Meant to be generic enough to work with any server generated with go. Features automated client management and coroutine (goroutine) generation and termination based on default or custom settings so you can optimize your long-polling application.
## Documentation
GoDoc available at http://godoc.org/github.com/corvuscrypto/gpoll

A basic example of starting a server is:

```go
package main

import  (
        "github.com/corvuscrypto/gpoll"
        "net/http"
        )

func HandleMessage(w http.ResponseWriter, r *http.Request) {

        /* code to detect message and do other stuff as 
           you please with the request. */
           
        var msg string = "blah blah blah"
           
        gpoll.Send(msg) //send the message string
        
        w.WriteHeader(200)
        
}

func main(){

        gpoll.ListenAndBroadcast("/path/to/listen/on")
        http.HandleFunc("/send", HandleMessage)
        http.ListenAndServe(":8080",nil)

}
```
And it's as easy as that!!

## The ONE rule
If you are using this package, the one rule that keeps this server ticking is that you must keep track of the uuid the server sends back after initializing the polling connection. This is actually not a bad thing since in most cases it is good to track your clients anonymously, but before allowing the poll to go through properly the gpoll package always expects a query with at least the following query: 
`?client-id=x$xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`

### Why?
The gpoll package uses this to maintain its routines and retrieve the proper channels to listen for messages. If you send an improper id, then you will receive an Error 404 response. **To initiate a client properly so the server returns a uuid to use for subsequent polls, send the following query:**
`?client-id=0`

##### if you're into the specifics...
the identifier before the `$` points to the goroutine that is tracking that client. The stuff after the `$` is the actual client ID.


## Features
* Automated client UUID (v4) generation
* Automated goroutine generation and termination as needed
* Optional Customization of the broadcasting server
* Easy to link to any TCP server already in place
  * includes TLS servers.
* Benchmarking suite (separate from the `go test -bench` suite) for fine tuning performance to fit your machine.
* Other stuff, like buffers to prevent your clients from missing notifications, so you can spend less time making the server, and more time on developing your APIs and front-end.
