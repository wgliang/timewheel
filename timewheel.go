package timewheel

import (
	"sync"
	"time"
)

// The slot of wheel,a slot represents all wheel's unit
// a slot can have some elements.You can choose a slot
// just for you or meet your need.
type slot struct {
	mux      *sync.Mutex
	id       int
	elements map[interface{}]interface{}
}

// New a slot and init the (id) and new a mutex just for
// concurrent operations.
func newSlot(id int) *slot {
	s := &slot{
		id:  id,
		mux: new(sync.Mutex),
	}
	s.elements = make(map[interface{}]interface{})
	return s
}

// Add a interface into slot.
func (s *slot) add(c interface{}) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.elements[c] = c
}

// Remove a interface form slot,if you won't keep it.
func (s *slot) remove(c interface{}) {
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.elements, c)
}

// The handler taht you want to with timewheel.
type handler func(manager interface{}, u interface{})

// Timewheel struct is the base struct for all timewheel.
type TimeWheel struct {
	mux              *sync.Mutex
	fmanager         interface{}
	tickDuration     time.Duration
	ticksOfWheel     int
	currentTickIndex int
	ticker           *time.Ticker
	onTick           handler
	wheel            []*slot
	indicator        map[interface{}]*slot

	taskChan chan interface{}
	quitChan chan interface{}
}

// New a timewheel then you can keep all you want to keep.
// And you must ensure that parameter (tickDuration >= 1)
// and (ticksOfWheel >= 1),or,no one can new it,you should
// know it.
func NewTimeWheel(tickDuration time.Duration, ticksOfWheel int, f handler, manager interface{}) *TimeWheel {
	if tickDuration < 1 || ticksOfWheel < 1 || nil == f {
		return nil
	}
	// A slot that will always come out and it just for 0.
	ticksOfWheel++
	// Whenever you new a timewheel, you will get currentTickIndex is 0.
	tw := &TimeWheel{
		fmanager:         manager,
		tickDuration:     tickDuration,
		ticksOfWheel:     ticksOfWheel,
		onTick:           f,
		currentTickIndex: 0,
		taskChan:         make(chan interface{}, 10),
		quitChan:         make(chan interface{}, 10),
		mux:              new(sync.Mutex),
	}
	// Make a indicator and init it 0
	tw.indicator = make(map[interface{}]*slot, 0)

	tw.wheel = make([]*slot, ticksOfWheel)
	// You can see :i is the id of slot.
	for i := 0; i < ticksOfWheel; i++ {
		tw.wheel[i] = newSlot(i)
	}

	return tw
}

// Start TimeWheel, what you should do is new a ticker and run it.
func (tw *TimeWheel) Start() {
	tw.ticker = time.NewTicker(tw.tickDuration)
	go tw.run()
}

// Add a interface into TimeWheel,and add c interface{} into task channal.
func (tw *TimeWheel) Add(c interface{}) {
	tw.taskChan <- c
}

// Remove c interface from TimeWheel.
func (tw *TimeWheel) Remove(c interface{}) {
	tw.mux.Lock()
	defer tw.mux.Unlock()
	if v, ok := tw.indicator[c]; ok {
		v.remove(c)
	}
}

// Get previous tick index
func (tw *TimeWheel) getPreviousTickIndex() int {
	tw.mux.Lock()
	defer tw.mux.Unlock()

	cti := tw.currentTickIndex
	// When cti==0 then cti's id wiil be ticksOfWheel-1
	if 0 == cti {
		return tw.ticksOfWheel - 1
	}
	return cti - 1
}

// Stop TimeWheel
func (tw *TimeWheel) Stop() {
	close(tw.quitChan)
}

// The internal run function, manage all slots and interface
func (tw *TimeWheel) run() {
	for {
		select {
		case <-tw.quitChan:
			tw.ticker.Stop()
			break
		case <-tw.ticker.C:
			tw.mux.Lock()
			if tw.ticksOfWheel == tw.currentTickIndex {
				tw.currentTickIndex = 0
			}
			slot := tw.wheel[tw.currentTickIndex]

			for _, v := range slot.elements {
				slot.remove(v)
				delete(tw.indicator, v)
				tw.onTick(tw.fmanager, v)
			}
			tw.currentTickIndex++

			tw.mux.Unlock()
		case v := <-tw.taskChan:
			c := v.(string)
			tw.Remove(c)
			slot := tw.wheel[tw.getPreviousTickIndex()]
			tw.mux.Lock()
			slot.add(c)
			tw.indicator[c] = slot
			tw.mux.Unlock()
		}
	}
}
