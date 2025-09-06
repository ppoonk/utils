package test

import (
	"sync"
	"sync/atomic"
	"testing"
	"utils/balancer"
)

func TestSmoothWeightedRoundRobin(t *testing.T) {
	server1 := newMockServer("server1", 1)
	server2 := newMockServer("server2", 2)
	server3 := newMockServer("server3", 3)

	scheduler := balancer.NewSmoothWeightedRoundRobin([]*mockServer{server1, server2, server3})

	var (
		concurrentTimes = 6000
		expectedServer1 = 1000
		expectedServer2 = 2000
		expectedServer3 = 3000
		resultServer1   atomic.Int32
		resultServer2   atomic.Int32
		resultServer3   atomic.Int32
	)

	var wg sync.WaitGroup
	wg.Add(concurrentTimes)

	for i := 0; i < concurrentTimes; i++ {
		go func() {
			defer wg.Done()
			switch scheduler.Select().GetName() {
			case "server1":
				resultServer1.Add(1)
			case "server2":
				resultServer2.Add(1)
			case "server3":
				resultServer3.Add(1)
			}
		}()
	}

	wg.Wait()

	t.Logf("Server1 weight distribution: expected %d times, actual %d times", expectedServer1, resultServer1.Load())
	t.Logf("Server2 weight distribution: expected %d times, actual %d times", expectedServer2, resultServer2.Load())
	t.Logf("Server3 weight distribution: expected %d times, actual %d times", expectedServer3, resultServer3.Load())
}

type mockServer struct {
	name      string
	weight    int          // Initial weight
	curWeight atomic.Int32 // Current weight
}

func (m *mockServer) GetWeight() int {
	return m.weight
}

func (m *mockServer) GetCurrentWeight() int {
	return int(m.curWeight.Load())
}

func (m *mockServer) SetCurrentWeight(weight int) {
	m.curWeight.Store(int32(weight))
}

func (m *mockServer) GetName() string {
	return m.name
}

func newMockServer(name string, weight int) *mockServer {
	return &mockServer{
		name:   name,
		weight: weight,
	}
}
