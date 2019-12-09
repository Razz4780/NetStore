package server

import (
	"NetStore/internal"
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"syscall"
)

func handleChunkRequest(readWriter io.ReadWriter, files []internal.FileInfo) error {
	request, err := internal.ReadChunkRequest(readWriter)
	if err != nil {
		return err
	}
	if request.Size == 0 {
		return internal.WriteRefusal(readWriter, internal.RefusalCauseBadSize)
	}
	for _, fileInfo := range files {
		if bytes.Equal(request.Filename, fileInfo.Name) {
			if uint64(request.Offset) >= fileInfo.Size {
				return internal.WriteRefusal(readWriter, internal.RefusalCauseBadFilename)
			}
			file, err := internal.OpenFile(string(fileInfo.Name), int64(request.Offset), syscall.O_RDONLY)
			if err != nil {
				return err
			}
			if err := internal.WriteChunkResponse(readWriter, file, request.Size); err != nil {
				_ = file.Close()
				return err
			}
			return file.Close()
		}
	}
	return internal.WriteRefusal(readWriter, internal.RefusalCauseBadFilename)
}

func handleConnection(conn net.Conn, files []internal.FileInfo) (rerr error) {
	defer func() {
		if err := conn.Close(); err != nil && rerr != nil {
			rerr = err
		}
	}()
	readWriter := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	requestType, err := internal.ReadRequestType(readWriter)
	if err != nil {
		return err
	}
	filenames := make([][]byte, 0, len(files))
	for _, fileInfo := range files {
		filenames = append(filenames, fileInfo.Name)
	}
	if requestType == internal.RequestTypeFilenames {
		if err := internal.WriteFilenamesResponse(readWriter, filenames); err != nil {
			return err
		}
		if err := readWriter.Flush(); err != nil {
			return err
		}
	} else if requestType == internal.RequestTypeChunk {
		if err := handleChunkRequest(readWriter, files); err != nil {
			return err
		}
		if err := readWriter.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	dirpath := flag.String("dir", ".", "path to files directory")
	port := flag.Uint("port", 5551, "port number")
	flag.Parse()
	if *port > uint(^uint16(0)) {
		log.Fatal("Invalid port number specified: ", *port)
	}
	files, err := internal.IndexFiles(*dirpath)
	if err != nil {
		log.Fatal("Could not read files directory: ", err)
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
		go func() {
			if err := handleConnection(conn, files); err != nil {
				log.Println("Handling connection failed: ", err)
			}
		}()
	}
}
