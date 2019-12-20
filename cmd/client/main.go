package client

import (
	"NetStore/internal"
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strconv"
)

func getFilenames(readwriter *bufio.ReadWriter) ([][]byte, error) {
	if err := internal.WriteFilenamesRequest(readwriter); err != nil {
		return nil, err
	}
	if err := readwriter.Flush(); err != nil {
		return nil, err
	}
	response, err := internal.ReadFilenamesResponse(readwriter)
	if err != nil {
		return nil, err
	}
	return response.Filenames, nil
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

func getFileChunk(readwriter *bufio.ReadWriter, filename []byte, offset, chunkSize uint32) (rerr error) {
	if err := internal.WriteChunkRequest(readwriter, offset, chunkSize, filename); err != nil {
		return err
	}
	if err := readwriter.Flush(); err != nil {
		return err
	}
	filepath := path.Join(internal.ReceivedFilesDir, string(filename))
	file, err := internal.OpenFile(filepath, int64(offset), os.O_CREATE|os.O_WRONLY)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil && rerr == nil {
			rerr = err
		}
	}()
	_, err = internal.ReadChunkResponse(readwriter, file)
	return err
}

func main() {
	serverAddress := flag.String(
		"server",
		fmt.Sprint("127.0.0.1:", internal.DefaultPort),
		"server address with port number",
	)
	flag.Parse()
	if err := internal.CreateReceivedFilesDir(); err != nil {
		log.Fatal("Could not create directory", internal.ReceivedFilesDir, ":", err)
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
	if len(filenames) == 0 {
		fmt.Println("No files available.")
		return
	}
	fmt.Println("Available files:")
	for i, filename := range filenames {
		fmt.Println(i, string(filename))
	}
	fileNumber := getNumberInRange("Choose file number: ", 1, uint32(len(filenames))-1)
	offset := getNumberInRange("Choose chunk offset: ", 0, ^uint32(0))
	chunkSize := getNumberInRange("Choose chunk size: ", 1, ^uint32(0))
	if err := getFileChunk(server, filenames[fileNumber-1], offset, chunkSize); err != nil {
		log.Fatal("Could not get file chunk: ", err)
	}
}
