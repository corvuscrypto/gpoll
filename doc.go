// This package is free for use under the GNU public license v3 and more info
// can be found in the License file.

/*
The gpoll package manages long-polling. Long-polling is a means of near-realtime
data transfer that utilizes keep-alive requests to allow clients to receive data
as it becomes available. Notable uses for long-polling include push notification
and chat clients.

This package has the aim of making message distribution as quick as possible. The
way this is accomplished in gpoll is by utilizing coroutines (goroutines) and
tracking them so that coroutines are automatically created and terminated as
necessary based on the number of clients connected and your configuration.

Getting Started

To use gpoll, you first need to setup a Broadcaster object. The Broadcaster object
has several fields you can tweak to tune the behavior of the Broadcaster. For instance,
the ClientBufferSize dictates how many massages you will allow each client buffer
to store. However, because client removal is automated, the package will sense
when a buffer is full, then remove that client since the program assumes this
means the client is no longer listening.

If you wish to just drop in the package and use it right away, you can just have
the server use the DefaultBroadcaster. This is already available and by calling
the ListenAndBroadcast() function in the global context, your server will
default to this object and you're ready to go. An example of such a setup is:

  package main

  import  (
          "github.com/corvuscrypto/gpoll"
          "net/http"
          )

  func HandleMessage(w http.ResponseWriter, r *http.Request) {

          // code to detect message and do other stuff as
          // you please with the request.

          var msg string = "blah blah blah"

          gpoll.Send(msg) //send the message string

          w.WriteHeader(200)

  }

  func main(){

          gpoll.ListenAndBroadcast("/path/to/listen/on")
          http.HandleFunc("/send", HandleMessage)
          http.ListenAndServe(":8080",nil)

  }

It's as easy as that.

Tricks

Total reset. Okay this is probably something that should never be done unless
absolutely necessary, but if you need to do a full reset and kick all your users,
just call the CleanUp() function. Since the CleanUp function takes an argument in
the form of an int slice made up of the INDICES (not IDs) of the goroutines present
within a Broadcaster, just pass in a slice containing all indices of the
Broadcaster.Routines slice. This would look like this:

  var resetSlice []int
  for i,_ := range Broadcaster.Routines {
    resetSlice = append(resetSlice, i)
  }
  Broadcaster.CleanUp(resetSlice)

I'll add in support for individual client removal, but you can even implement your
own client removal since the number before the '$' in a client ID is the goroutine
ID. Just search the Routine object that contains that ID for your client within
the Clients slice and remove it.

*/
