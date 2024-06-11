package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", ":4221")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Listening for http://localhost:4221")

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)
	}
}
func handleConnection(conn net.Conn) {
	buf := make([]byte, 5000)
	conn.Read(buf)
	rawReq := string(buf[:])

	req := strings.Split(rawReq, "\r\n")
	path := strings.Split(req[0], " ")
	splitPath, err := splitURL(path[1])

	if err != nil {
		splitPath = append(splitPath, "/")
	}

	log.Printf("%s", req[1])
	for i, line := range req {
		log.Println(i, line)
	}

	switch splitPath[0] {

	case "/echo":
		sliceString := strings.Split(path[1], "/echo")
		restString := sliceString[1]
		contLen := len(restString) - 1
		if contLen < 0 {
			contLen = 0
		}
		if restString != "" {
			restString = restString[1:]
		}
		conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprint(contLen) + "\r\n\r\n" + restString))
	case "/user-agent":
		var headerContents []string
		var headerContent string
		var err error
		for _, line := range req {
			if strings.HasPrefix(line, "User-Agent:") {
				headerContents = strings.Split(line, " ")
				headerContent = headerContents[1]
				err = nil
				break
			}
			err = errors.New("no user-agent header")
		}
		if err != nil {
			headerContent = ""
		}
		contentLen := len(headerContent)
		conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + fmt.Sprint(contentLen) + "\r\n\r\n" + headerContent))
	case "/":
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	default:
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))

	}

	defer conn.Close()
}

func splitURL(url string) ([]string, error) {
	var segments []string

	if url == "/" {
		return segments, errors.New("empty url")
	}

	parts := strings.Split(url, "/")
	for _, part := range parts {
		if part != "" {
			segments = append(segments, "/"+part)
		}
	}

	return segments, nil
}
