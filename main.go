package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", ":4221")
	if err != nil {
		log.Fatal(err)
	}

	defer l.Close()
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
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	rawReq := string(buf[:n])

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

	if strings.HasPrefix(req[0], "GET") {
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

		case "/files":
			status := "HTTP/1.1 200 OK"
			if len(splitPath) <= 1 {
				splitPath = append(splitPath, "")
			}
			filePath, err := findFile(os.Args, splitPath[1])
			if err != nil {
				log.Println(err)
				status = "HTTP/1.1 404 Not Found"
			}

			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				log.Println(err)
				status = "HTTP/1.1 404 Not Found"
			}

			contentLen := len(string(fileContent[:]))

			log.Println(contentLen)

			response := fmt.Sprintf("%s\r\nContent-Type: application/octet-stream\r\nContent-Length: %s\r\n\r\n%s", status, fmt.Sprint(contentLen), string(fileContent[:]))

			conn.Write([]byte(response))

		case "/":
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

		default:
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))

		}

	} else if strings.HasPrefix(req[0], "POST") {
		switch splitPath[0] {
		case "/":
			log.Println("compas posteando ahi")
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

		case "/files":
			response := "HTTP/1.1 201 Created\r\n\r\n"
			if len(splitPath) <= 1 {
				response = "HTTP/1.1 404 Not Found\r\n\r\n"
			}

			dirPath, err := findDir(os.Args)
			if err != nil {
				response = "HTTP/1.1 404 Not Found\r\n\r\n"
			}

			fileContent := []byte(req[len(req)-1])

			err = os.WriteFile(dirPath+splitPath[1], fileContent, 0777)
			if err != nil {
				response = "HTTP/1.1 404 Not Found\r\n\r\n"
			}

			conn.Write([]byte(response))

		default:
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}
	}

	conn.Close()
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

func findFile(args []string, fileName string) (string, error) {
	if args[1] != "--directory" {
		return "", errors.New("directory flag not provided")
	}
	if len(args) <= 2 {
		return "", errors.New("no path provided")
	}

	noSlashName := fileName[1:]
	items, err := os.ReadDir(args[2])
	if err != nil {
		return "", err
	}

	for _, item := range items {
		if item.Name() == noSlashName {
			return args[2] + fileName[1:], nil
		}
	}

	return "", errors.New("not founded file")
}

func findDir(args []string) (string, error) {
	if args[1] != "--directory" {
		return "", errors.New("directory flag not provided")
	}

	if len(args) <= 2 {
		return "", errors.New("no path provided")
	}

	_, err := os.ReadDir(args[2])

	if err != nil {
		return "", errors.New("folder not found")
	}

	return args[2], nil
}
