package server

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
)

func sendFilenames(writer *bufio.Writer, filenames []string) error {
	filenamesLength := 0
	for _, filename := range filenames {
		filenamesLength += len(filename)
	}
	response := make([]byte, 6)
	binary.BigEndian.PutUint16(response, 1)
	binary.BigEndian.PutUint32(response[2:], uint32(filenamesLength+len(filenames)))
	if _, err := writer.Write(response); err != nil {
		return err
	}
	for _, filename := range filenames {
		if _, err := writer.WriteString(filename); err != nil {
			return err
		}
		if err := writer.WriteByte(0); err != nil {
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return nil
}

func sendFileChunk(readwriter *bufio.ReadWriter, filenames []string) error {
	request := make([]byte, 10)
	if _, err := io.ReadFull(readwriter, request); err != nil {
		return err
	}
	offset := int64(binary.BigEndian.Uint32(request))
	chunkSize := int64(binary.BigEndian.Uint32(request[4:]))
	if chunkSize == 0 {
		response := make([]byte, 6)
		binary.BigEndian.PutUint16(response, 2)
		binary.BigEndian.PutUint32(response[2:], 3)
		if _, err := readwriter.Write(response); err != nil {
			return err
		}
		return readwriter.Flush()
	}
	filenameLen := int(binary.BigEndian.Uint16(request[8:]))
	requestedFile := make([]byte, filenameLen)
	if _, err := io.ReadFull(readwriter, requestedFile); err != nil {
		return err
	}
	requestedFileName := string(requestedFile)
	for _, filename := range filenames {
		if filename == requestedFileName {
			file, err := os.Open(filename)
			if err != nil {
				return err
			}
			stat, err := file.Stat()
			if err != nil {
				return err
			}
			if stat.Size() >= offset {
				response := make([]byte, 6)
				binary.BigEndian.PutUint16(response, 2)
				binary.BigEndian.PutUint32(response[2:], 2)
				if _, err := readwriter.Write(response); err != nil {
					return err
				}
				return readwriter.Flush()
			}
			fileReader := io.LimitReader(file, chunkSize)
			if _, err := io.Copy(readwriter, fileReader); err != nil {
				return err
			}
			return readwriter.Flush()
		}
	}
	response := make([]byte, 6)
	binary.BigEndian.PutUint16(response, 2)
	binary.BigEndian.PutUint32(response[2:], 1)
	if _, err := readwriter.Write(response); err != nil {
		return err
	}
	return readwriter.Flush()
}

func handleConnection(conn net.Conn, filenames []string) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println("Handling connection failed: ", err)
		}
	}()
	readwriter := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	request := make([]byte, 2)
	if _, err := io.ReadFull(readwriter, request); err != nil {
		log.Println("Handling connection failed: ", err)
	}
	requestType := binary.BigEndian.Uint16(request)
	if requestType == 1 {
		if err := sendFilenames(readwriter.Writer, filenames); err != nil {
			log.Println("Handling connection failed: ", err)
		}
	} else if requestType == 2 {
		if err := sendFileChunk(readwriter, filenames); err != nil {
			log.Println("Handling conection failed: ", err)
		}
	} else {
		log.Println("Invalid message received")
	}
}

func main() {
	dirpath := flag.String("dir", ".", "path to files directory")
	port := flag.Uint("port", 5551, "port number")
	flag.Parse()
	if *port > uint(^uint16(0)) {
		log.Fatal("Invalid port number specified: ", port)
	}
	files, err := ioutil.ReadDir(*dirpath)
	if err != nil {
		log.Fatal("Could not read files directory: ", err)
	}
	filenames := make([]string, 0, len(files))
	for _, file := range files {
		if file.Mode().IsRegular() {
			filenames = append(filenames, file.Name())
		}
	}
	ln, err := net.Listen("tcp", fmt.Sprint(":", port))
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Handling connection failed: ", err)
		}
		go handleConnection(conn, filenames)
	}
}
