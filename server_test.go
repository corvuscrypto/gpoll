package gpoll

import (
        "testing"
        "time"
        "net/http"
        "net/http/httptest"
        "io/ioutil"
        "fmt"
        "strings"
)

func TestTotalRoutineDump(t *testing.T){

  // tests to ensure that routines get cleaned up when there
  // are no more clients.

  b := &Broadcaster{RoutineMaxClients: 10, ClientBufferSize:10, RoutineBufferSize:10}
  //test whole dump
  b.Benchmark(1000,10,10)

  if len(b.Routines)>0 {

    t.Errorf("Expected at least 1 routine, detected %v", len(b.Routines))

  }

}

func TestSingleRoutineDump(t *testing.T){

  b := &Broadcaster{RoutineMaxClients:10,
                    ClientBufferSize:1,
                    RoutineBufferSize:10}

  //test removal of single routine
  // if the client buffer is 1, it will take 3 sends to clean up.
  // This is because it takes one message to fill up the buffer,
  // then one more to initiate client removal within the routine.
  // Finally, one more send is required for the server to realize the
  // routine's client array is empty and initiate coroutine termination.

  for i:=0; i<10;i++ {
    b.AddClient()
  }

  for i:=0; i<3;i++{
    if i == 1 {
      for j:=0; j<1;j++ {
        b.AddClient()
      }
    }
    b.Send("test")
    time.Sleep(1*time.Millisecond) //necessary since the asynchronicity
                                     //will mean it takes some time
                                     //for the commands to be performed
                                     //this does not cause problems in
                                     //production and doesn't impact speed
                                     //significantly
  }

  if len(b.Routines) != 1 {
    t.Errorf("Expected 1 routine to be left, detected %v", len(b.Routines))
  }

  //then send one more message to ensure that there are 0 after the next send

  b.Send("test")

  if len(b.Routines) != 0 {
    t.Errorf("Expected 0 routine to be left, detected %v", len(b.Routines))
  }

}

func TestDefaultSendAndClean(t *testing.T){

  DefaultBroadcaster.AddClient()
  Send("test")
  if len(DefaultBroadcaster.Routines) != 1 {
    t.Errorf("Expected 1 routine to be left, detected %v",len(DefaultBroadcaster.Routines))
  }
  for i:=0; i< DefaultBroadcaster.ClientBufferSize+2;i++{
    Send("test")
    time.Sleep(1*time.Millisecond)
  }

  if len(DefaultBroadcaster.Routines) != 0 {
    t.Errorf("Expected 0 routine to be left, detected %v",len(DefaultBroadcaster.Routines))
  }

}

func TestDummyServer(t *testing.T){

  dummySrv := httptest.NewServer(nil)
  defer dummySrv.Close()

  //test with default Broadcaster
  ListenAndBroadcast("/poll0")
  ListenAndBroadcast("/poll1", http.DefaultServeMux)
  for i:=0;i<2;i++ {
    resp,_ := http.Get(fmt.Sprintf(dummySrv.URL+"/poll%d",i))
    cont,_ := ioutil.ReadAll(resp.Body)
    if !strings.Contains(string(cont),"client-id") {
      t.Errorf("Expected to receive a new client id, but instead received: %s", cont)
    }
    resp.Body.Close()
  }
  //test with custom Broadcaster
  b := &Broadcaster{RoutineMaxClients:10, ClientBufferSize:1, RoutineBufferSize:10}
  b.ListenAndBroadcast("/poll2")
  b.ListenAndBroadcast("/poll3", http.DefaultServeMux)
  var uuid string
  for i:=2;i<4;i++ {
    resp,_ := http.Get(fmt.Sprintf(dummySrv.URL+"/poll%d",i))
    cont,_ := ioutil.ReadAll(resp.Body)
    if !strings.Contains(string(cont),"client-id") {
      t.Errorf("Expected to receive a new client id, but instead received: %s", cont)
    }
    uuid = string(cont)
    resp.Body.Close()
  }

  //test request with an actual client-id
  uuid = strings.Split(uuid, ":")[1]
  uuid = strings.Split(uuid,"\"")[1]
  b.Send("test")
  resp,_ := http.Get(dummySrv.URL+"/poll3?client-id="+uuid)
  cont,_ := ioutil.ReadAll(resp.Body)
  resp.Body.Close()
  if string(cont) != "test" {
    t.Errorf("Expected to find the test message, but instead received: %s", cont)
  }

  //ensure that broadcasters don't cross-handle clients
  Send("test")
  resp,_ = http.Get(dummySrv.URL+"/poll1?client-id="+uuid)
  if resp.StatusCode != 404 {
    t.Errorf("Expected to find an Error 404 message, but instead received: %v", resp.StatusCode)
  }

  //test requests with bad uuid.....
  resp,_ = http.Get(dummySrv.URL+"/poll3?client-id="+uuid[3:])
  if resp.StatusCode != 404 {
    t.Errorf("Expected to find an Error 404 message, but instead received: %v", resp.StatusCode)
  }
  resp.Body.Close()


  //...And now with an un-findable uuid
  resp,_ = http.Get(dummySrv.URL+"/poll3?client-id="+uuid[:len(uuid)-1]+"{")
  if resp.StatusCode != 404 {
    t.Errorf("Expected to find an Error 404 message, but instead received: %v", resp.StatusCode)
  }
  resp.Body.Close()

}
