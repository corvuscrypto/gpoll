package gpoll

import (
        "testing"
)

func TestBenchmark(t *testing.T){

      b := &Broadcaster{RoutineMaxClients: 10, ClientBufferSize:1, RoutineBufferSize:1}

      err := b.Benchmark(1,1,1)

      if err != nil {

        t.Errorf("An error occured:  %v", err)

      }

      err = b.Benchmark(0,0,0)
      if err == nil {
        t.Error("Expected an error, functioned returned nil")
      }
      err = b.Benchmark(1,0,0)
      if err == nil {
        t.Error("Expected an error, functioned returned nil")
      }
      err = b.Benchmark(1,1,0)
      if err == nil {
        t.Error("Expected an error, functioned returned nil")
      }
      err = b.Benchmark(1,1,2)
      if err == nil {
        t.Error("Expected an error, functioned returned nil")
      }

}
