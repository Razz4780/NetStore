package internal

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

const (
	DefaultPort             uint16 = 5551
	RequestTypeFilenames    uint16 = 1
	RequestTypeChunk        uint16 = 2
	ResponseTypeFilenames   uint16 = 1
	ResponseTypeRefusal     uint16 = 2
	ResponseTypeChunk       uint16 = 3
	FilenamesDelimiter      byte   = 0
	RefusalCauseBadFilename uint32 = 1
	RefusalCauseBadOffset   uint32 = 2
	RefusalCauseBadSize     uint32 = 3
)

func readUint16(reader io.Reader) (uint16, error) {
	buff := make([]byte, 2)
	if _, err := io.ReadFull(reader, buff); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buff), nil
}

func ReadRequestType(reader io.Reader) (uint16, error) {
	requestType, err := readUint16(reader)
	if err != nil {
		return 0, err
	}
	if requestType != RequestTypeFilenames && requestType != RequestTypeChunk {
		return 0, fmt.Errorf("unknown request type: %d", requestType)
	}
	return requestType, nil
}

type ChunkRequest struct {
	Offset   uint32
	Size     uint32
	Filename []byte
}

func ReadChunkRequest(reader io.Reader) (ChunkRequest, error) {
	buff := make([]byte, 10)
	if _, err := io.ReadFull(reader, buff); err != nil {
		return ChunkRequest{}, err
	}
	offset := binary.BigEndian.Uint32(buff)
	size := binary.BigEndian.Uint32(buff[4:])
	filenameLen := binary.BigEndian.Uint16(buff[8:])
	filename := make([]byte, filenameLen)
	if _, err := io.ReadFull(reader, filename); err != nil {
		return ChunkRequest{}, err
	}
	return ChunkRequest{offset, size, filename}, nil
}

func WriteFilenamesRequest(writer io.Writer) error {
	buff := make([]byte, 2)
	binary.BigEndian.PutUint16(buff, RequestTypeFilenames)
	_, err := writer.Write(buff)
	return err
}

func WriteChunkRequest(writer io.Writer, offset, size uint32, filename []byte) error {
	buff := make([]byte, 12)
	binary.BigEndian.PutUint16(buff, RequestTypeChunk)
	binary.BigEndian.PutUint32(buff[2:], offset)
	binary.BigEndian.PutUint32(buff[6:], size)
	binary.BigEndian.PutUint16(buff[10:], uint16(len(filename)))
	buffWriter := bufio.NewWriter(writer)
	if _, err := buffWriter.Write(buff); err != nil {
		return err
	}
	if _, err := buffWriter.Write(filename); err != nil {
		return err
	}
	return buffWriter.Flush()
}

func ReadResponseType(reader io.Reader) (uint16, error) {
	responseType, err := readUint16(reader)
	if err != nil {
		return 0, nil
	}
	if responseType != ResponseTypeFilenames &&
		responseType != ResponseTypeRefusal &&
		responseType != ResponseTypeChunk {
		return 0, fmt.Errorf("unknown request type: %d", responseType)
	}
	return responseType, nil
}

type FilenamesResponse struct {
	Filenames [][]byte
}

func ReadFilenamesResponse(reader io.Reader) (FilenamesResponse, error) {
	buff := make([]byte, 4)
	if _, err := io.ReadFull(reader, buff); err != nil {
		return FilenamesResponse{}, err
	}
	filenamesFieldLen := binary.BigEndian.Uint32(buff)
	buffReader := bufio.NewReader(io.LimitReader(reader, int64(filenamesFieldLen)))
	filenames := make([][]byte, 0, 32)
	var processedBytes uint32 = 0
	for processedBytes < filenamesFieldLen {
		filename, err := buffReader.ReadBytes(FilenamesDelimiter)
		if err != nil {
			return FilenamesResponse{}, err
		}
		processedBytes += uint32(len(filename))
		filenames = append(filenames, filename[:len(filename)-1])
	}
	return FilenamesResponse{filenames}, nil
}

func WriteFilenamesResponse(writer io.Writer, filenames [][]byte) error {
	buff := make([]byte, 6)
	binary.BigEndian.PutUint16(buff, ResponseTypeFilenames)
	var filenamesFieldLen uint32 = 0
	for _, file := range filenames {
		filenamesFieldLen += uint32(len(file))
		filenamesFieldLen += 1
	}
	binary.BigEndian.PutUint32(buff[2:], filenamesFieldLen)
	buffWriter := bufio.NewWriter(writer)
	if _, err := buffWriter.Write(buff); err != nil {
		return err
	}
	for _, file := range filenames {
		if _, err := buffWriter.Write(file); err != nil {
			return err
		}
		if err := buffWriter.WriteByte(FilenamesDelimiter); err != nil {
			return err
		}
	}
	return buffWriter.Flush()
}

func ReadRefusal(reader io.Reader) (uint32, error) {
	buff := make([]byte, 4)
	if _, err := io.ReadFull(reader, buff); err != nil {
		return 0, err
	}
	refusalCause := binary.BigEndian.Uint32(buff)
	if refusalCause != RefusalCauseBadFilename &&
		refusalCause != RefusalCauseBadOffset &&
		refusalCause != RefusalCauseBadSize {
		return 0, fmt.Errorf("unknown refusal cause: %d", refusalCause)
	}
	return refusalCause, nil
}

func ReadChunkResponse(reader io.Reader, writer io.Writer) (uint32, error) {
	buff := make([]byte, 4)
	if _, err := io.ReadFull(reader, buff); err != nil {
		return 0, err
	}
	chunkSize := binary.BigEndian.Uint32(buff)
	if _, err := io.CopyN(writer, reader, int64(chunkSize)); err != nil {
		return 0, err
	}
	return chunkSize, nil
}

func WriteChunkResponse(writer io.Writer, reader io.Reader, chunkSize uint32) error {
	buff := make([]byte, 4)
	binary.BigEndian.PutUint32(buff, chunkSize)
	buffWriter := bufio.NewWriter(writer)
	if _, err := buffWriter.Write(buff); err != nil {
		return err
	}
	if _, err := io.CopyN(buffWriter, reader, int64(chunkSize)); err != nil {
		return err
	}
	return buffWriter.Flush()
}

func WriteRefusal(writer io.Writer, cause uint32) error {
	buff := make([]byte, 6)
	binary.BigEndian.PutUint16(buff, ResponseTypeRefusal)
	binary.BigEndian.PutUint32(buff[2:], cause)
	_, err := writer.Write(buff)
	return err
}
