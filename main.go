package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// metrics for SRE observability
type ProxyMetrics struct {
	TotalRequests    uint64
	HeartbeatSuccess uint64
	HeartbeatFail    uint64
	mu               sync.Mutex
}

var (
	targetConn net.Conn
	connMutex  sync.Mutex
	metrics    ProxyMetrics
)

// convertToToon recursively transforms arbitrary JSON into Tabular TOON.
func convertToToon(data interface{}, indent int, keyName string) string {
	prefix := strings.Repeat("  ", indent)
	switch v := data.(type) {
	case map[string]interface{}:
		var lines []string
		for key, val := range v {
			lines = append(lines, convertToToon(val, indent, key))
		}
		return strings.Join(lines, "\n")
	case []interface{}:
		if len(v) == 0 {
			return fmt.Sprintf("%s%s[0]:", prefix, keyName)
		}
		if keys, uniform := getUniformKeys(v); uniform {
			header := fmt.Sprintf("%s%s[%d]{%s}:", prefix, keyName, len(v), strings.Join(keys, ","))
			var rows []string
			rows = append(rows, header)
			for _, item := range v {
				obj := item.(map[string]interface{})
				var vals []string
				for _, k := range keys {
					vals = append(vals, fmt.Sprintf("%v", obj[k]))
				}
				rows = append(rows, fmt.Sprintf("%s  %s", prefix, strings.Join(vals, ",")))
			}
			return strings.Join(rows, "\n")
		}
		return fmt.Sprintf("%s%s[%d]: %v", prefix, keyName, len(v), v)
	default:
		if keyName != "" {
			return fmt.Sprintf("%s%s: %v", prefix, keyName, v)
		}
		return fmt.Sprintf("%v", v)
	}
}

func getUniformKeys(arr []interface{}) ([]string, bool) {
	if len(arr) == 0 {
		return nil, false
	}
	first, ok := arr[0].(map[string]interface{})
	if !ok {
		return nil, false
	}
	var keys []string
	for k := range first {
		keys = append(keys, k)
	}
	for i := 1; i < len(arr); i++ {
		item, ok := arr[i].(map[string]interface{})
		if !ok || len(item) != len(first) {
			return nil, false
		}
		for _, k := range keys {
			if _, exists := item[k]; !exists {
				return nil, false
			}
		}
	}
	return keys, true
}

func startHeartbeat(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		connMutex.Lock()
		if targetConn != nil {
			_, err := targetConn.Write([]byte("\n"))
			metrics.mu.Lock()
			if err != nil {
				log.Printf("[SRE] Heartbeat failed. Cleaning connection.")
				targetConn.Close()
				targetConn = nil
				metrics.HeartbeatFail++
			} else {
				metrics.HeartbeatSuccess++
			}
			metrics.mu.Unlock()
		}
		connMutex.Unlock()
	}
}

// observability Go routine: Prints an executive summary every 60 seconds
func logMetrics(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		metrics.mu.Lock()
		log.Printf("[METRICS] Req: %d | HB Success: %d | HB Fail: %d", 
			metrics.TotalRequests, metrics.HeartbeatSuccess, metrics.HeartbeatFail)
		metrics.mu.Unlock()
	}
}

func getTargetConn(addr string) (net.Conn, error) {
	connMutex.Lock()
	defer connMutex.Unlock()
	if targetConn != nil {
		return targetConn, nil
	}
	dialer := net.Dialer{Timeout: 5 * time.Second}
	c, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	targetConn = c
	return targetConn, nil
}

func handleConnection(localConn net.Conn, targetAddr string) {
	defer localConn.Close()
	var rawData interface{}
	if err := json.NewDecoder(localConn).Decode(&rawData); err != nil {
		return
	}
	payload := convertToToon(rawData, 0, "") + "\n"
	remote, err := getTargetConn(targetAddr)
	if err != nil {
		log.Printf("[ERR] Target %s down", targetAddr)
		return
	}
	connMutex.Lock()
	_, err = remote.Write([]byte(payload))
	metrics.mu.Lock()
	if err != nil {
		targetConn.Close()
		targetConn = nil
		log.Printf("[ERR] Persistent pipe broken")
	} else {
		metrics.TotalRequests++
	}
	metrics.mu.Unlock()
	connMutex.Unlock()
}

func main() {
	lPort := flag.String("l", "8080", "Listen port")
	tAddr := flag.String("t", "127.0.0.1:9999", "Target address")
	flag.Parse()

	go startHeartbeat(30 * time.Second)
	go logMetrics(60 * time.Second)

	ln, err := net.Listen("tcp", ":"+*lPort)
	if err != nil {
		log.Fatalf("Fatal: %v", err)
	}

	log.Printf("TOON Proxy + Observability: :%s -> %s", *lPort, *tAddr)
	for {
		conn, _ := ln.Accept()
		go handleConnection(conn, *tAddr)
	}
}