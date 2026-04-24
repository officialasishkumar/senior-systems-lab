package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	mode := flag.String("mode", "http", "http, tcp, or udp")
	addr := flag.String("addr", "127.0.0.1:8080", "target address")
	topic := flag.String("topic", "ops.events", "message topic")
	payload := flag.String("payload", "probe", "message payload")
	flag.Parse()

	var err error
	switch *mode {
	case "http":
		err = probeHTTP(*addr, *topic, *payload)
	case "tcp":
		err = probeTCP(*addr, *topic, *payload)
	case "udp":
		err = probeUDP(*addr)
	default:
		err = fmt.Errorf("unsupported mode %q", *mode)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func probeHTTP(addr, topic, payload string) error {
	body, _ := json.Marshal(map[string]string{"topic": topic, "payload": payload})
	resp, err := http.Post("http://"+addr+"/queue/publish", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("http probe failed: %s %s", resp.Status, string(data))
	}
	fmt.Print(string(data))
	return nil
}

func probeTCP(addr, topic, payload string) error {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	body, _ := json.Marshal(map[string]string{"topic": topic, "payload": payload})
	if err := writeFrame(conn, body); err != nil {
		return err
	}
	reply, err := readFrame(conn)
	if err != nil {
		return err
	}
	fmt.Println(string(reply))
	return nil
}

func probeUDP(addr string) error {
	conn, err := net.DialTimeout("udp", addr, 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := conn.Write([]byte("PING")); err != nil {
		return err
	}
	buf := make([]byte, 128)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}
	fmt.Println(string(buf[:n]))
	return nil
}

func readFrame(r io.Reader) ([]byte, error) {
	var size uint32
	if err := binary.Read(r, binary.BigEndian, &size); err != nil {
		return nil, err
	}
	payload := make([]byte, size)
	_, err := io.ReadFull(r, payload)
	return payload, err
}

func writeFrame(w io.Writer, payload []byte) error {
	if err := binary.Write(w, binary.BigEndian, uint32(len(payload))); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}
