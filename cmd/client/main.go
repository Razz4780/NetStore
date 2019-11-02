package client

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
)

func createTmpDirectory() error {
	if _, err := os.Stat("tmp"); os.IsNotExist(err) {
		return os.Mkdir("tmp", os.ModeDir)
	} else {
		return nil
	}
}

type ServerError struct {
	message string
}

func (serror ServerError) Error() string {
	return serror.message
}

func getFilenames(readwriter *bufio.ReadWriter) ([]string, error) {
	request := make([]byte, 2)
	binary.BigEndian.PutUint16(request, 1)
	if _, err := readwriter.Write(request); err != nil {
		return nil, err
	}
	if err := readwriter.Flush(); err != nil {
		return nil, err
	}
	response := make([]byte, 6)
	if _, err := io.ReadFull(readwriter, response); err != nil {
		return nil, err
	}
	responseCode := binary.BigEndian.Uint16(response[:2])
	if responseCode != 1 {
		return nil, ServerError{"invalid message received"}
	}
	filenamesLength := binary.BigEndian.Uint32(response[2:6])
	filenames := make([]string, 0, 64)
	var bytesRead uint32 = 0
	for bytesRead < filenamesLength {
		filename, err := readwriter.ReadString(0)
		if err != nil {
			return nil, err
		}
		bytesRead += uint32(len(filename))
		filenames = append(filenames, filename[:len(filename)-1])
	}
	return filenames, nil
}

func getNumberInRange(message string, min, max uint32) uint32 {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(message)
		text, _ := reader.ReadString('\n')
		parsedNumber, err := strconv.ParseUint(text, 10, 32)
		number := uint32(parsedNumber)
		if err != nil {
			fmt.Println("Parsing failed: ", err)
		} else if number < min {
			fmt.Println("Value to small, minimal value is ", min)
		} else if number > max {
			fmt.Println("Value to big, maximal value is ", max)
		} else {
			return number
		}
	}
}

func getFileChunk(readwriter *bufio.ReadWriter, filename string, offset, chunkSize uint32) error {
	request := make([]byte, 12)
	binary.BigEndian.PutUint16(request, 2)
	binary.BigEndian.PutUint32(request[2:], offset)
	binary.BigEndian.PutUint32(request[6:], chunkSize)
	binary.BigEndian.PutUint16(request[10:], uint16(len(filename)))
	if _, err := readwriter.Write(request); err != nil {
		return err
	}
	if _, err := readwriter.WriteString(filename); err != nil {
		return err
	}
	if err := readwriter.Flush(); err != nil {
		return err
	}
	response := make([]byte, 6)
	if _, err := io.ReadFull(readwriter, response); err != nil {
		return err
	}
	responseType := binary.BigEndian.Uint16(response[:2])
	if responseType == 2 {
		refusalType := binary.BigEndian.Uint32(response[:4])
		if refusalType == 1 {
			return ServerError{"wrong filename"}
		} else if refusalType == 2 {
			return ServerError{"invalid offset"}
		} else if refusalType == 3 {
			return ServerError{"invalid chunk size"}
		} else {
			return ServerError{"invalid message received"}
		}
	} else if responseType == 3 {
		chunkSize = binary.BigEndian.Uint32(response[:4])
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0)
		if err != nil {
			return err
		}
		if _, err := file.Seek(int64(offset), 0); err != nil {
			return err
		}
		_, err = io.CopyN(file, readwriter.Reader, int64(chunkSize))
		return err
	} else {
		return ServerError{"invalid message received"}
	}
}

func main() {
	serverAddress := flag.String("server", "127.0.0.1:5551", "server address with port number")
	flag.Parse()
	if err := createTmpDirectory(); err != nil {
		log.Fatal("Could not create directory 'tmp': ", err)
	}
	conn, err := net.Dial("tcp", *serverAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	server := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	filenames, err := getFilenames(server)
	if err != nil {
		log.Fatal("Could not get filenames: ", err)
	}
	fmt.Println("Available files:")
	for i, filename := range filenames {
		fmt.Println(i, filename)
	}
	fileNumber := getNumberInRange("Choose file number: ", 1, uint32(len(filenames))-1)
	offset := getNumberInRange("Choose chunk offset: ", 0, ^uint32(0))
	chunkSize := getNumberInRange("Choose chunk size: ", 1, ^uint32(0))
	if err := getFileChunk(server, filenames[fileNumber], offset, chunkSize); err != nil {
		log.Fatal("Could not get file chunk: ", err)
	}
}
