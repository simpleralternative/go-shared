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

// example library of transforms
func jsonMarshaller(_ context.Context, res *Result[[]Row]) *Result[[]byte] {
	return NewResult(json.Marshal(res.Value))
}

func jsonUnmarshaller[T any](_ context.Context, res *Result[[]byte]) *Result[T] {
	data := *new(T)
	err := json.Unmarshal(res.Value, &data)
	return NewResult(data, err)
}

func gzipEncoder(_ context.Context, res *Result[[]byte]) *Result[[]byte] {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write(res.Value)
	writer.Close()
	return NewResult(buf.Bytes(), err)
}

func gzipDecoder(_ context.Context, res *Result[[]byte]) *Result[io.ReadCloser] {
	return NewResult[io.ReadCloser](gzip.NewReader(bytes.NewReader(res.Value)))
}

func fileReplacer(
	filename string,
) func(context.Context, *Result[[]byte]) *Result[int] {
	return func(_ context.Context, res *Result[[]byte]) *Result[int] {
		if err := os.MkdirAll(path.Dir(filename), 0777); err != nil {
			return NewResult(0, err)
		}

		file, err := os.Create(filename)
		if err != nil {
			return NewResult(0, err)
		}
		defer file.Close()

		return NewResult(file.Write(res.Value))
	}
}

func readAndClose(_ context.Context, res *Result[io.ReadCloser]) *Result[[]byte] {
	defer res.Value.Close()
	return NewResult(io.ReadAll(res.Value))
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
			func(_ context.Context, res *Result[[]Row]) *Result[[]byte] {
				return NewResult(json.Marshal(res.Value))
			})

		compressed := Transform(marshalled,
			func(ctx context.Context, res *Result[[]byte]) *Result[[]byte] {
				var buf bytes.Buffer
				writer := gzip.NewWriter(&buf)
				_, err := writer.Write(res.Value)
				writer.Close()
				return NewResult(buf.Bytes(), err)
			})

		result := (<-Transform(compressed,
			func(ctx context.Context, res *Result[[]byte]) *Result[int] {
				// this could technically be done by returning a Result with a
				// struct of a pair with the file reference and the data, but
				// there's value in balance.
				if err := os.MkdirAll(path.Dir(goodfile), 0777); err != nil {
					return NewResult(0, err)
				}

				file, err := os.Create(goodfile)
				if err != nil {
					return NewResult(0, err)
				}
				defer file.Close()

				return NewResult(file.Write(res.Value))
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
