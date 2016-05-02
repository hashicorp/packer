package git

import (
	"sort"

	"gopkg.in/src-d/go-git.v3/core"

	. "gopkg.in/check.v1"
)

type SuiteTree struct {
	repos map[string]*Repository
}

var _ = Suite(&SuiteTree{})

// create the repositories of the fixtures
func (s *SuiteTree) SetUpSuite(c *C) {
	treeFixtures := []packedFixture{
		{"https://github.com/tyba/git-fixture.git", "formats/packfile/fixtures/git-fixture.ofs-delta"},
		{"https://github.com/cpcs499/Final_Pres_P.git", "formats/packfile/fixtures/Final_Pres_P.ofs-delta"},
		{"https://github.com/jamesob/desk.git", "formats/packfile/fixtures/jamesob-desk.pack"},
		{"https://github.com/spinnaker/spinnaker.git", "formats/packfile/fixtures/spinnaker-spinnaker.pack"},
		{"https://github.com/alcortesm/binary-relations.git", "formats/packfile/fixtures/alcortesm-binary-relations.pack"},
		{"https://github.com/Tribler/dispersy.git", "formats/packfile/fixtures/tribler-dispersy.pack"},
	}
	s.repos = unpackFixtures(c, treeFixtures)
}

func (s *SuiteTree) TestFile(c *C) {
	for i, t := range []struct {
		repo     string // the repo name as in localRepos
		commit   string // the commit to search for the file
		path     string // the path of the file to find
		blobHash string // expected hash of the returned file
		size     int64  // expected size of the returned file
		found    bool   // expected found value
	}{
		// use git ls-tree commit to get the hash of the blobs
		{
			"https://github.com/tyba/git-fixture.git",
			"b029517f6300c2da0f4b651b8642506cd6aaf45d", "not-found",
			"", 0, false,
		},
		{
			"https://github.com/tyba/git-fixture.git",
			"b029517f6300c2da0f4b651b8642506cd6aaf45d", ".gitignore",
			"32858aad3c383ed1ff0a0f9bdf231d54a00c9e88", 189, true,
		},
		{
			"https://github.com/tyba/git-fixture.git",
			"b029517f6300c2da0f4b651b8642506cd6aaf45d", "LICENSE",
			"c192bd6a24ea1ab01d78686e417c8bdc7c3d197f", 1072, true,
		},
		{
			"https://github.com/tyba/git-fixture.git",
			"6ecf0ef2c2dffb796033e5a02219af86ec6584e5", "not-found",
			"", 0, false,
		},
		{
			"https://github.com/tyba/git-fixture.git",
			"6ecf0ef2c2dffb796033e5a02219af86ec6584e5", ".gitignore",
			"32858aad3c383ed1ff0a0f9bdf231d54a00c9e88", 189, true,
		},
		{
			"https://github.com/tyba/git-fixture.git",
			"6ecf0ef2c2dffb796033e5a02219af86ec6584e5", "binary.jpg",
			"d5c0f4ab811897cadf03aec358ae60d21f91c50d", 76110, true,
		},
		{
			"https://github.com/tyba/git-fixture.git",
			"6ecf0ef2c2dffb796033e5a02219af86ec6584e5", "LICENSE",
			"c192bd6a24ea1ab01d78686e417c8bdc7c3d197f", 1072, true,
		},
		{
			"https://github.com/tyba/git-fixture.git",
			"35e85108805c84807bc66a02d91535e1e24b38b9", "binary.jpg",
			"d5c0f4ab811897cadf03aec358ae60d21f91c50d", 76110, true,
		},
		{
			"https://github.com/tyba/git-fixture.git",
			"b029517f6300c2da0f4b651b8642506cd6aaf45d", "binary.jpg",
			"", 0, false,
		},
		{
			"https://github.com/tyba/git-fixture.git",
			"6ecf0ef2c2dffb796033e5a02219af86ec6584e5", "CHANGELOG",
			"d3ff53e0564a9f87d8e84b6e28e5060e517008aa", 18, true,
		},
		{
			"https://github.com/tyba/git-fixture.git",
			"1669dce138d9b841a518c64b10914d88f5e488ea", "CHANGELOG",
			"d3ff53e0564a9f87d8e84b6e28e5060e517008aa", 18, true,
		},
		{
			"https://github.com/tyba/git-fixture.git",
			"a5b8b09e2f8fcb0bb99d3ccb0958157b40890d69", "CHANGELOG",
			"d3ff53e0564a9f87d8e84b6e28e5060e517008aa", 18, true,
		},
		{
			"https://github.com/tyba/git-fixture.git",
			"35e85108805c84807bc66a02d91535e1e24b38b9", "CHANGELOG",
			"d3ff53e0564a9f87d8e84b6e28e5060e517008aa", 0, false,
		},
		{
			"https://github.com/tyba/git-fixture.git",
			"b8e471f58bcbca63b07bda20e428190409c2db47", "CHANGELOG",
			"d3ff53e0564a9f87d8e84b6e28e5060e517008aa", 18, true,
		},
		{
			"https://github.com/tyba/git-fixture.git",
			"b029517f6300c2da0f4b651b8642506cd6aaf45d", "CHANGELOG",
			"d3ff53e0564a9f87d8e84b6e28e5060e517008aa", 0, false,
		},
		// git submodule
		{
			"https://github.com/cpcs499/Final_Pres_P.git",
			"70bade703ce556c2c7391a8065c45c943e8b6bc3", "Final",
			"", 0, false,
		},
		{
			"https://github.com/cpcs499/Final_Pres_P.git",
			"70bade703ce556c2c7391a8065c45c943e8b6bc3", "Final/not-found",
			"", 0, false,
		},
		{
			"https://github.com/jamesob/desk.git",
			"d4edaf0e8101fcea437ebd982d899fe2cc0f9f7b", "LICENSE",
			"49c45e6cc893d6f5ebd5c9343fe4492360f339bf", 1058, true,
		},
		{
			"https://github.com/jamesob/desk.git",
			"d4edaf0e8101fcea437ebd982d899fe2cc0f9f7b", "examples",
			"", 0, false,
		},
		{
			"https://github.com/jamesob/desk.git",
			"d4edaf0e8101fcea437ebd982d899fe2cc0f9f7b", "examples/desk.sh",
			"d9c7751138824cd2d539c23d5afe3f9d29836854", 265, true,
		},
		{
			"https://github.com/jamesob/desk.git",
			"d4edaf0e8101fcea437ebd982d899fe2cc0f9f7b", "examples/not-found",
			"", 0, false,
		},
		{
			"https://github.com/jamesob/desk.git",
			"d4edaf0e8101fcea437ebd982d899fe2cc0f9f7b", "test/bashrc",
			"e69de29bb2d1d6434b8b29ae775ad8c2e48c5391", 0, true,
		},
		{
			"https://github.com/jamesob/desk.git",
			"d4edaf0e8101fcea437ebd982d899fe2cc0f9f7b", "test/not-found",
			"", 0, false,
		},
		{
			"https://github.com/spinnaker/spinnaker.git",
			"b32b2aecae2cfca4840dd480f8082da206a538da", "etc/apache2/sites-available/spinnaker.conf",
			"1d452c616be4fb16d2cc6b8a7e7a2208a6e64d2d", 67, true,
		},
		{
			"https://github.com/alcortesm/binary-relations.git",
			"c44b5176e99085c8fe36fa27b045590a7b9d34c9", "Makefile",
			"2dd2ad8c14de6612ed15813679a6554bad99330b", 1254, true,
		},
		{
			"https://github.com/alcortesm/binary-relations.git",
			"c44b5176e99085c8fe36fa27b045590a7b9d34c9", "src/binrels",
			"", 0, false,
		},
		{
			"https://github.com/alcortesm/binary-relations.git",
			"c44b5176e99085c8fe36fa27b045590a7b9d34c9", "src/map-slice",
			"", 0, false,
		},
		{
			"https://github.com/alcortesm/binary-relations.git",
			"c44b5176e99085c8fe36fa27b045590a7b9d34c9", "src/map-slice/map-slice.go",
			"12431e98381dd5097e1a19fe53429c72ef1f328e", 179, true,
		},
		{
			"https://github.com/alcortesm/binary-relations.git",
			"c44b5176e99085c8fe36fa27b045590a7b9d34c9", "src/map-slice/map-slice.go/not-found",
			"", 0, false,
		},
	} {
		commit, err := s.repos[t.repo].Commit(core.NewHash(t.commit))
		c.Assert(err, IsNil, Commentf("subtest %d: %v (%s)", i, err, t.commit))

		tree := commit.Tree()
		file, err := tree.File(t.path)
		found := err == nil

		comment := Commentf("subtest %d, path=%s, commit=%s", i, t.path, t.commit)
		c.Assert(found, Equals, t.found, comment)
		if !found {
			continue
		}

		c.Assert(file.Size, Equals, t.size, comment)
		c.Assert(file.Hash.IsZero(), Equals, false, comment)
		c.Assert(file.Hash, Equals, file.ID(), comment)
		c.Assert(file.Hash.String(), Equals, t.blobHash, comment)
	}
}

func (s *SuiteTree) TestFiles(c *C) {
	for i, t := range []struct {
		repo   string   // the repo name as in localRepos
		commit string   // the commit to search for the file
		files  []string // the expected files in the commit
	}{
		{"https://github.com/alcortesm/binary-relations.git", "b373f85fa2594d7dcd9989f4a5858a81647fb8ea", []string{
			"binary-relations.tex",
			".gitignore",
			"imgs-gen/simple-graph/fig.fig",
			"imgs-gen/simple-graph/Makefile",
			"Makefile",
			"src/map-slice/map-slice.go",
			"src/simple-arrays/simple-arrays.go",
		}},
		{"https://github.com/Tribler/dispersy.git", "f5a1fca709f760bf75a7adaa480bf0f0e1a547ee", []string{
			"authentication.py",
			"bloomfilter.py",
			"bootstrap.py",
			"cache.py",
			"callback.py",
			"candidate.py",
			"community.py",
			"conversion.py",
			"crypto.py",
			"database.py",
			"debugcommunity.py",
			"debug.py",
			"decorator.py",
			"destination.py",
			"dispersydatabase.py",
			"dispersy.py",
			"distribution.py",
			"dprint.py",
			"encoding.py",
			"endpoint.py",
			"__init__.py",
			"member.py",
			"message.py",
			"meta.py",
			"payload.py",
			"requestcache.py",
			"resolution.py",
			"script.py",
			"singleton.py",
			"timeline.py",
			"tool/callbackscript.py",
			"tool/__init__.py",
			"tool/scenarioscript.py",
		}},
		{"https://github.com/Tribler/dispersy.git", "9d38ff85ca03adcf68dc14f5b68b8994f15229f4", []string(nil)},
	} {
		commit, err := s.repos[t.repo].Commit(core.NewHash(t.commit))
		c.Assert(err, IsNil, Commentf("subtest %d: %v (%s)", i, err, t.commit))

		tree := commit.Tree()
		var output []string
		iter := tree.Files()
		defer iter.Close()
		for file, err := iter.Next(); err == nil; file, err = iter.Next() {
			c.Assert(file.Mode.String(), Equals, "-rw-r--r--")
			output = append(output, file.Name)
		}
		sort.Strings(output)
		sort.Strings(t.files)
		c.Assert(output, DeepEquals, t.files, Commentf("subtest %d, repo=%s, commit=%s", i, t.repo, t.commit))
	}
}
