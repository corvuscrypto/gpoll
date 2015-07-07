package gpoll

import  (
        "reflect"
        "testing"
        )

func TestAddClients(t *testing.T){

  b := &Broadcaster{RoutineMaxClients: 10}

  for i:=0; i< 100; i++ {

    b.AddClient()

  }
  if len(b.Routines) != 10 {
    t.Errorf("Test expected 10 routines, but encountered %v!", len(b.Routines))
  }

}

func TestUUID(t *testing.T){

  uid := UUID()

  if len(uid)!= 36{

    t.Errorf("Expected valid uuid, but received uuid of length %v",len(uid))

  }

}

func TestGetChannel(t *testing.T){

    b := &Broadcaster{RoutineMaxClients: 10}

    var uids []string
    var tids []int

    for i := 0; i<10; i++{

      tuid, ttid := b.AddClient()

      uids = append(uids, tuid)
      tids = append(tids, ttid)
    }

    ch, err := b.GetChannel(uids[1],tids[1])


    if err != nil {

      t.Errorf("Test encountered the following error: %v", err)

    }
    if reflect.TypeOf(ch) != reflect.ChanOf(reflect.BothDir, reflect.TypeOf("")) {

      t.Errorf("Test expected to retrieve channel, but received %v", reflect.TypeOf(ch))
    }

    ch, err = b.GetChannel("asd",3)
    if err == nil {

      t.Error("An error was expected, but not received")

    }
}
