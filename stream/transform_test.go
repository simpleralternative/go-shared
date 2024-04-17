package stream

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// example data structure
type Row struct {
	Id    int
	Name  string
	Phone string
}

func TestTransform(t *testing.T) {
	data := []Row{
		{1, "J", "2125551234"},
		{1, "Jonah", "2125551234"},
		{1, "Jameson", "2125551234"},
	}
	filename := "testfiles/example_good.gz"

	t.Run("reference file write process", func(t *testing.T) {
		t.Skip()

		b, err := json.Marshal(data)
		if err != nil {
			t.Error(err)
		}

		var buf bytes.Buffer
		writer := gzip.NewWriter(&buf)
		defer writer.Close()
		_, err = writer.Write(b)
		if err != nil {
			t.Error(err)
		}

		file, err := os.Create(filename)
		if err != nil {
			t.Error(err)
		}
		defer file.Close()

		bytesWritten, err := file.Write(buf.Bytes())

		require.NoError(t, err)
		require.Greater(t, bytesWritten, 0)
	})

	t.Run("complex file write process", func(t *testing.T) {
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
				defer writer.Close()
				_, err := writer.Write(res.Value)
				return NewResult(buf.Bytes(), err)
			})

		result := (<-Transform(compressed,
			func(ctx context.Context, res *Result[[]byte]) *Result[int] {
				// this could technically be done by returning a Result with a
				// struct of a pair with the file reference and the data, but
				// there's value in balance.
				file, err := os.Create(filename)
				if err != nil {
					return NewResult(0, err)
				}
				defer file.Close()

				return NewResult(file.Write(res.Value))
			}))

		require.NoError(t, result.Error)
		require.Greater(t, result.Value, 0)
	})

	t.Run("complex file read process", func(t *testing.T) {
		file := Stream(func(_ context.Context, output chan<- *Result[*os.File]) {
			output <- NewResult(os.Open(filename))
		})

		gzipped := Transform(file,
			func(_ context.Context, res *Result[*os.File]) *Result[[]byte] {
				return NewResult(io.ReadAll(res.Value))
			})

		gzipReader := Transform(gzipped,
			func(ctx context.Context, res *Result[[]byte]) *Result[*gzip.Reader] {
				return NewResult(gzip.NewReader(bytes.NewReader(res.Value)))
			})

		unzipped := Transform(gzipReader,
			func(ctx context.Context, res *Result[*gzip.Reader]) *Result[[]byte] {
				return NewResult(io.ReadAll(res.Value))
			})

		rows := Transform(unzipped,
			func(ctx context.Context, res *Result[[]byte]) *Result[[]Row] {
				var data []Row
				err := json.Unmarshal(res.Value, &data)
				return NewResult(data, err)
			})

		require.Equal(t, data, rows)
	})

}
