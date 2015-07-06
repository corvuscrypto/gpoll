package gpoll

import (
        "testing"
)

func TestBenchmark(t *testing.T){

      b := &Broadcaster{ROUTINE_MAX_CLIENTS: 10, CLIENT_BUFFER_SIZE:1, ROUTINE_BUFFER_SIZE:1}

      err := b.benchmark(1,1,1)

      if err != nil {

        t.Errorf("An error occured:  %v", err)

      }

      err = b.benchmark(0,0,0)
      if err == nil {
        t.Error("Expected an error, functioned returned nil")
      }
      err = b.benchmark(1,0,0)
      if err == nil {
        t.Error("Expected an error, functioned returned nil")
      }
      err = b.benchmark(1,1,0)
      if err == nil {
        t.Error("Expected an error, functioned returned nil")
      }
      err = b.benchmark(1,1,2)
      if err == nil {
        t.Error("Expected an error, functioned returned nil")
      }

}
