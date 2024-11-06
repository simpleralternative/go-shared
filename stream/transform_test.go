package stream

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

// example data structure
type Row struct {
	Id    int
	Name  string
	Phone string
}

// an example library of simple transforms that defer to the library to handle
// most of the traditional housekeeping.

func jsonMarshaller(_ context.Context, input []Row) ([]byte, error) {
	return Trace(json.Marshal(input))
}

func jsonUnmarshaller[T any](_ context.Context, input []byte) (T, error) {
	var data T
	err := json.Unmarshal(input, &data)
	return Trace(data, err)
}

func gzipEncoder(_ context.Context, input []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, writeErr := writer.Write(input)
	// note: syntactically, you could inline the writer.Close in the
	// errors.Join, but it will be executed /after/ the buffer.Bytes call,
	// likely returning incomplete data.
	closeErr := writer.Close()
	return Trace(buf.Bytes(), errors.Join(writeErr, closeErr))
}

func gzipDecoder(_ context.Context, input []byte) (io.ReadCloser, error) {
	return Trace(gzip.NewReader(bytes.NewReader(input)))
}

// fileReplacer demonstrates a library function pattern with injectable
// configuration.
func fileReplacer(
	filename string,
) func(_ context.Context, input []byte) (int, error) {
	// return conforming lambda with filename closure
	return func(_ context.Context, input []byte) (int, error) {
		if err := os.MkdirAll(path.Dir(filename), 0777); err != nil {
			return Trace(0, err)
		}

		file, err := os.Create(filename)
		if err != nil {
			return Trace(0, err)
		}
		defer file.Close()

		return Trace(file.Write(input))
	}
}

func readAndClose(_ context.Context, input io.ReadCloser) ([]byte, error) {
	defer input.Close()
	return Trace(io.ReadAll(input))
}

// end library

func TestTransform(t *testing.T) {
	data := []Row{
		{1, "J", "2125551234"},
		{1, "Jonah", "2125551235"},
		{1, "Jameson", "2125551236"},
	}
	goodfile := "testfiles/example_good.gz"
	nofile := "testfiles/example_nonexistant.gz"
	badfile := "testfiles/example_bad.gz"

	t.Run("reference file write process", func(t *testing.T) {
		b, err := json.Marshal(data)
		if err != nil {
			t.Error(err)
		}

		var buf bytes.Buffer
		func() {
			writer := gzip.NewWriter(&buf)
			defer writer.Close()

			_, err = writer.Write(b)
			if err != nil {
				t.Error(err)
			}
		}()

		if err := os.MkdirAll(path.Dir(goodfile), 0777); err != nil {
			t.Error(err)
		}

		bytesWritten, err := func() (int, error) {
			file, err := os.Create(goodfile)
			if err != nil {
				t.Error(err)
			}
			defer file.Close()

			return file.Write(buf.Bytes())
		}()

		require.NoError(t, err)
		require.Equal(t, 88, bytesWritten)
	})

	t.Run("complex file write process with anonymous funcs", func(t *testing.T) {
		rows := Stream(func(_ context.Context, output chan<- *Result[[]Row]) {
			output <- NewResult(data, nil)
		})

		marshalled := Transform(rows,
			func(_ context.Context, input []Row) ([]byte, error) {
				return json.Marshal(input)
			})

		compressed := Transform(marshalled,
			func(ctx context.Context, input []byte) ([]byte, error) {
				var buf bytes.Buffer
				writer := gzip.NewWriter(&buf)
				_, err := writer.Write(input)
				writer.Close()
				return buf.Bytes(), err
			})

		result := (<-Transform(compressed,
			func(ctx context.Context, input []byte) (int, error) {
				// this could technically be done by returning a Result with a
				// struct of a pair with the file reference and the data, but
				// there's value in balance.
				if err := os.MkdirAll(path.Dir(goodfile), 0777); err != nil {
					return 0, err
				}

				file, err := os.Create(goodfile)
				if err != nil {
					return 0, err
				}
				defer file.Close()

				return file.Write(input)
			}))

		require.NoError(t, result.Error)
		require.Equal(t, 88, result.Value)
	})

	t.Run("complex file write process from lib", func(t *testing.T) {
		rows := Stream(func(_ context.Context, output chan<- *Result[[]Row]) {
			output <- NewResult(data, nil)
		})

		marshalled := Transform(rows, jsonMarshaller)

		compressed := Transform(marshalled, gzipEncoder)

		result := (<-Transform(compressed, fileReplacer(goodfile)))

		require.NoError(t, result.Error)
		require.Equal(t, 88, result.Value)
	})

	t.Run("reference file read process", func(t *testing.T) {
		file, err := os.Open(goodfile)
		if err != nil {
			t.Error(err)
		}
		defer file.Close()

		gzipped, err := io.ReadAll(file)
		if err != nil {
			t.Error(err)
		}

		gzipReader, err := gzip.NewReader(bytes.NewReader(gzipped))
		if err != nil {
			t.Error(err)
		}
		defer gzipReader.Close()

		unzipped, err := io.ReadAll(gzipReader)
		if err != nil {
			t.Error(err)
		}

		var rows []Row
		err = json.Unmarshal(unzipped, &rows)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("complex file read process from lib", func(t *testing.T) {
		// note: usage of io.ReadCloser interface rather than *os.File
		file := Stream(func(_ context.Context, output chan<- *Result[io.ReadCloser]) {
			output <- NewResult[io.ReadCloser](os.Open(goodfile))
		})

		gzipped := Transform(file, readAndClose)

		gzipReader := Transform(gzipped, gzipDecoder)

		unzipped := Transform(gzipReader, readAndClose)

		rows := (<-Transform(unzipped, jsonUnmarshaller[[]Row]))

		require.NoError(t, rows.Error)
		require.Equal(t, data, rows.Value)
	})

	t.Run("complex file read failure", func(t *testing.T) {
		file := Stream(func(_ context.Context, output chan<- *Result[io.ReadCloser]) {
			output <- NewResult[io.ReadCloser](os.Open(nofile))
		})

		gzipped := Transform(file, readAndClose)

		gzipReader := Transform(gzipped, gzipDecoder)

		unzipped := Transform(gzipReader, readAndClose)

		rows := (<-Transform(unzipped, jsonUnmarshaller[[]Row]))

		require.True(t, errors.Is(rows.Error, os.ErrNotExist))
	})

	t.Run("complex file unzip failure", func(t *testing.T) {
		file := Stream(func(_ context.Context, output chan<- *Result[io.ReadCloser]) {
			output <- NewResult[io.ReadCloser](os.Open(badfile))
		})

		gzipped := Transform(file, readAndClose)

		gzipReader := Transform(gzipped, gzipDecoder)

		unzipped := Transform(gzipReader, readAndClose)

		rows := (<-Transform(unzipped, jsonUnmarshaller[[]Row]))

		require.True(t, errors.Is(rows.Error, gzip.ErrHeader))
	})
}
