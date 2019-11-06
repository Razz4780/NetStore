package internal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
)

func TestReadUint16FromValidReader(t *testing.T) {
	numbers := []uint16{0, 1, 2, 3, ^uint16(0)}
	for _, number := range numbers {
		t.Run(fmt.Sprint("reading number ", number), func(t *testing.T) {
			buff := make([]byte, 2)
			binary.BigEndian.PutUint16(buff, number)
			result, err := readUint16(bytes.NewReader(buff))
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if result != number {
				t.Fatal("read", result, ", expected", number)
			}
		})
	}
}

func TestReadUint16FromReaderTooShort(t *testing.T) {
	t.Run(fmt.Sprint("0 bytes available"), func(t *testing.T) {
		buff := make([]byte, 0)
		_, err := readUint16(bytes.NewReader(buff))
		if err == nil {
			t.Fatal("expected error not returned")
		}
	})

	t.Run(fmt.Sprint("1 byte available"), func(t *testing.T) {
		buff := make([]byte, 1)
		_, err := readUint16(bytes.NewReader(buff))
		if err == nil {
			t.Fatal("expected error not returned")
		}
	})
}

func TestReadUint16ReadsExactlyTwoBytes(t *testing.T) {
	buff := []byte{0, 1, 2}
	reader := bytes.NewReader(buff)
	_, err := readUint16(reader)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	nextByte, err := reader.ReadByte()
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if nextByte != buff[2] {
		t.Fatal("read", nextByte, ", expected", buff[2])
	}
}

func TestReadRequestTypeOfValidValues(t *testing.T) {
	validTypes := []uint16{RequestTypeFilenames, RequestTypeChunk}
	for _, requestType := range validTypes {
		t.Run(fmt.Sprint("reading type ", requestType), func(t *testing.T) {
			buff := make([]byte, 2)
			binary.BigEndian.PutUint16(buff, requestType)
			result, err := ReadRequestType(bytes.NewReader(buff))
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if result != requestType {
				t.Fatal("read", result, ", expected", requestType)
			}
		})
	}
}

func TestReadRequestTypeOfInvalidValues(t *testing.T) {
	invalidValues := []uint16{RequestTypeChunk + 1, ^uint16(0)}
	for _, value := range invalidValues {
		t.Run(fmt.Sprint("reading invalid type ", value), func(t *testing.T) {
			buff := make([]byte, 2)
			binary.BigEndian.PutUint16(buff, value)
			_, err := ReadRequestType(bytes.NewReader(buff))
			if err == nil {
				t.Fatal("expected error not returned")
			}
		})
	}
}

func TestReadRequestTypeReadsExactlyTwoBytes(t *testing.T) {
	buff := make([]byte, 2)
	binary.BigEndian.PutUint16(buff, RequestTypeFilenames)
	buff = append(buff, ^byte(0))
	reader := bytes.NewReader(buff)
	_, err := ReadRequestType(reader)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	nextByte, err := reader.ReadByte()
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if nextByte != buff[2] {
		t.Fatal("read", nextByte, ", expected", buff[2])
	}
}

func TestReadChunkRequestOfValidRequests(t *testing.T) {
	dataSets := []struct {
		offset   uint32
		size     uint32
		filename string
	}{
		{0, 0, "filename"},
		{2, 13, "qwerty"},
		{^uint32(0), ^uint32(0), "asdf"},
	}
	for i, dataSet := range dataSets {
		t.Run(fmt.Sprint("dataset ", i), func(t *testing.T) {
			buff := make([]byte, 10)
			binary.BigEndian.PutUint32(buff, dataSet.offset)
			binary.BigEndian.PutUint32(buff[4:], dataSet.size)
			binary.BigEndian.PutUint16(buff[8:], uint16(len(dataSet.filename)))
			buff = append(buff, dataSet.filename...)
			result, err := ReadChunkRequest(bytes.NewReader(buff))
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if result.Size != dataSet.size {
				t.Error("read size", result.Size, ", expected", dataSet.size)
			}
			if result.Offset != dataSet.offset {
				t.Error("read offset", result.Offset, ", expected", dataSet.offset)
			}
			if string(result.Filename) != dataSet.filename {
				t.Error("read filename", string(result.Filename), ", expected", dataSet.offset)
			}
		})
	}
}

func TestReadChunkRequestFromReaderTooShort(t *testing.T) {
	t.Run("0 bytes available", func(t *testing.T) {
		buff := make([]byte, 0)
		_, err := ReadChunkRequest(bytes.NewReader(buff))
		if err == nil {
			t.Fatal("expected error not returned")
		}
	})

	t.Run("too few bytes available", func(t *testing.T) {
		buff := make([]byte, 10)
		binary.BigEndian.PutUint16(buff[8:], uint16(len("filename")))
		buff = append(buff, "filename"...)
		_, err := ReadChunkRequest(bytes.NewReader(buff[:len(buff)-1]))
		if err == nil {
			t.Fatal("expected error not returned")
		}
	})
}

func TestReadChunkRequestReadsOnlyRequest(t *testing.T) {
	buff := make([]byte, 10)
	binary.BigEndian.PutUint16(buff[8:], uint16(len("filename")))
	buff = append(buff, "filename"...)
	buff = append(buff, ^byte(0))
	reader := bytes.NewReader(buff)
	_, err := ReadChunkRequest(reader)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	nextByte, err := reader.ReadByte()
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if nextByte != buff[len(buff)-1] {
		t.Fatal("read", nextByte, ", expected", buff[len(buff)-1])
	}
}

func TestWriteChunkRequest(t *testing.T) {
	dataSets := []struct {
		offset   uint32
		size     uint32
		filename string
	}{
		{0, 0, "asdf"},
		{1, 2, "random filename"},
		{4, 12, "name"},
	}
	for i, dataSet := range dataSets {
		t.Run(fmt.Sprint("dataset ", i), func(t *testing.T) {
			buff := make([]byte, 0)
			writer := bytes.NewBuffer(buff)
			err := WriteChunkRequest(writer, dataSet.offset, dataSet.size, []byte(dataSet.filename))
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			buff = writer.Bytes()
			requestType := binary.BigEndian.Uint16(buff)
			offset := binary.BigEndian.Uint32(buff[2:])
			size := binary.BigEndian.Uint32(buff[6:])
			filenameLen := binary.BigEndian.Uint16(buff[10:])
			filename := string(buff[12:])
			if requestType != RequestTypeChunk {
				t.Error("read request type", requestType, ", expected", RequestTypeChunk)
			}
			if offset != dataSet.offset {
				t.Error("read offset", offset, ", expected", dataSet.offset)
			}
			if size != dataSet.size {
				t.Error("read size", size, ", expected", dataSet.size)
			}
			if filenameLen != uint16(len(dataSet.filename)) {
				t.Error("read filename length", filenameLen, ", expected", len(dataSet.filename))
			}
			if filename != dataSet.filename {
				t.Error("read filename", filename, ", expected", dataSet.filename)
			}
		})
	}
}

func TestReadResponseTypeOfValidValues(t *testing.T) {
	validTypes := []uint16{ResponseTypeFilenames, ResponseTypeRefusal, ResponseTypeChunk}
	for _, responseType := range validTypes {
		t.Run(fmt.Sprint("reading type ", responseType), func(t *testing.T) {
			buff := make([]byte, 2)
			binary.BigEndian.PutUint16(buff, responseType)
			result, err := ReadResponseType(bytes.NewReader(buff))
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if result != responseType {
				t.Fatal("read", result, ", expected", responseType)
			}
		})
	}
}

func TestReadResponseTypeOfInvalidValues(t *testing.T) {
	invalidValues := []uint16{ResponseTypeChunk + 1, ^uint16(0)}
	for _, value := range invalidValues {
		buff := make([]byte, 2)
		binary.BigEndian.PutUint16(buff, value)
		_, err := ReadResponseType(bytes.NewReader(buff))
		if err == nil {
			t.Fatal("expected error not returned")
		}
	}
}

func TestReadResponseTypeReadsExactlyTwoBytes(t *testing.T) {
	buff := make([]byte, 2)
	binary.BigEndian.PutUint16(buff, ResponseTypeFilenames)
	buff = append(buff, ^byte(0))
	reader := bytes.NewReader(buff)
	_, err := ReadResponseType(reader)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	nextByte, err := reader.ReadByte()
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if nextByte != buff[2] {
		t.Fatal("read", nextByte, ", expected", buff[2])
	}
}

func TestReadFilenamesResponseOfValidResponses(t *testing.T) {
	dataSets := [][]string{
		{"asdf", "qwer", "zxcv"},
		{},
		{"aaaaa"},
	}
	for i, filenames := range dataSets {
		t.Run(fmt.Sprint("dataset ", i), func(t *testing.T) {
			filenamesLen := 0
			for _, filename := range filenames {
				filenamesLen += len(filename)
			}
			buff := make([]byte, 4, 4+filenamesLen+len(filenames))
			binary.BigEndian.PutUint32(buff, uint32(filenamesLen+len(filenames)))
			for _, filename := range filenames {
				buff = append(buff, filename...)
				buff = append(buff, FilenamesDelimiter)
			}
			result, err := ReadFilenamesResponse(bytes.NewReader(buff))
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if len(result.Filenames) != len(filenames) {
				t.Fatal("read", len(result.Filenames), "filenames, expected", len(filenames))
			}
			for i := 0; i < len(filenames); i++ {
				if string(result.Filenames[i]) != filenames[i] {
					t.Error(string(result.Filenames[i]), "does not match", filenames[i])
				}
			}
		})
	}
}

func TestReadFilenamesResponseOfInvalidResponses(t *testing.T) {
	t.Run("response not ending with delimiter", func(t *testing.T) {
		filename := "filename"
		buff := make([]byte, 4, 4+len(filename))
		binary.BigEndian.PutUint32(buff, uint32(len(filename)))
		buff = append(buff, filename...)
		_, err := ReadFilenamesResponse(bytes.NewReader(buff))
		if err == nil {
			t.Fatal("expected error not returned")
		}
	})

	t.Run("response too short", func(t *testing.T) {
		filename := "filename"
		buff := make([]byte, 4, 4+len(filename)+1)
		binary.BigEndian.PutUint32(buff, uint32(len(filename)+2))
		buff = append(buff, filename...)
		buff = append(buff, FilenamesDelimiter)
		_, err := ReadFilenamesResponse(bytes.NewReader(buff))
		if err == nil {
			t.Fatal("expected error not returned")
		}
	})
}

func TestReadFilenamesResponseReadsOnlyResponse(t *testing.T) {
	filename := "filename"
	buff := make([]byte, 4, 4+len(filename)+1+1)
	binary.BigEndian.PutUint32(buff, uint32(len(filename))+1)
	buff = append(buff, filename...)
	buff = append(buff, FilenamesDelimiter, ^byte(0))
	reader := bytes.NewBuffer(buff)
	_, err := ReadFilenamesResponse(reader)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	nextByte, err := reader.ReadByte()
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if nextByte != buff[len(buff)-1] {
		t.Fatal("read", nextByte, ", expected", buff[len(buff)-1])
	}
}

func TestReadRefusalOfValidValues(t *testing.T) {
	validCauses := []uint32{RefusalCauseBadFilename, RefusalCauseBadOffset, RefusalCauseBadSize}
	for _, cause := range validCauses {
		t.Run(fmt.Sprint("reading cause ", cause), func(t *testing.T) {
			buff := make([]byte, 4)
			binary.BigEndian.PutUint32(buff, cause)
			result, err := ReadRefusal(bytes.NewReader(buff))
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if result != cause {
				t.Fatal("read", result, ", expected", cause)
			}
		})
	}
}

func TestReadRefusalOfInvalidValues(t *testing.T) {
	invalidValues := []uint32{RefusalCauseBadSize + 1, ^uint32(0)}
	for _, value := range invalidValues {
		t.Run(fmt.Sprint("reading invalid value ", value), func(t *testing.T) {
			buff := make([]byte, 4)
			binary.BigEndian.PutUint32(buff, value)
			_, err := ReadRefusal(bytes.NewReader(buff))
			if err == nil {
				t.Fatal("expected error not returned")
			}
		})
	}
}

func TestReadRefusalReadsExactlyFourBytes(t *testing.T) {
	buff := make([]byte, 5)
	binary.BigEndian.PutUint32(buff, RefusalCauseBadSize)
	buff[4] = ^byte(0)
	reader := bytes.NewReader(buff)
	_, err := ReadRefusal(reader)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	nextByte, err := reader.ReadByte()
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if nextByte != buff[4] {
		t.Fatal("read", nextByte, ", expected", buff[4])
	}
}
