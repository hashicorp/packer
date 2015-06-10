// +build ignore

package main

import (
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	"github.com/biogo/hts/bgzf"
	"github.com/klauspost/pgzip"
	"github.com/pierrec/lz4"
)

type Compressor struct {
	r  *os.File
	w  *os.File
	sr int64
	sw int64
}

func (c *Compressor) Close() error {
	var err error

	fi, _ := c.w.Stat()
	c.sw = fi.Size()
	if err = c.w.Close(); err != nil {
		return err
	}

	fi, _ = c.r.Stat()
	c.sr = fi.Size()
	if err = c.r.Close(); err != nil {
		return err
	}

	return nil
}

func NewCompressor(src, dst string) (*Compressor, error) {
	r, err := os.Open(src)
	if err != nil {
		return nil, err
	}

	w, err := os.Create(dst)
	if err != nil {
		r.Close()
		return nil, err
	}

	c := &Compressor{r: r, w: w}
	return c, nil
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	var resw testing.BenchmarkResult
	var resr testing.BenchmarkResult

	c, err := NewCompressor("/tmp/image.r", "/tmp/image.w")
	if err != nil {
		panic(err)
	}
	resw = testing.Benchmark(c.BenchmarkGZIPWriter)
	c.w.Seek(0, 0)
	resr = testing.Benchmark(c.BenchmarkGZIPReader)
	c.Close()
	fmt.Printf("gzip:\twriter %s\treader %s\tsize %d\n", resw.T.String(), resr.T.String(), c.sw)

	c, err = NewCompressor("/tmp/image.r", "/tmp/image.w")
	if err != nil {
		panic(err)
	}
	resw = testing.Benchmark(c.BenchmarkBGZFWriter)
	c.w.Seek(0, 0)
	resr = testing.Benchmark(c.BenchmarkBGZFReader)
	c.Close()
	fmt.Printf("bgzf:\twriter %s\treader %s\tsize %d\n", resw.T.String(), resr.T.String(), c.sw)

	c, err = NewCompressor("/tmp/image.r", "/tmp/image.w")
	if err != nil {
		panic(err)
	}
	resw = testing.Benchmark(c.BenchmarkPGZIPWriter)
	c.w.Seek(0, 0)
	resr = testing.Benchmark(c.BenchmarkPGZIPReader)
	c.Close()
	fmt.Printf("pgzip:\twriter %s\treader %s\tsize %d\n", resw.T.String(), resr.T.String(), c.sw)

	c, err = NewCompressor("/tmp/image.r", "/tmp/image.w")
	if err != nil {
		panic(err)
	}
	resw = testing.Benchmark(c.BenchmarkLZ4Writer)
	c.w.Seek(0, 0)
	resr = testing.Benchmark(c.BenchmarkLZ4Reader)
	c.Close()
	fmt.Printf("lz4:\twriter %s\treader %s\tsize %d\n", resw.T.String(), resr.T.String(), c.sw)

}

func (c *Compressor) BenchmarkGZIPWriter(b *testing.B) {
	cw, _ := gzip.NewWriterLevel(c.w, flate.BestSpeed)
	b.ResetTimer()

	_, err := io.Copy(cw, c.r)
	if err != nil {
		b.Fatal(err)
	}
	cw.Close()
	c.w.Sync()
}

func (c *Compressor) BenchmarkGZIPReader(b *testing.B) {
	cr, _ := gzip.NewReader(c.w)
	b.ResetTimer()

	_, err := io.Copy(ioutil.Discard, cr)
	if err != nil {
		b.Fatal(err)
	}
}

func (c *Compressor) BenchmarkBGZFWriter(b *testing.B) {
	cw, _ := bgzf.NewWriterLevel(c.w, flate.BestSpeed, runtime.NumCPU())
	b.ResetTimer()

	_, err := io.Copy(cw, c.r)
	if err != nil {
		b.Fatal(err)
	}
	c.w.Sync()
}

func (c *Compressor) BenchmarkBGZFReader(b *testing.B) {
	cr, _ := bgzf.NewReader(c.w, 0)
	b.ResetTimer()

	_, err := io.Copy(ioutil.Discard, cr)
	if err != nil {
		b.Fatal(err)
	}
}

func (c *Compressor) BenchmarkPGZIPWriter(b *testing.B) {
	cw, _ := pgzip.NewWriterLevel(c.w, flate.BestSpeed)
	b.ResetTimer()

	_, err := io.Copy(cw, c.r)
	if err != nil {
		b.Fatal(err)
	}
	cw.Close()
	c.w.Sync()
}

func (c *Compressor) BenchmarkPGZIPReader(b *testing.B) {
	cr, _ := pgzip.NewReader(c.w)
	b.ResetTimer()

	_, err := io.Copy(ioutil.Discard, cr)
	if err != nil {
		b.Fatal(err)
	}
}

func (c *Compressor) BenchmarkLZ4Writer(b *testing.B) {
	cw := lz4.NewWriter(c.w)
	//	cw.Header.HighCompression = true
	cw.Header.NoChecksum = true
	b.ResetTimer()

	_, err := io.Copy(cw, c.r)
	if err != nil {
		b.Fatal(err)
	}
	cw.Close()
	c.w.Sync()
}

func (c *Compressor) BenchmarkLZ4Reader(b *testing.B) {
	cr := lz4.NewReader(c.w)
	b.ResetTimer()

	_, err := io.Copy(ioutil.Discard, cr)
	if err != nil {
		b.Fatal(err)
	}
}
