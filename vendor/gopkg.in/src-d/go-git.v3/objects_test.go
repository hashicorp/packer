package git

import (
	"io/ioutil"
	"time"

	. "gopkg.in/check.v1"
	"gopkg.in/src-d/go-git.v3/core"
	"gopkg.in/src-d/go-git.v3/storage/memory"
)

type ObjectsSuite struct {
	r *Repository
}

var _ = Suite(&ObjectsSuite{})

func (s *ObjectsSuite) SetUpTest(c *C) {
	var err error
	s.r, err = NewRepository(RepositoryFixture, nil)
	s.r.Remotes["origin"].upSrv = &MockGitUploadPackService{}

	s.r.Pull("origin", "refs/heads/master")
	c.Assert(err, IsNil)
}

func (s *ObjectsSuite) TestNewCommit(c *C) {
	hash := core.NewHash("a5b8b09e2f8fcb0bb99d3ccb0958157b40890d69")
	commit, err := s.r.Commit(hash)
	c.Assert(err, IsNil)

	c.Assert(commit.Hash, Equals, commit.ID())
	c.Assert(commit.Hash.String(), Equals, "a5b8b09e2f8fcb0bb99d3ccb0958157b40890d69")
	c.Assert(commit.Tree().Hash.String(), Equals, "c2d30fa8ef288618f65f6eed6e168e0d514886f4")

	parents := commit.Parents()
	parentCommit, err := parents.Next()
	c.Assert(err, IsNil)
	c.Assert(parentCommit.Hash.String(), Equals, "b029517f6300c2da0f4b651b8642506cd6aaf45d")

	parentCommit, err = parents.Next()
	c.Assert(err, IsNil)
	c.Assert(parentCommit.Hash.String(), Equals, "b8e471f58bcbca63b07bda20e428190409c2db47")

	c.Assert(commit.Author.Email, Equals, "mcuadros@gmail.com")
	c.Assert(commit.Author.Name, Equals, "Máximo Cuadros")
	c.Assert(commit.Author.When.Format(time.RFC3339), Equals, "2015-03-31T13:47:14+02:00")
	c.Assert(commit.Committer.Email, Equals, "mcuadros@gmail.com")
	c.Assert(commit.Message, Equals, "Merge pull request #1 from dripolles/feature\n\nCreating changelog\n")
}

func (s *ObjectsSuite) TestParseTree(c *C) {
	hash := core.NewHash("a8d315b2b1c615d43042c3a62402b8a54288cf5c")
	tree, err := s.r.Tree(hash)
	c.Assert(err, IsNil)

	c.Assert(tree.Entries, HasLen, 8)

	tree.buildMap()
	c.Assert(tree.m, HasLen, 8)
	c.Assert(tree.m[".gitignore"].Name, Equals, ".gitignore")
	c.Assert(tree.m[".gitignore"].Mode.String(), Equals, "-rw-r--r--")
	c.Assert(tree.m[".gitignore"].Hash.String(), Equals, "32858aad3c383ed1ff0a0f9bdf231d54a00c9e88")

	count := 0
	iter := tree.Files()
	defer iter.Close()
	for f, err := iter.Next(); err == nil; f, err = iter.Next() {
		count++
		if f.Name == "go/example.go" {
			reader, err := f.Reader()
			c.Assert(err, IsNil)
			defer func() { c.Assert(reader.Close(), IsNil) }()
			content, _ := ioutil.ReadAll(reader)
			c.Assert(content, HasLen, 2780)
		}
	}

	c.Assert(count, Equals, 9)
}

func (s *ObjectsSuite) TestBlobHash(c *C) {
	o := &memory.Object{}
	o.SetType(core.BlobObject)
	o.SetSize(3)

	writer, err := o.Writer()
	c.Assert(err, IsNil)
	defer func() { c.Assert(writer.Close(), IsNil) }()

	writer.Write([]byte{'F', 'O', 'O'})

	blob := &Blob{}
	c.Assert(blob.Decode(o), IsNil)

	c.Assert(blob.Size, Equals, int64(3))
	c.Assert(blob.Hash.String(), Equals, "d96c7efbfec2814ae0301ad054dc8d9fc416c9b5")

	reader, err := blob.Reader()
	c.Assert(err, IsNil)
	defer func() { c.Assert(reader.Close(), IsNil) }()

	data, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	c.Assert(string(data), Equals, "FOO")
}

func (s *ObjectsSuite) TestParseSignature(c *C) {
	cases := map[string]Signature{
		`Foo Bar <foo@bar.com> 1257894000 +0100`: {
			Name:  "Foo Bar",
			Email: "foo@bar.com",
			When:  MustParseTime("2009-11-11 00:00:00 +0100"),
		},
		`Foo Bar <foo@bar.com> 1257894000 -0700`: {
			Name:  "Foo Bar",
			Email: "foo@bar.com",
			When:  MustParseTime("2009-11-10 16:00:00 -0700"),
		},
		`Foo Bar <> 1257894000 +0100`: {
			Name:  "Foo Bar",
			Email: "",
			When:  MustParseTime("2009-11-11 00:00:00 +0100"),
		},
		` <> 1257894000`: {
			Name:  "",
			Email: "",
			When:  MustParseTime("2009-11-10 23:00:00 +0000"),
		},
		`Foo Bar <foo@bar.com>`: {
			Name:  "Foo Bar",
			Email: "foo@bar.com",
			When:  time.Time{},
		},
		``: {
			Name:  "",
			Email: "",
			When:  time.Time{},
		},
		`<`: {
			Name:  "",
			Email: "",
			When:  time.Time{},
		},
	}

	for raw, exp := range cases {
		got := &Signature{}
		got.Decode([]byte(raw))

		c.Assert(got.Name, Equals, exp.Name)
		c.Assert(got.Email, Equals, exp.Email)
		c.Assert(got.When.Format(time.RFC3339), Equals, exp.When.Format(time.RFC3339))
	}
}

func MustParseTime(value string) time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05 -0700", value)
	return t
}
