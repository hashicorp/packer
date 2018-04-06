package git

import (
	"io"
	"os"
	"strconv"

	"gopkg.in/src-d/go-git.v3/core"

	. "gopkg.in/check.v1"
)

type expectedTreeWalkerEntry struct {
	Kind core.ObjectType
	Mode string
	Name string
	Hash string
}

var treeWalkerFixtures = []packedFixture{
	{"https://github.com/alcortesm/binary-relations.git", "formats/packfile/fixtures/alcortesm-binary-relations.pack"},
	{"https://github.com/Tribler/dispersy.git", "formats/packfile/fixtures/tribler-dispersy.pack"},
}

var treeWalkerTests = []struct {
	repo   string                    // the repo name as in localRepos
	commit string                    // the commit to search for the file
	objs   []expectedTreeWalkerEntry // the expected objects in the commit
}{
	// https://api.github.com/repos/alcortesm/binary-relations/git/trees/b373f85fa2594d7dcd9989f4a5858a81647fb8ea
	{"https://github.com/alcortesm/binary-relations.git", "b373f85fa2594d7dcd9989f4a5858a81647fb8ea", []expectedTreeWalkerEntry{
		{core.BlobObject, "100644", ".gitignore", "7f41905b4d77ab4a9a2d334fcd0fb5db6e8e2183"},
		{core.BlobObject, "100644", "Makefile", "d441e4e769b53cbd4b1215a1387f8c3108bac97d"},
		{core.BlobObject, "100644", "binary-relations.tex", "cb50b067cc8cd9f639611d41416575c991ad8e97"},
		{core.TreeObject, "040000", "imgs-gen", "b33007b7e83a738576c3f44369fe2f674bb23d5d"},
		{core.TreeObject, "040000", "imgs-gen/simple-graph", "056633542b8ee990d6c89b7a812209dba13d6766"},
		{core.BlobObject, "100644", "imgs-gen/simple-graph/Makefile", "49560402c1707f29c159ad14f369027250fb154a"},
		{core.BlobObject, "100644", "imgs-gen/simple-graph/fig.fig", "2c414eb36f0c2e9a2f9c6382d85e63355752170c"},
		{core.TreeObject, "040000", "src", "ec9d27c4df99caec3a817e9c018812a6c56c1b00"},
		{core.TreeObject, "040000", "src/map-slice", "00cefb8e77f7a8c61b99dd2491ff48a3b0b16679"},
		{core.BlobObject, "100644", "src/map-slice/map-slice.go", "12431e98381dd5097e1a19fe53429c72ef1f328e"},
		{core.TreeObject, "040000", "src/simple-arrays", "9a3781b7fd9d2851e2a4488c035ed9ac905aec79"},
		{core.BlobObject, "100644", "src/simple-arrays/simple-arrays.go", "104fbb4b0520c192f2e207a2dfd39162f6cdabf7"},
	}},
	// https://api.github.com/repos/Tribler/dispersy/git/trees/f5a1fca709f760bf75a7adaa480bf0f0e1a547ee
	{"https://github.com/Tribler/dispersy.git", "f5a1fca709f760bf75a7adaa480bf0f0e1a547ee", []expectedTreeWalkerEntry{
		{core.BlobObject, "100644", "__init__.py", "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391"},
		{core.BlobObject, "100644", "authentication.py", "ca2fb017dce4506c4144ba81d3cbb70563487718"},
		{core.BlobObject, "100644", "bloomfilter.py", "944e8ccc76779aad923f88f4a73b0d4e8999b6ea"},
		{core.BlobObject, "100644", "bootstrap.py", "379a4400b992310f54ea56a4691760bdea8b1592"},
		{core.BlobObject, "100644", "cache.py", "744d48dce50703e7d4ff14531ab2ab77e6b54685"},
		{core.BlobObject, "100644", "callback.py", "f3a380cbe9eb1c02fb305f2bd2fc0fcfb103892f"},
		{core.BlobObject, "100644", "candidate.py", "87309a51d3681bf6c46b22ce044dad41c97d32d2"},
		{core.BlobObject, "100644", "community.py", "38226ffc2139a2349edaf016747d02b199508d41"},
		{core.BlobObject, "100644", "conversion.py", "4e2fcefba40d99c2a6237768ed0fbb8e2e770c83"},
		{core.BlobObject, "100644", "crypto.py", "8a6bb00df982fa806ce18838673ab1ef3fd52fed"},
		{core.BlobObject, "100644", "database.py", "bb484bfd31a92f7775dbd3acf8740abf00bb3d74"},
		{core.BlobObject, "100644", "debug.py", "3743f20d321f7b2b6d3b47211f93317818c3673e"},
		{core.BlobObject, "100644", "debugcommunity.py", "1598ec5a773cc561430c5bb9b87157ef7d3e1c7c"},
		{core.BlobObject, "100644", "decorator.py", "a1e913e674aec5402cc7b4e9fc0801e8155d2cec"},
		{core.BlobObject, "100644", "destination.py", "d5c02588117d260e728d5c64aba885522ba508c5"},
		{core.BlobObject, "100644", "dispersy.py", "63a08602e2ac8294b20543f0c89c75c740bf6c1c"},
		{core.BlobObject, "100644", "dispersydatabase.py", "76dd222444c308c470efabde7ed511825004b4d3"},
		{core.BlobObject, "100644", "distribution.py", "55a11beca7c09013f5b8ff46baa85b15948c756a"},
		{core.BlobObject, "100644", "dprint.py", "fd6a8608d62bf415a65e78c9e1ca8df97513e598"},
		{core.BlobObject, "100644", "encoding.py", "f29b0ebf65f06a0aa7b2ff1aea364f7889c58d56"},
		{core.BlobObject, "100644", "endpoint.py", "5aa76efd3501de522dbdf2e6374440cf64131423"},
		{core.BlobObject, "100644", "member.py", "c114c73f710b4c291305e353b4aa0106fafabd52"},
		{core.BlobObject, "100644", "message.py", "e55bfe0efa851c4e94264dc745141f7f65b1d239"},
		{core.BlobObject, "100644", "meta.py", "0f62db0fb93326daad6b4925a7d12155a1687f67"},
		{core.BlobObject, "100644", "payload.py", "0aef13194c51dab3624665340b33dd4040516c86"},
		{core.BlobObject, "100644", "requestcache.py", "7772c7d81b4b205970cac1a3cdabc2c2deb48b12"},
		{core.BlobObject, "100644", "resolution.py", "525d6ec81c1fb098d2fe12ae0d5b10a368bfcace"},
		{core.BlobObject, "100644", "script.py", "ef64e12cc1a4c0b3a5d42ff1b33adef202f30da3"},
		{core.BlobObject, "100644", "singleton.py", "34662093edf45bbffa91125c13735e37410a185b"},
		{core.BlobObject, "100644", "timeline.py", "826bb5e1802fb5eaf3144a9a195a994920101880"},
		{core.TreeObject, "040000", "tool", "da97281af01b5b2dad1de6c84c5acb44da60ef7a"},
		{core.BlobObject, "100644", "tool/__init__.py", "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391"},
		{core.BlobObject, "100644", "tool/callbackscript.py", "eb9f8184ef08d9e031936e61bfa86fb9b45b965c"},
		{core.BlobObject, "100644", "tool/scenarioscript.py", "245c41af66aab8f0a6fd00259b30a47f4d6c00dd"},
	}},
	{"https://github.com/Tribler/dispersy.git", "9d38ff85ca03adcf68dc14f5b68b8994f15229f4", []expectedTreeWalkerEntry(nil)},
}

type SuiteTreeWalker struct {
	repos map[string]*Repository
}

var _ = Suite(&SuiteTreeWalker{})

// create the repositories of the fixtures
func (s *SuiteTreeWalker) SetUpSuite(c *C) {
	s.repos = unpackFixtures(c, treeWalkerFixtures)
}

func (s *SuiteTreeWalker) TestNext(c *C) {
	for i, t := range treeWalkerTests {
		r := s.repos[t.repo]
		commit, err := r.Commit(core.NewHash(t.commit))
		c.Assert(err, IsNil, Commentf("subtest %d: %v (%s)", i, err, t.commit))

		walker := NewTreeWalker(r, commit.Tree())
		for k := 0; k < len(t.objs); k++ {
			info := t.objs[k]
			mode, err := strconv.ParseInt(info.Mode, 8, 32)
			c.Assert(err, IsNil)
			name, entry, obj, err := walker.Next()

			c.Assert(err, IsNil, Commentf("subtest %d, iter %d, err=%v", i, k, err))
			c.Assert(name, Equals, info.Name, Commentf("subtest %d, iter %d, name=%v, expected=%s, stack=%v, base=%v", i, k, name, info.Name, walker.stack, walker.base))
			c.Assert(entry.Mode, Equals, os.FileMode(mode), Commentf("subtest %d, iter %d, entry.Mode=%v expected=%v", i, k, entry.Mode, mode))
			c.Assert(obj.Type(), Equals, info.Kind, Commentf("subtest %d, iter %d, obj.Type()=%v expected=%v", i, k, obj.Type(), info.Kind))
			c.Assert(entry.Hash.String(), Equals, info.Hash, Commentf("subtest %d, iter %d, entry.Hash=%v, expected=%s", i, k, entry.Hash, info.Hash))
			c.Assert(obj.ID().String(), Equals, info.Hash, Commentf("subtest %d, iter %d, obj.ID()=%v, expected=%s", i, k, obj.ID(), info.Hash))
		}
		_, _, _, err = walker.Next()
		c.Assert(err, Equals, io.EOF)
	}
}
