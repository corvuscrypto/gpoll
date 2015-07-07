package gpoll

import (
		"fmt"
		"errors"
		"crypto/rand"
		)

//UUID generator. I can't take credit for this as it was obtained from a golang-nuts
// discussion board (https://groups.google.com/forum/#!msg/golang-nuts/d0nF_k4dSx4/rPGgfXv6QCoJ).
// Acknowledgements to Russ Cox for the code and for saving me time from reading through the
// official UUID documentation.
func UUID() string {
  b := make([]byte, 16)
	rand.Read(b)
  b[6] = (b[6] & 0x0f) | 0x40
  b[8] = (b[8] & 0x3f) | 0x80
  return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4],b[4:6],b[6:8],b[8:10],b[10:])
}

// AddClient adds a client to the associated Broadcaster and returns the client ID
// and the routine id that manages the client.
func (b *Broadcaster) AddClient() (string, int){
	uid := UUID()
	nc := &Client{uid,make(chan string, b.ClientBufferSize)}
	var tid int
	b.Mux.Lock()
	last := len(b.Routines)-1
	if len(b.Routines) == 0 {
		nT := new(Routine);
		nT.Pipe = make(chan string, b.RoutineBufferSize)
		nT.ID = 0
		b.Routines = append(b.Routines, *nT)
		b.Mux.Unlock()
		go b.HandleMessage(nT.Pipe, len(b.Routines))

	} else if len(b.Routines[last].Clients) == b.RoutineMaxClients{
		nT := new(Routine);
		nT.Pipe = make(chan string, b.RoutineBufferSize)
		i := true
		n := -1
		for i {
			n++
			for _,t:= range b.Routines {
				i=false
				if t.ID == n {
					i = true
					break
				}
			}

		}
		nT.ID = n
		tid = n
		b.Routines = append(b.Routines, *nT)
		b.Mux.Unlock()
		go b.HandleMessage(nT.Pipe, len(b.Routines))
	} else {
		b.Mux.Unlock()
		tid = b.Routines[len(b.Routines)-1].ID
	}
	b.Routines[len(b.Routines)-1].Clients = append(b.Routines[len(b.Routines)-1].Clients, *nc)
	return nc.ID,tid
}

// GetChannel retrieves the reference to the receiving channel for a client based on the
// supplied client id ('cid') and the routine id ('rid').
func (b *Broadcaster) GetChannel(cid string, rid int) (chan string, error) {
	var err error = errors.New("Client not found")
	var ch chan string
	for i,t := range b.Routines {
		if t.ID == rid {
			for _,v := range b.Routines[i].Clients {
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
