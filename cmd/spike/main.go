//go:build cgo && mpv

// Command spike measures the playback semantics we depend on, headless:
// 4 simultaneous files, per-pane speed, relative seek, and CPU cost.
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"mvplayer/internal/player"
)

func cpuTicks() (u, s int64) {
	b, err := os.ReadFile("/proc/self/stat")
	if err != nil {
		return 0, 0
	}
	f := strings.Fields(string(b))
	if len(f) < 15 {
		return 0, 0
	}
	u, _ = strconv.ParseInt(f[13], 10, 64)
	s, _ = strconv.ParseInt(f[14], 10, 64)
	return u, s
}

func main() {
	file := flag.String("file", "", "mkv used for every pane")
	n := flag.Int("n", 4, "number of panes")
	rate := flag.Float64("rate", 1.0, "initial rate for every pane")
	flag.Parse()
	if *file == "" {
		fmt.Fprintln(os.Stderr, "usage: spike -file x.mkv [-n 4] [-rate 1]")
		os.Exit(2)
	}

	var engines []*player.Engine
	defer func() {
		for _, e := range engines {
			e.Close()
		}
	}()

	u0, s0 := cpuTicks()
	start := time.Now()
	for i := 0; i < *n; i++ {
		e, err := player.NewEngine()
		if err != nil {
			fmt.Println("ERR new:", err)
			os.Exit(1)
		}
		engines = append(engines, e)
		if err := e.SetOption("vo", "null"); err != nil {
			fmt.Println("ERR vo:", err)
		}
		if err := e.SetOption("ao", "null"); err != nil {
			fmt.Println("ERR ao:", err)
		}
		if err := e.Open(*file); err != nil {
			fmt.Println("ERR open:", err)
		}
	}
	// let all four reach file-loaded
	time.Sleep(1500 * time.Millisecond)
	for i, e := range engines {
		fmt.Printf("load pane=%d dur=%dms pos=%dms\n", i, e.Duration()/1, e.Position())
	}
	for _, e := range engines {
		if err := e.SetRate(*rate); err != nil {
			fmt.Println("ERR rate:", err)
		}
		if err := e.Play(); err != nil {
			fmt.Println("ERR play:", err)
		}
	}
	time.Sleep(4 * time.Second)
	for i, e := range engines {
		fmt.Printf("playing pane=%d pos=%dms rate=%.2f paused=%v\n", i, e.Position(), e.Rate(), e.Paused())
	}

	// per-pane rates: 0.5 / 1 / 2 / 4
	rates := []float64{0.5, 1, 2, 4}
	before := make([]int64, len(engines))
	for i, e := range engines {
		if i < len(rates) {
			_ = e.SetRate(rates[i])
		}
		before[i] = e.Position()
	}
	time.Sleep(3000 * time.Millisecond)
	for i, e := range engines {
		delta := e.Position() - before[i]
		fmt.Printf("rate pane=%d set=%.2f effective=%.2f\n", i, rates[min(i, len(rates)-1)], float64(delta)/3000.0)
	}

	// global seek: +10s on all, clamped per pane
	for _, e := range engines {
		_ = e.SetRate(1)
	}
	_ = engines[0].Seek(0)
	time.Sleep(300 * time.Millisecond)
	for i, e := range engines {
		pos := e.Position()
		target := pos + 10000
		if d := e.Duration(); d > 0 && target > d {
			target = d
		}
		_ = e.Seek(target)
		fmt.Printf("seek pane=%d from=%dms target=%dms\n", i, pos, target)
	}
	time.Sleep(400 * time.Millisecond)
	for i, e := range engines {
		fmt.Printf("seeked pane=%d pos=%dms\n", i, e.Position())
	}

	u1, s1 := cpuTicks()
	wall := time.Since(start).Seconds()
	cpu := float64((u1-u0)+(s1-s0)) / 100.0 / wall * 100
	fmt.Printf("cpu avg=%.1f%% wall=%.1fs panes=%d\n", cpu, wall, len(engines))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
