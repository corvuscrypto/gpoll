package gpoll

import (
		"fmt"
		"errors"
		"crypto/rand"
		)

//###### SIMPLE UUID V4 GENERATOR ######

func UUID() (string) {
  b := make([]byte, 16)
	rand.Read(b)
  b[6] = (b[6] & 0x0f) | 0x40
  b[8] = (b[8] & 0x3f) | 0x80
  return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4],b[4:6],b[6:8],b[8:10],b[10:])
}

//###### CLIENT FUNCTIONS ######

func (b *Broadcaster) addClient() (string, int){
	uid := UUID()
	nc := &Client{uid,make(chan string, b.CLIENT_BUFFER_SIZE)}
	var tid int
	b.mux.Lock()
	last := len(b.routines)-1
	if len(b.routines) == 0 {
		nT := new(Routine);
		nT.Pipe = make(chan string, b.ROUTINE_BUFFER_SIZE)
		nT.ID = 0
		b.routines = append(b.routines, *nT)
		b.mux.Unlock()
		go b.handleMessage(nT.Pipe, len(b.routines))

	} else if len(b.routines[last].Clients) == b.ROUTINE_MAX_CLIENTS{
		nT := new(Routine);
		nT.Pipe = make(chan string, b.ROUTINE_BUFFER_SIZE)
		i := true
		n := -1
		for i {
			n++
			for _,t:= range b.routines {
				i=false
				if t.ID == n {
					i = true
					break
				}
			}

		}
		nT.ID = n
		tid = n
		b.routines = append(b.routines, *nT)
		b.mux.Unlock()
		go b.handleMessage(nT.Pipe, len(b.routines))
	} else {
		b.mux.Unlock()
		tid = b.routines[len(b.routines)-1].ID
	}
	b.routines[len(b.routines)-1].Clients = append(b.routines[len(b.routines)-1].Clients, *nc)
	return nc.ID,tid
}

func (b *Broadcaster) getChannel(cid string, tid int) (chan string, error) {
	var err error = errors.New("Client not found")
	var ch chan string
	for i,t := range b.routines {
		if t.ID == tid {
			for _,v := range b.routines[i].Clients {
				if v.ID == cid {
					ch = v.Pipe
				  err = nil
					return ch, err
				}
			}
		}
	}
	return ch, err
}
