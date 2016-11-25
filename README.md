# timewheel

timewheel-golang

[![Build Status](https://travis-ci.org/wgliang/appmonitor.svg?branch=master)](https://travis-ci.org/wgliang/appmonitor)
[![GoDoc](https://godoc.org/github.com/wgliang/appmonitor?status.svg)](https://godoc.org/github.com/wgliang/appmonitor)
[![Join the chat at https://gitter.im/appmonitor/Lobby](https://badges.gitter.im/appmonitor/Lobby.svg)](https://gitter.im/appmonitor/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Code Health](https://landscape.io/github/wgliang/appmonitor/master/landscape.svg?style=flat)](https://landscape.io/github/wgliang/appmonitor/master)
[![Code Issues](https://www.quantifiedcode.com/api/v1/project/98b2cb0efd774c5fa8f9299c4f96a8c5/badge.svg)](https://www.quantifiedcode.com/app/project/98b2cb0efd774c5fa8f9299c4f96a8c5)
[![Go Report Card](https://goreportcard.com/badge/github.com/wgliang/appmonitor)](https://goreportcard.com/report/github.com/wgliang/appmonitor)
[![License](https://img.shields.io/badge/LICENSE-Apache2.0-ff69b4.svg)](http://www.apache.org/licenses/LICENSE-2.0.html)

timewheel is a library which provides a timer on your resource,such as net-connection 
or data-in-memory and so on.  The timewheel contain interface 'TimeWheel.Start()',
'TimeWheel.Add()','TimeWheel.Stop()' and 'TimeWheel.Remove()'.

Major additional concepts are:

In addition to the [godoc API documentation](http://godoc.org/github.com/wgliang/timewheel)

## Recent Changes

support type-string,but you can rewrite it in line [155](https://github.com/wgliang/timewheel/blob/master/timewheel.go#L155).

## Install

    go get github.com/wgliang/timewheel

## Usage

Below is an example which shows some common use cases for timewheel.  Check 
[wheel_test.go](https://github.com/wgliang/timewheel/blob/master/wheel_test.go) for more
usage.


```go
package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/wgliang/timewheel"
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

func goValue(tw *timewheel.TimeWheel, w *Wheel) {
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
			// 业务中对资源的管理
			w.Add(strconv.Itoa(key), value)
			fmt.Println("add to Wheel...", value)
			// 不要忘记同时在时间轮里也要做改变，原则就是业务中的改变记得通知时间轮，但时间轮做的工作我们无需关心
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

func main() {
    // 初始化你的资源和接口
	w := NewWheel()
	// 传入你要对你的资源要做的操作，以及传入回调函数
	wheel := timewheel.NewTimeWheel(time.Second*1, 10, func(w interface{}, key interface{}) {
		w.(*Wheel).Remove(key.(string))
	}, w)
	// 开启我们的时间轮
	wheel.Start()
	// 在这里代表你的项目中对资源做的改变，增加等等，然后剩下的就交给时间轮管理吧
	go goValue(wheel, w)
	go PrintWheel(w)
	time.Sleep(30 * time.Second)
}
```

