package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ISI0007/pulseroute-network-lab/internal/bench"
	"github.com/ISI0007/pulseroute-network-lab/internal/network"
	"github.com/ISI0007/pulseroute-network-lab/internal/store"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

func main() {
	mode := flag.String("mode", "serve", "serve, chat, send, bench, discover, selftest")
	addr := flag.String("addr", "127.0.0.1:42191", "TCP endpoint")
	udp := flag.String("udp", "127.0.0.1:42192", "UDP endpoint")
	dir := flag.String("data", "data", "private data directory")
	name := flag.String("name", "operator", "peer name")
	file := flag.String("file", "", "file to send")
	protocol := flag.String("protocol", "tcp", "tcp or udp")
	count := flag.Int("count", 100, "echo probe count")
	size := flag.Int("bytes", 1024, "payload bytes")
	flag.Parse()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	switch *mode {
	case "serve":
		if e := os.MkdirAll(*dir, 0700); e != nil {
			log.Fatal(e)
		}
		db, e := store.Open(filepath.Join(*dir, "history.db"))
		if e != nil {
			log.Fatal(e)
		}
		defer db.Close()
		s, e := network.Start(*addr, filepath.Join(*dir, "received"), db.Record)
		if e != nil {
			log.Fatal(e)
		}
		defer s.Close()
		u, e := bench.StartEcho(*udp)
		if e != nil {
			log.Fatal(e)
		}
		defer u.Close()
		fmt.Printf("TCP %s | UDP %s\n", s.Addr(), u.Addr())
		<-ctx.Done()
	case "chat":
		c, e := network.Connect(ctx, *addr, *name)
		if e != nil {
			log.Fatal(e)
		}
		defer c.Close()
		go func() {
			for h := range c.Events {
				b, _ := json.Marshal(h)
				fmt.Println(string(b))
			}
		}()
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if _, e = c.Send(scanner.Text()); e != nil {
				log.Fatal(e)
			}
		}
		if e = scanner.Err(); e != nil {
			log.Fatal(e)
		}
	case "send":
		h, e := network.SendFile(ctx, *addr, *file, func(ev network.Event) { fmt.Fprintf(os.Stderr, "%s %d/%d\n", ev.Status, ev.Bytes, ev.Total) })
		if e != nil {
			log.Fatal(e)
		}
		enc.Encode(h)
	case "bench":
		target := *addr
		if *protocol == "udp" {
			target = *udp
		}
		r, e := bench.Run(ctx, *protocol, target, *count, *size)
		if e != nil {
			log.Fatal(e)
		}
		enc.Encode(r)
	case "discover":
		p, e := network.Discover(ctx, *udp)
		if e != nil {
			log.Fatal(e)
		}
		enc.Encode(p)
	case "selftest":
		s, e := network.Start("127.0.0.1:0", filepath.Join(*dir, "received"), nil)
		if e != nil {
			log.Fatal(e)
		}
		defer s.Close()
		u, e := bench.StartEcho("127.0.0.1:0")
		if e != nil {
			log.Fatal(e)
		}
		defer u.Close()
		results := []bench.Result{}
		for i := 0; i < 5; i++ {
			for _, p := range []string{"tcp", "udp"} {
				target := s.Addr()
				if p == "udp" {
					target = u.Addr()
				}
				r, e := bench.Run(ctx, p, target, *count, *size)
				if e != nil {
					log.Fatal(e)
				}
				results = append(results, r)
				time.Sleep(20 * time.Millisecond)
			}
		}
		enc.Encode(map[string]any{"environment": "Windows amd64; loopback; sequential echo; no impairment", "recordedAt": time.Now().UTC(), "runs": results})
	default:
		log.Fatal("unknown mode")
	}
}
