package util

import (
	"bytes"
	"compress/zlib"
	"errors"
	"io"
)

var (
	ErrorCompressedSizeLarger = errors.New("uncompressed data is smaller")
)

//TryCompress will try to zlib compress the data given, if any error occurs, data is returned unchanged. If the compressed data is larger than the original, the original data and ErrorCompressedSizeLarger is returned
func TryCompress(data []byte) ([]byte, error) {
	var err error

	var b bytes.Buffer
	zWriter := zlib.NewWriter(&b)

	if _, err = zWriter.Write(data); err != nil {
		return data, err
	}

	if err = zWriter.Close(); err != nil {
		return data, err
	}

	compressedBytes := b.Bytes()

	if len(data) <= len(compressedBytes) {
		return data, ErrorCompressedSizeLarger
	} else {
		return compressedBytes, nil
	}
}

//TryDecompress tries to decompress some zlib compressed data with a declared size, will return nil and an error on failure
func TryDecompress(data []byte, ucSize int64) ([]byte, error) {
	var err error

	var outBuffer bytes.Buffer
	dataBuffer := bytes.NewBuffer(data)

	var zReader io.ReadCloser
	if zReader, err = zlib.NewReader(dataBuffer); err != nil {
		return nil, err
	}

	if readUCBytes, err := io.Copy(&outBuffer, io.LimitReader(zReader, ucSize+1)); err != nil {
		return nil, err
	} else if readUCBytes > ucSize {
		return nil, errors.New("uncompressed size is not as declared")
	}

	if err = zReader.Close(); err != nil {
		return nil, err
	}

	return outBuffer.Bytes(), nil
}
