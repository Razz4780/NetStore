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
	if reader.Len() != 1 {
		t.Fatal(reader.Len(), "bytes consumed")
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
	if reader.Len() != 1 {
		t.Fatal(reader.Len(), "bytes consumed")
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
	buff = append(buff, 0)
	reader := bytes.NewReader(buff)
	_, err := ReadChunkRequest(reader)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if reader.Len() != 1 {
		t.Fatal(reader.Len(), "bytes consumed")
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
	buff := make([]byte, 3)
	binary.BigEndian.PutUint16(buff, ResponseTypeFilenames)
	reader := bytes.NewReader(buff)
	_, err := ReadResponseType(reader)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if reader.Len() != 1 {
		t.Fatal(reader.Len(), "bytes consumed")
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
	buff = append(buff, FilenamesDelimiter, 0)
	reader := bytes.NewBuffer(buff)
	_, err := ReadFilenamesResponse(reader)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if reader.Len() != 1 {
		t.Fatal(reader.Len(), "bytes consumed")
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
	reader := bytes.NewReader(buff)
	_, err := ReadRefusal(reader)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if reader.Len() != 1 {
		t.Fatal(reader.Len(), "bytes consumed")
	}
}

func TestReadChunkResponseOfValidResponses(t *testing.T) {
	dataSets := []string{
		"example",
		"asdasdads",
		"chunk",
	}
	for i, chunk := range dataSets {
		t.Run(fmt.Sprint("dataset ", i), func(t *testing.T) {
			buff := make([]byte, 4, 4+len(chunk))
			binary.BigEndian.PutUint32(buff, uint32(len(chunk)))
			buff = append(buff, chunk...)
			writer := bytes.NewBuffer(make([]byte, 0, len(chunk)))
			result, err := ReadChunkResponse(bytes.NewReader(buff), writer)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if result != uint32(len(chunk)) {
				t.Error(result, "bytes received, expected", len(chunk))
			}
			received := string(writer.Bytes())
			if received != chunk {
				t.Error("received", received, ", expected", chunk)
			}
		})
	}
}

func TestReadChunkResponseOfInvalidResponses(t *testing.T) {
	dataSets := []string{
		"example",
		"asdasdads",
		"chunk",
	}
	for i, chunk := range dataSets {
		t.Run(fmt.Sprint("dataset ", i), func(t *testing.T) {
			buff := make([]byte, 4, 4+len(chunk))
			binary.BigEndian.PutUint32(buff, uint32(len(chunk)))
			buff = append(buff, chunk...)
			writer := bytes.NewBuffer(make([]byte, 0, len(chunk)))
			_, err := ReadChunkResponse(bytes.NewReader(buff[:len(buff)-1]), writer)
			if err == nil {
				t.Fatal("expected error not returned")
			}
		})
	}
}

func TestWriteFilenamesRequest(t *testing.T) {
	buffer := bytes.NewBuffer(make([]byte, 0, 2))
	err := WriteFilenamesRequest(buffer)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	requestType, err := ReadRequestType(buffer)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if requestType != RequestTypeFilenames {
		t.Fatal("read", requestType, ", expected", RequestTypeFilenames)
	}
}

func TestWriteChunkRequest(t *testing.T) {
	dataSets := []struct {
		offset   uint32
		size     uint32
		filename string
	}{
		{0, 0, "asd"},
		{1, 2, "example"},
		{^uint32(0), ^uint32(0), "big_vals"},
		{2, 2, "  whitespaces  "},
	}
	for i, dataSet := range dataSets {
		t.Run(fmt.Sprint("dataset ", i), func(t *testing.T) {
			buffer := bytes.NewBuffer(make([]byte, 0, 12+len(dataSet.filename)))
			err := WriteChunkRequest(buffer, dataSet.offset, dataSet.size, []byte(dataSet.filename))
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			requestType, err := ReadRequestType(buffer)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if requestType != RequestTypeChunk {
				t.Error("read request type", requestType, ", expected", RequestTypeChunk)
			}
			request, err := ReadChunkRequest(buffer)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if request.Offset != dataSet.offset {
				t.Error("read offset", request.Offset, ", expected", dataSet.offset)
			}
			if request.Size != dataSet.size {
				t.Error("read size", request.Size, ", expected", dataSet.size)
			}
			if string(request.Filename) != dataSet.filename {
				t.Error("read filename", string(request.Filename), ", expected", dataSet.filename)
			}
		})
	}
}

func TestWriteFilenamesResponse(t *testing.T) {
	dataSets := [][]string{
		{},
		{"filename"},
		{"more", "than", "one"},
	}
	for i, dataSet := range dataSets {
		t.Run(fmt.Sprint("dataset ", i), func(t *testing.T) {
			filenames := make([][]byte, 0, len(dataSet))
			length := 0
			for _, name := range dataSet {
				filenames = append(filenames, []byte(name))
				length += len(name)
			}
			buffer := bytes.NewBuffer(make([]byte, 0, 6+length+len(dataSet)))
			err := WriteFilenamesResponse(buffer, filenames)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			responseType, err := ReadResponseType(buffer)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if responseType != ResponseTypeFilenames {
				t.Error("read response type", responseType, ", expected", ResponseTypeFilenames)
			}
			response, err := ReadFilenamesResponse(buffer)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if len(response.Filenames) != len(dataSet) {
				t.Fatal("read", len(response.Filenames), "filenames, expected", len(dataSet))
			}
			for i := 0; i < len(dataSet); i++ {
				if string(response.Filenames[i]) != dataSet[i] {
					t.Error("filenames don't match:", response.Filenames[i], dataSet[i])
				}
			}
		})
	}
}

func TestWriteRefusal(t *testing.T) {
	causes := []uint32{RefusalCauseBadFilename, RefusalCauseBadOffset, RefusalCauseBadSize}
	for _, cause := range causes {
		t.Run(fmt.Sprint("writing cause ", cause), func(t *testing.T) {
			buffer := bytes.NewBuffer(make([]byte, 0, 6))
			err := WriteRefusal(buffer, cause)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			responseType, err := ReadResponseType(buffer)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if responseType != ResponseTypeRefusal {
				t.Error("read response type", responseType, ", expected", ResponseTypeRefusal)
			}
			refusalCause, err := ReadRefusal(buffer)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if refusalCause != cause {
				t.Error("read refusal cause", refusalCause, ", expected", cause)
			}
		})
	}
}

func TestWriteChunkResponseFull(t *testing.T) {
	dataSets := []string{
		"chunk",
		"   whitespaces   ",
		"a",
	}
	for i, chunk := range dataSets {
		t.Run(fmt.Sprint("dataset ", i), func(t *testing.T) {
			reader := bytes.NewBuffer([]byte(chunk))
			writer := bytes.NewBuffer(make([]byte, 0, len(chunk)))
			err := WriteChunkResponse(writer, reader, uint32(len(chunk)))
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			received, err := ReadChunkResponse(writer, reader)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			if received != uint32(len(chunk)) {
				t.Error("received", received, "bytes, expected", len(chunk))
			}
			receivedChunk := string(reader.Bytes())
			if receivedChunk != chunk {
				t.Fatal("received", receivedChunk, ", expected", chunk)
			}
		})
	}
}

func TestWriteChunkResponsePartial(t *testing.T) {
	chunk := "file chunk"
	reader := bytes.NewReader([]byte(chunk))
	writer := bytes.NewBuffer(make([]byte, 0, len(chunk)))
	err := WriteChunkResponse(writer, reader, uint32(len(chunk))-1)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	clientWriter := bytes.NewBuffer(make([]byte, 0, len(chunk)))
	copied, err := ReadChunkResponse(writer, clientWriter)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if copied != uint32(len(chunk))-1 {
		t.Error("received", copied, "bytes, expected", len(chunk)-1)
	}
	if string(clientWriter.Bytes()) != chunk[:len(chunk)-1] {
		t.Error("received", string(clientWriter.Bytes()), ", expected", chunk[:len(chunk)-1])
	}
}
