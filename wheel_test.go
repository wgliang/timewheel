package timewheel

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

type Wheel struct {
	mux   *sync.Mutex
	Cache map[string]string
}

func NewWheel() *Wheel {
	w := &Wheel{
		mux: new(sync.Mutex),
	}
	w.Cache = make(map[string]string, 0)
	return w
}

func (w *Wheel) Get(key string) string {
	w.mux.Lock()
	defer w.mux.Unlock()
	return w.Cache[key]
}

func (w *Wheel) Add(key, value string) {
	w.mux.Lock()
	defer w.mux.Unlock()
	w.Cache[key] = value
}

func (w *Wheel) Remove(key string) {
	w.mux.Lock()
	defer w.mux.Unlock()
	delete(w.Cache, key)
}

func goValue(tw *TimeWheel, w *Wheel) {
	ti := time.Tick(1 * time.Second)
	key := 0
	for {
		select {
		case <-ti:
			if key == 20 {
				return
			}
			key++
			value := time.Now().Format("2016-11-14 22:22:22")
			w.Add(strconv.Itoa(key), value)
			fmt.Println("add to Wheel...", value)
			tw.Add(strconv.Itoa(key))
			fmt.Println("add to TimeWheel...", key)
		}
	}
}

func PrintWheel(w *Wheel) {
	ti := time.Tick(1 * time.Second)
	key := 0
	for {
		select {
		case <-ti:
			if key == 30 {
				return
			}
			key++
			w.mux.Lock()
			fmt.Println(w.Cache)
			w.mux.Unlock()
		}
	}
}

func TestTimeWheel(t *testing.T) {
	w := NewWheel()
	wheel := NewTimeWheel(time.Second*1, 10, func(w interface{}, key interface{}) {
		w.(*Wheel).Remove(key.(string))
	}, w)
	wheel.Start()
	go goValue(wheel, w)
	go PrintWheel(w)

	time.Sleep(30 * time.Second)
}
