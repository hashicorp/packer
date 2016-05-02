package packfile

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"gopkg.in/src-d/go-git.v3/core"
	"gopkg.in/src-d/go-git.v3/storage/memory"

	"github.com/dustin/go-humanize"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type ReaderSuite struct{}

var _ = Suite(&ReaderSuite{})

var packFileWithEmptyObjects = "UEFDSwAAAAIAAAALnw54nKXMQWoDMQxA0b1PoX2hSLIm44FSAlmXnEG2NYlhXAfHgdLb5Cy9WAM5Qpb/Lf7oZqArUpakyYtQjCoxZ5lmWXwwyuzJbHqAuYt2+x6QoyCyhYCKIa67lGameSLWvPh5JU0hsCg7vY1z6/D1d/8ptcHhprm3Kxz7KL/wUdOz96eqZXtPrX4CCeOOPU8Eb0iI7qG1jGGvXdxaNoPs/gHeNkp8lA94nKXMQUpDMRCA4X1OMXtBZpI3L3kiRXAtPcMkmWjgxZSYQultPEsv1oJHcPl/i38OVRC0IXF0lshrJorZEcpKmTEJYbA+B3aFzEmGfk9gpqJEsmnZNutXF71i1IURU/G0bsWWwJ6NnOdXH/Bx+73U1uH9LHn0HziOWa/w2tJfv302qftz6u0AtFh0wQdmeEJCNA9tdU7938WUuivEF5CczR11ZEsNnw54nKWMUQoCIRRF/13F+w/ijY6jQkTQd7SGpz5LyAxzINpNa2ljTbSEPu/hnNsbM4TJTzqyt561GdUUmJKT6K2MeiCVgnZWoY/iRo2vHVS0URrUS+e+dkqIEp11HMhh9IaUkRM6QXM/1waH9+uRS4X9TLHVOxxbz0/YlPDbu1OhfFmHWrYwjBKVNVaNsMIBUSy05N75vxeR8oXBiw8GoErCnwt4nKXMzQkCMRBA4XuqmLsgM2M2ZkAWwbNYQ341sCEQsyB2Yy02pmAJHt93eKOnBFpMNJqtl5CFxVIMomViomQSEWP2JrN3yq3j1jqc369HqQ1Oq4u93eHSR3nCoYZfH6/VlWUbWp2BNOPO7i1OsEFCVF+tZYz030XlsiRw6gPZ0jxaqwV4nDM0MDAzMVFIZHg299HsTRevOXt3a64rj7px6ElP8ERDiGQSQ2uoXe8RrcodS5on+J4/u8HjD4NDKFQyRS8tPx+rbgDt3yiEMHicAwAAAAABPnicS0wEAa4kMOACACTjBKdkZXici7aaYAUAA3gBYKoDeJwzNDAwMzFRSGR4NvfR7E0Xrzl7d2uuK4+6cehJT/BEQ4hkEsOELYFJvS2eX47UJdVttFQrenrmzQwA13MaiDd4nEtMBAEuAApMAlGtAXicMzQwMDMxUUhkeDb30exNF685e3drriuPunHoSU/wRACvkA258N/i8hVXx9CiAZzvFXNIhCuSFmE="

func (s *ReaderSuite) TestReadPackfile(c *C) {
	data, _ := base64.StdEncoding.DecodeString(packFileWithEmptyObjects)
	d := bytes.NewReader(data)

	r := NewReader(d)

	storage := memory.NewObjectStorage()
	_, err := r.Read(storage)
	c.Assert(err, IsNil)

	AssertObjects(c, storage, []string{
		"778c85ff95b5514fea0ba4c7b6a029d32e2c3b96",
		"db4002e880a08bf6cc7217512ad937f1ac8824a2",
		"551fe11a9ef992763b7e0be4500cf7169f2f8575",
		"3d8d2705c6b936ceff0020989eca90db7a372609",
		"af01d4cac3441bba4bdd4574938e1d231ee5d45e",
		"e69de29bb2d1d6434b8b29ae775ad8c2e48c5391",
		"85553e8dc42a79b8a483904dcfcdb048fc004055",
		"a028c5b32117ed11bd310a61d50ca10827d853f1",
		"c6b65deb8be57436ceaf920b82d51a3fc59830bd",
		"90b451628d8449f4c47e627eb1392672e5ccec98",
		"496d6428b9cf92981dc9495211e6e1120fb6f2ba",
	})
}

func (s *ReaderSuite) TestReadPackfileOFSDelta(c *C) {
	s.testReadPackfileGitFixture(c, "fixtures/git-fixture.ofs-delta", OFSDeltaFormat)

}
func (s *ReaderSuite) TestReadPackfileREFDelta(c *C) {
	s.testReadPackfileGitFixture(c, "fixtures/git-fixture.ref-delta", REFDeltaFormat)
}

func (s *ReaderSuite) testReadPackfileGitFixture(c *C, file string, f Format) {
	d, err := os.Open(file)
	c.Assert(err, IsNil)

	r := NewReader(d)
	r.Format = f

	storage := memory.NewObjectStorage()
	_, err = r.Read(storage)
	c.Assert(err, IsNil)

	AssertObjects(c, storage, []string{
		"918c48b83bd081e863dbe1b80f8998f058cd8294",
		"af2d6a6954d532f8ffb47615169c8fdf9d383a1a",
		"1669dce138d9b841a518c64b10914d88f5e488ea",
		"a5b8b09e2f8fcb0bb99d3ccb0958157b40890d69",
		"b8e471f58bcbca63b07bda20e428190409c2db47",
		"35e85108805c84807bc66a02d91535e1e24b38b9",
		"b029517f6300c2da0f4b651b8642506cd6aaf45d",
		"32858aad3c383ed1ff0a0f9bdf231d54a00c9e88",
		"d3ff53e0564a9f87d8e84b6e28e5060e517008aa",
		"c192bd6a24ea1ab01d78686e417c8bdc7c3d197f",
		"d5c0f4ab811897cadf03aec358ae60d21f91c50d",
		"49c6bb89b17060d7b4deacb7b338fcc6ea2352a9",
		"cf4aa3b38974fb7d81f367c0830f7d78d65ab86b",
		"9dea2395f5403188298c1dabe8bdafe562c491e3",
		"586af567d0bb5e771e49bdd9434f5e0fb76d25fa",
		"9a48f23120e880dfbe41f7c9b7b708e9ee62a492",
		"5a877e6a906a2743ad6e45d99c1793642aaf8eda",
		"c8f1d8c61f9da76f4cb49fd86322b6e685dba956",
		"a8d315b2b1c615d43042c3a62402b8a54288cf5c",
		"a39771a7651f97faf5c72e08224d857fc35133db",
		"880cd14280f4b9b6ed3986d6671f907d7cc2a198",
		"fb72698cab7617ac416264415f13224dfd7a165e",
		"4d081c50e250fa32ea8b1313cf8bb7c2ad7627fd",
		"eba74343e2f15d62adedfd8c883ee0262b5c8021",
		"c2d30fa8ef288618f65f6eed6e168e0d514886f4",
		"8dcef98b1d52143e1e2dbc458ffe38f925786bf2",
		"aa9b383c260e1d05fbbf6b30a02914555e20c725",
		"6ecf0ef2c2dffb796033e5a02219af86ec6584e5",
	})
}

func AssertObjects(c *C, s *memory.ObjectStorage, expects []string) {
	c.Assert(len(expects), Equals, len(s.Objects))
	for _, expected := range expects {
		obtained, err := s.Get(core.NewHash(expected))
		c.Assert(err, IsNil)
		c.Assert(obtained.Hash().String(), Equals, expected)
	}
}

func (s *ReaderSuite) BenchmarkFixtureRef(c *C) {
	for i := 0; i < c.N; i++ {
		readFromFile(c, "fixtures/git-fixture.ref-delta", REFDeltaFormat)
	}
}

func (s *ReaderSuite) BenchmarkFixtureOfs(c *C) {
	for i := 0; i < c.N; i++ {
		readFromFile(c, "fixtures/git-fixture.ofs-delta", OFSDeltaFormat)
	}
}

func (s *ReaderSuite) BenchmarkCandyJS(c *C) {
	for i := 0; i < c.N; i++ {
		readFromFile(c, "/tmp/go-candyjs", REFDeltaFormat)
	}
}

func (s *ReaderSuite) BenchmarkSymfony(c *C) {
	for i := 0; i < c.N; i++ {
		readFromFile(c, "/tmp/symonfy", REFDeltaFormat)
	}
}

func (s *ReaderSuite) BenchmarkGit(c *C) {
	for i := 0; i < c.N; i++ {
		readFromFile(c, "/tmp/git", REFDeltaFormat)
	}
}

func (s *ReaderSuite) _TestMemoryOFS(c *C) {
	var b, a runtime.MemStats

	start := time.Now()
	runtime.ReadMemStats(&b)
	p := readFromFile(c, "/tmp/symfony.ofs-delta", OFSDeltaFormat)
	runtime.ReadMemStats(&a)

	fmt.Println("OFS--->")
	fmt.Println("Alloc", a.Alloc-b.Alloc, humanize.Bytes(a.Alloc-b.Alloc))
	fmt.Println("TotalAlloc", a.TotalAlloc-b.TotalAlloc, humanize.Bytes(a.TotalAlloc-b.TotalAlloc))
	fmt.Println("HeapAlloc", a.HeapAlloc-b.HeapAlloc, humanize.Bytes(a.HeapAlloc-b.HeapAlloc))
	fmt.Println("HeapSys", a.HeapSys, humanize.Bytes(a.HeapSys-b.HeapSys))

	fmt.Println("objects", len(p.Objects))
	fmt.Println("time", time.Since(start))
}

func (s *ReaderSuite) _TestMemoryREF(c *C) {
	var b, a runtime.MemStats

	start := time.Now()
	runtime.ReadMemStats(&b)
	p := readFromFile(c, "/tmp/symonfy", REFDeltaFormat)
	runtime.ReadMemStats(&a)

	fmt.Println("REF--->")
	fmt.Println("Alloc", a.Alloc-b.Alloc, humanize.Bytes(a.Alloc-b.Alloc))
	fmt.Println("TotalAlloc", a.TotalAlloc-b.TotalAlloc, humanize.Bytes(a.TotalAlloc-b.TotalAlloc))
	fmt.Println("HeapAlloc", a.HeapAlloc-b.HeapAlloc, humanize.Bytes(a.HeapAlloc-b.HeapAlloc))
	fmt.Println("HeapSys", a.HeapSys, humanize.Bytes(a.HeapSys-b.HeapSys))

	fmt.Println("objects", len(p.Objects))
	fmt.Println("time", time.Since(start))
}

func readFromFile(c *C, file string, f Format) *memory.ObjectStorage {
	d, err := os.Open(file)
	c.Assert(err, IsNil)

	r := NewReader(d)
	r.Format = f

	storage := memory.NewObjectStorage()
	_, err = r.Read(storage)
	c.Assert(err, IsNil)

	return storage
}
