package git

import (
	"fmt"
	"io"
	"time"

	. "gopkg.in/check.v1"
	"gopkg.in/src-d/go-git.v3/core"
)

type expectedTag struct {
	Object      string          // tagged object hash
	Type        core.ObjectType // tagged object type
	Tag         string          // tag name
	TaggerName  string          // tag author name
	TaggerEmail string          // tag author email
	When        string          // tag time
	Message     string          // tag message
}

var tagFixtures = []packedFixture{
	{"https://github.com/spinnaker/spinnaker.git", "formats/packfile/fixtures/spinnaker-spinnaker.pack"},
}

var tagTests = []struct {
	repo string                 // the repo name in the test suite's map of fixtures
	tags map[string]expectedTag // the expected tags to test, mapped by tag hash
}{
	// https://api.github.com/repos/spinnaker/spinnaker/git/tags/TAGHASH
	{"https://github.com/spinnaker/spinnaker.git", map[string]expectedTag{
		"48b655898fa9c72d62e8dd73b022ecbddd6e4cc2": {"a77d88e40e86ae81b3ce1c19d04fd73f473f5644", core.CommitObject, "v0.13.0", "cfieber", "cfieber@netflix.com", "2015-11-20T19:37:31Z", "Release of 0.13.0\n\n- a77d88e40e86ae81b3ce1c19d04fd73f473f5644: Merge pull request #606 from duftler/check-log-dirs\n- a174b873e97fb9a2d551d007c92aa5889c081a99: Always check if log dirs exist before starting services via spinnaker umbrella service. Stop/start spinnaker instead of restart since restart seems not happy about a service that is not already running.\n- 8586b7cd3f70fe63053fd5fa321bc86c6b803622: Merge pull request #603 from ewiseblatt/remove_obsolete_scripts\n- 3525c938ab51af81cff2448c8c784b925af2fd0f: Merge pull request #604 from kenzanlabs/kenzan-master\n- b1b5146a77d363e136336923429134d0759eb9c8: Script to generate ami_table.md and ami_table.json files after the jenkins ami build is complete\n- 1ef157853d770a26e7682e543ac42de485b34f77: Removed obsolete scripts.\n"},
		"82562fa518f0a2e2187ea2604b07b67f2e7049ae": {"1ea743cd62e8e60f97f55a434a3f46400b49f606", core.CommitObject, "v0.12.0", "cfieber", "cfieber@netflix.com", "2015-11-19T22:04:22Z", "Release of 0.12.0\n\n- 1ea743cd62e8e60f97f55a434a3f46400b49f606: Merge pull request #599 from ewiseblatt/no_distribution\n- 8fe3f13ad04ee25fde0add4ed19d29acd49a5916: Up GCE source image to trusty-v20151113\n- 855e3b979f1d65fbfbcc68df905dafb9945f3825: Merge pull request #601 from ewiseblatt/no_longer_tar\n- d79f2736da55c123ba638710284e2856041262a5: Removed obsolete tar packaging.\n- 2b1ab713af3789204594a45f265cc93858807e98: Merge pull request #600 from ewiseblatt/remove_bootstrap\n- 52831ed7689ab0f481486f62e81d2b4e9e1c535b: Removed obsolete BootstrapSpinnaker.sh\n- 1b44b5467f78a3f5e1915b6fe78f7d0814c29427: Merge pull request #595 from duftler/sync-settings-file\n- 637ba49300f701cfbd859c1ccf13c4f39a9ba1c8: Sync feature block in settings.js.\n- 4f3c7375fa7c661735a6a69beeeeac1aaa43f7c9: Merge pull request #593 from ewiseblatt/2_run_dev_autoconfigure\n- a74422026841e05debdcc417190428b419a99f39: Merge pull request #564 from mstantoncook/master\n- d73f9cee49a5ad27a42a6e18af7c49a8f28ad8a8: Auto-generate spinnaker-local.yml on run_dev first's run.\n- b260ce026a2505037876b4c21c0985882ff373b7: Merge pull request #591 from ewiseblatt/1_move_transform_yaml\n- bb6325e4e629fc7348a6d0e6842280d5304160ff: Moved yaml transform method.\n- 608976766959bdb1b18eaa53b3ca33ee6782bc3c: Merge pull request #590 from kenzanlabs/master\n- cfdd19354e2a3981484a7cfe4b0d95c9abce9296: Merge pull request #589 from ewiseblatt/readme\n- 8ef83dd443a05e9122681950399edaa58a38d466: Updated run_dev instructions.\n- 769ce2a32e60bf2219ffb5b8467d62f71f1e4877: Merge pull request #1 from skorten/cassandra-thrift-fix\n- b2c7142082d52b09ca20228606c31c7479c0833e: using apt-mark to put a hold on cassandra packages so they will not be get upgraded from 2.1 and break thrift\n- d25148149d6a67989be79cdb7452cdab8d2f1a4b: Merge pull request #586 from ewiseblatt/reconfigure\n- c89dab0d42f1856d157357e9010f8cc6a12f5b1f: Fixes reconfigure_spinnaker.sh to behave properly when not run as root.\n- 8a9804234551d61209f67b3c89f7706f248ae805: Merge pull request #581 from ewiseblatt/03_fix_run_dev\n- b45ffa99a6daaf045043ab0b0d8bcf823f10e157: Merge pull request #580 from ewiseblatt/02_create_dev\n- 827682091dd09c1887e82686e36822695b88bb1e: Merge pull request #579 from ewiseblatt/01_install_dev\n- 4f9cd01b6e533c3b1261660b9cc3302879e5b303: Merge pull request #554 from ewiseblatt/fix_transform\n- d1ff4e13e9e0b500821aa558373878f93487e34b: Refactored install_development.sh to use production install\n- 1c370109898641253617a4d48d77f2c9b0a4ccf5: Merge pull request #584 from dpeterka/master\n- 8d1e069744321ff97cbefeaba593c778105c3aad: Cosmetic changes. Fix ports in login script\n- dd7e66c862209e8b912694a582a09c0db3227f0d: Update InstallSpinnaker.sh\n- 4cce5f988005be72dca910fb53e4b2f5802bf7cf: need to add front50 url to clouddriver in configs\n- 0ae9771322873f03893180d90b0af5e3b30154e9: Merge pull request #583 from dpeterka/master\n- e805183c72f0426fb073728c01901c2fd2db1da6: Reconfigure AWS only on boot\n- f98b6099746b849abfb9d5b1db7e861363747be2: Consistent naming for packages\n- 52edbd4c10193f87f8f9768c92789637bfedb867: Don't prompt, just install packer so it's available for rush.\n- d7a3eedbf9fa133d7c4366afae555a2ed46d4849: Merge pull request #582 from saulshanabrook/patch-1\n- 9944d6cf72b8f82d622d85dad7434472bc8f397d: Fix readme link\n- 6694fb99ca6fbf469798f1fb9386b55ff80f0128: Merge pull request #578 from spinnaker/readme-cleanup\n- 174bdbf9edfb0ca88415dd4a673852d5b22e7036: Remove note about packer.\n- 2b28ea424acc8f2817d3298c143fae68bcad91a7: Fix run_dev\n- 206033f8afb2609982fdc6e929a94a340bc80054: Updated create_google_dev_vm (and run_dev)\n- 811795c8a185e88f5d269195cb68b29c8d0fe170: Update README.adoc\n- 4584fab37e93d66fd1896d07fa3427f8056711bc: Removed redudnant attribute\n"},
		"3e349f806a0d02bf658c3544c46a0a7a9ee78673": {"6ea37d18b706aab813532254ce0d412843c68782", core.CommitObject, "v0.11.0", "cfieber", "cfieber@netflix.com", "2015-11-17T22:07:00Z", "Release of 0.11.0\n\n- 6ea37d18b706aab813532254ce0d412843c68782: Merge pull request #571 from dpeterka/changeRepoName\n- fad219f07e362f97eda945790320f1f0552a919c: Merge pull request #574 from duftler/always-install-packer\n- 376599177551c3f04ccc94d71bbb4d037dec0c3f: Don't prompt, just install packer so it's available for rush.\n- 9414750a933037ec4f0bc42af7ad81ec4f360c0a: Merge pull request #572 from erjohnso/master\n- d6e6fe0194447cc280f942d6a2e0521b68ea7796: Point non-devs to getting started user docs\n- e259e024b1c7a221e8329fb942a4992738bc81af: update docker compose to use /opt/spinnaker/config\n- b32b2aecae2cfca4840dd480f8082da206a538da: Merge pull request #566 from ewiseblatt/refactor_install\n- 8eed01ff4f2ef7c9c68ab031b54e0cf84a0b1cc9: Consistent naming for packages\n- 66ee9032d57be4bac236edec0e501aaa0501a57d: Merge pull request #570 from spinnaker/cleanup-instructions\n- 24551a5d486969a2972ee05e87f16444890f9555: Update instructions.\n- d4b48a39aba7d3bd3e8abef2274a95b112d1ae73: Add option to only install dependencies without installing spinnaker services.\n- 5ad50e028c59d67ae5d8160e685947582dc68f36: Merge pull request #569 from analytically/master\n- 9a06d3f20eabb254d0a1e2ff7735ef007ccd595e: Fix Ubuntu version.\n- c0a70a0f5aa494f0ae01c55ba191f2325556489a: change heading to setting up spinnaker for development\n- d6905eab6fec1841c7cf8e4484499f5c8d7d423e: update Readme to point to the getting started guide\n- f5300bb86b22eda66eb4baef6b2a211c85f14690: Merge pull request #560 from ewiseblatt/autogen_packages\n- d3046b5b2f7aafa0832da6806ee8c7dab7d0da9e: Merge pull request #559 from ewiseblatt/remove_obsolete_instructions\n- ca87222cb609773c56d43c960e8f0ade554fc138: Removed obsolete instructions output.\n- bd42370d3fe8d410e78acb96f81cb3d838ad1c21: change url for join slack button\n- 67f0a0f488b3592bb611391150f2e1d0ee037231: Merge pull request #558 from gregturn/convert-to-asciidoc\n- 638f61b3331695f46f1a88095e26dea0f09f176b: Convert README to asciidoctor\n- 09a4ea729b25714b6368959eea5113c99938f7b6: Generate bintray packages if needed.\n- 8731e9edc1619e798a76fedb30b26cf48fa62897: Merge pull request #555 from dpeterka/master\n- bcbbd656c19dbc47ffd5b247927ea99f3949c78a: Add VPC Scripts\n"},
		"d081d66c2a76d04ff479a3431dc36e44116fde40": {"e0005f50e22140def60260960b21667f1fdfff80", core.CommitObject, "v0.10.0", "cfieber", "cfieber@netflix.com", "2015-11-16T15:25:36Z", "Release of 0.10.0\n\n- e0005f50e22140def60260960b21667f1fdfff80: Merge pull request #553 from ewiseblatt/rendezvous\n- e1a2b26b784179e6903a7ae967c037c721899eba: Wait for cassandra before starting spinnaker\n- c756e09461d071e98b8660818cf42d90c90f2854: Merge pull request #552 from duftler/google-c2d-tweaks\n- 0777fadf4ca6f458d7071de414f9bd5417911037: Fix incorrect config prop names:   s/SPINNAKER_GOOGLE_PROJECT_DEFAULT_REGION/SPINNAKER_GOOGLE_DEFAULT_REGION   s/SPINNAKER_GOOGLE_PROJECT_DEFAULT_ZONE/SPINNAKER_GOOGLE_DEFAULT_ZONE Hardcode profile name in generated ~/.aws/credentials to [default]. Restart all of spinnaker after updating cassandra and reconfiguring spinnaker, instead of just restarting clouddriver.\n- d8d031c1ac45801074418c43424a6f2c0dff642c: Merge pull request #551 from kenzanmedia/fixGroup\n- 626d23075f9e92aad19015f2964c95d45f41fa3a: Put in correct block for public image. Delineate cloud provider.\n"},
		"776914ef8a097f5683957719c49215a5db17c2cb": {"c24f0caac157254e480055fb605a71465d13bc00", core.CommitObject, "v0.9.0", "cfieber", "cfieber@netflix.com", "2015-11-16T09:34:54Z", "Release of 0.9.0\n\n- c24f0caac157254e480055fb605a71465d13bc00: Merge pull request #549 from spinnaker/duftler-patch-1\n- 7622add2bc8c47d1a37244f39b94bcc187bf671d: Merge pull request #550 from spinnaker/duftler-patch-2\n- a57b08a9072f6a865f760551be2a4944f72f804a: Same thing, different day.\n- 50d0556563599366f29cb286525780004fa5a317: Redirect more stuff.\n"},
		"8526c58617f68de076358873b8aa861a354b48a9": {"f69376bd065db787894bd2775d447c8d87d3b50c", core.CommitObject, "v0.8.0", "cfieber", "cfieber@netflix.com", "2015-11-16T09:25:50Z", "Release of 0.8.0\n\n- f69376bd065db787894bd2775d447c8d87d3b50c: Merge pull request #547 from spinnaker/duftler-patch-1\n- 3b0f2a5fbc354b116452e9f3e366af74ce1f1321: Merge pull request #548 from spinnaker/duftler-patch-2\n- 2a3b1d3b134e937c7bafdab6cc2950e264bf5dee: Redirect nodetool output to /dev/null.\n- 4bbcad219ec55a465fb48ce236cb10ca52d43b1f: Redirect nodetool output to /dev/null.\n"},
		"3f36d8f1d67538afd1f089ffd0d242fc4fda736f": {"0ce1393c24c7083ec7f9f04b4cf461c047ad2192", core.CommitObject, "v0.7.0", "cfieber", "cfieber@netflix.com", "2015-11-16T09:06:36Z", "Release of 0.7.0\n\n- 0ce1393c24c7083ec7f9f04b4cf461c047ad2192: Merge pull request #546 from ewiseblatt/fix_race\n- dd2d03c19658ff96d371aef00e75e2e54702da0e: retry nodetool\n"},
		"dc22e2035292ccf020c30d226f3cc2da651773f6": {"46670eb6477c353d837dbaba3cf36c5f8b86f037", core.CommitObject, "v0.6.0", "cfieber", "cfieber@netflix.com", "2015-11-16T09:02:03Z", "Release of 0.6.0\n\n- 46670eb6477c353d837dbaba3cf36c5f8b86f037: Merge pull request #543 from spinnaker/default_repo_url\n- 2b20a9a5149deadbe43227b70445bf6699fd3a3a: Merge pull request #545 from spinnaker/cleanup\n- 92e5c1a4fb59d01ece44004c4e1daa78fa4b7f87: remove micronolith and buildDeb from experimental - superceded by top level gradle build\n- 99280af2aaf171fe056400938ae2dbf6d93d3736: Merge pull request #544 from ewiseblatt/fix_race\n- 495c7118e7cf757aa04eab410b64bfb5b5149ad2: Wait for cassandra to come up before calling nodetool\n- a47d0aaeda421f06df248ad65bd58230766bf118: Changes the default package repo to spinnaker/debians. Adds gpg key from the bintray org\n- 079e42e7c979541b6fab7343838f7b9fd4a360cd: Put back config that shouldn't have been removed.\n"},
		"0a3fb06ff80156fb153bcdcc58b5e16c2d27625c": {"b7b9e7c464c3c343133ed17e778a2f600b5863b8", core.CommitObject, "v0.5.0", "cfieber", "cfieber@netflix.com", "2015-11-16T07:43:35Z", "Release of 0.5.0\n\n- b7b9e7c464c3c343133ed17e778a2f600b5863b8: lastest deb publishing\n- 03e24883d2f0a60419b0d43074aa2b3341bb2a97: Merge pull request #542 from kenzanmedia/updateHelper\n- 237166c72299ec287d1f6ab96aea7af07a2df160: Print host name. Fix typo\n- 855c220530cb8aa8e9ff2598fc873240bf4a543b: Merge pull request #541 from ewiseblatt/postinstall\n- cacc42e050181fd1f74069a41c642f038d395c2d: reconfigure after installing configuration\n- d98e03d4dc5d87aa5a1b2a5dd74feb14de965128: Merge pull request #540 from kenzanmedia/packerAmiHelp\n- 82c28940a4b2d6a7e03c9349a7c2a37c9e164810: update helper text\n- bb702a749521496ea7e542df78806671d8d8c657: Merge pull request #539 from ewiseblatt/fix_configurator\n- 023d4fb17b76e0fe0764971df8b8538b735a1d67: Add environment variabels to configurator.\n- 125eceff9807da34b6e6ad7888441e7d9b7d629b: Merge pull request #538 from dstengle/aws-auto\n- 8a594011096b65f5b455254f95d2c7d99ec64c11: Merge pull request #537 from kenzanmedia/packerAmi\n- 01575d8fc3845c69bbf522f93cc4189f436eaf8a: Initial commit of packer role\n- b41d7c0e5b20bbe7c8eb6606731a3ff68f4e3941: - Auto detect aws - Add shell defaults for packer for vpc and subnet\n- 66d1c8f2fa2e32c2c936679c8b10e2134b2ac187: Merge pull request #534 from ewiseblatt/fix_upgrade\n- 6eb5d9c5225224bfe59c401182a2939d6c27fc00: Fix thrift after upgrades\n- 46ac02f5fbeb5e9c026bd85ee56d828836e0c323: Merge pull request #536 from kenzanmedia/fixApachePort\n- 23da1763950b26aaa23551d798e3a52f1526fcc6: Make sure Apache listens only on localhost\n- 36152fb0265180a42b7a79be31848de9845a81b5: Merge pull request #535 from duftler/set-project-id\n- ba486de7c025457963701114c683dcd4708e1dee: If we have the Google project id from the environment, set it in /etc/default/spinnaker regardless of google.enabled's value.\n- 743a148328362ff93312329de0165fab07641546: Merge pull request #533 from ewiseblatt/fix_upgrade\n- c4a9091e4076cb740fa46e790dd5b658e19012ad: fix cassandra after upgrade\n- b5c6053a46993b20d1b91e7b7206bffa54669ad7: Fix packer install.\n- 505577dc87d300cf562dc4702a05a5615d90d855: Remove spaces.\n- 921a8a191aff8b0333c08ab78803878fdc26e9f5: Merge pull request #532 from duftler/more-auto-creds\n- 370d61cdbc1f3c90db6759f1599ccbabd40ad6c1: Improve handling of writing to /etc/default/spinnaker.\n- ebe1cd8da4246d8b9b3f1c4717e99309a00490f6: Merge pull request #531 from ewiseblatt/fix_permissions\n- 9467ec579708b3c71dd9e5b3906772841c144a30: Fixed what looks like merge error\n- 88e841aad37b71b78a8fb88bc75fe69499d527c7: Remove hard-coded aws metadata value.\n- bbeb98f59f4f0b373c7d764964d8c23522804ef9: Merge pull request #530 from duftler/auto-creds\n- 8eb116de9128c314ac8a6f5310ca500b8c74f5db: Improveed param handling. Added placeholders for aws env work.\n- 26a83567b8d80ed7523fc1d5d1e13d3f095bb70d: Merge pull request #529 from duftler/auto-credentials\n- 8980daf661408a3faa1f22c225702a5c1d11d5c9: Patch in ewiseblatt's work from the pending PR. Add support for deriving project id, region and zone from GCE environment. Get generic bits in place to derive aws creds from env as well. Derived parameters are used as defaults for each prompt. Explicitly passed command-line arguments take precedence.\n- 66ac94f0b4442707fb6f695fbed91d62b3bd9d4a: Merge pull request #528 from ewiseblatt/c2d\n- a596972a661d9a7deca8abd18b52ce1a39516e89: Added transformation from old yaml to new yaml for click to deploy.\n- 0a67d98c7a0eaa27bf6b62450f6f54aadbb961ed: Merge pull request #527 from duftler/support-bootstrap\n- 5422a86a10a8c5a1ef6728f5fc8894d9a4c54cb9: Rewrite url in BootstrapSpinnaker.sh when build_release.py is run. Push BootstrapSpinnaker.sh to bintray repo alongside InstallSpinnaker.sh.\n- 3de4f77c105f700f50d9549d32b9a05a01b46c4b: sudo start spinnaker !!\n- 7119ad9cf7d4e4d8b059e5337374baae4adc7458: Add orca-baseurl\n- d4553dac205023fa77652308af1a2d1cf52138fb: Add default value for services.rush.configDir.\n- 304cac16bddf7bfbcc1663bf408ac452d29762f2: Merge pull request #526 from duftler/one-line-install\n- 6328ee836affafc1b52127147b5ca07300ac78e6: Create one-line installation command.\n- 23a14bd9cbe1808001a88ce8218d5b6d0948fa8a: Merge pull request #524 from kenzanmedia/kenzan-develop\n- 01e65d67eed8afcb67a6bdf1c962541f62b299c9: Open JDK 8 from a PPA\n- 18fc95490bcee25a4669d9ab7640e729cef32df4: Merge pull request #520 from ewiseblatt/post_install\n- ec0ff22492361ac3a9c5c6c49a223adbb4afeb7a: Add deprecated paths\n- 889c1f2bafc2f74258f608a9beab14dd4a70edb9: Merge pull request #522 from ewiseblatt/build_release\n- a6a4e6112a5009e53a37f7325620a93db7eadd9c: Merge pull request #523 from ewiseblatt/build_google_image\n- e51871f45f3848ec1ed37aab052277198c98fff1: Merge pull request #521 from ewiseblatt/first_boot\n- f66196ceed7d6aeca313b0632657ab762487ced3: build google_image\n- de25f576b888569192e6442b0202d30ca7b2d8ec: Updated first_boot\n- 6986d885626792dee4ef6b7474dfc9230c5bda54: less brittle build_release\n"},
		"95ee6e6c750ded1f4dc5499bad730ce3f58c6c3a": {"2c748387f5e9c35d001de3c9ba3072d0b3f10a72", core.CommitObject, "v0.4.0", "cfieber", "cfieber@netflix.com", "2015-11-15T18:18:51Z", "Release of 0.4.0\n\n- 2c748387f5e9c35d001de3c9ba3072d0b3f10a72: Create LICENSE.txt\n- 0c6968e24796a67fa602c94d7a82fd4fd375ec59: Create AUTHORS\n- 3ce7b902a51bac2f10994f7d1f251b616c975e54: Drop trailing slash.\n"},
		"8b6002b614b454d45bafbd244b127839421f92ff": {"65e37611b1ff9cb589e3060507427a9a2645907e", core.CommitObject, "v0.3.0", "cfieber", "cfieber@netflix.com", "2015-11-15T18:07:03Z", "Release of 0.3.0\n\n- 65e37611b1ff9cb589e3060507427a9a2645907e: need java plugin for manifest generation\n- bc02440df2ff95a014a7b3cb11b98c3a2bded777: newer gradle plugin\n- 791bcd1592828d9d5d16e83f3a825fb08b0ba22d: Don't need to rename the service.\n"},
	}},
	// TODO: Add fixture with tagged trees
	// TODO: Add fixture with tagged blobs
}

type SuiteTag struct {
	repos map[string]*Repository
}

var _ = Suite(&SuiteTag{})

func (s *SuiteTag) SetUpSuite(c *C) {
	s.repos = unpackFixtures(c, tagFixtures)
}

func (s *SuiteTag) TestCommit(c *C) {
	for _, t := range tagTests {
		r, ok := s.repos[t.repo]
		c.Assert(ok, Equals, true)
		k := 0
		for hash, expected := range t.tags {
			if expected.Type != core.CommitObject {
				continue
			}
			tag, err := r.Tag(core.NewHash(hash))
			c.Assert(err, IsNil)
			commit, err := tag.Commit()
			c.Assert(err, IsNil)
			c.Assert(commit.Type(), Equals, core.CommitObject)
			c.Assert(commit.Hash.String(), Equals, expected.Object)
			k++
		}
	}
}

func (s *SuiteTag) TestTree(c *C) {
	for _, t := range tagTests {
		r, ok := s.repos[t.repo]
		c.Assert(ok, Equals, true)
		k := 0
		for hash, expected := range t.tags {
			if expected.Type != core.TreeObject {
				continue
			}
			tag, err := r.Tag(core.NewHash(hash))
			c.Assert(err, IsNil)
			tree, err := tag.Tree()
			c.Assert(err, IsNil)
			c.Assert(tree.Type(), Equals, core.TreeObject)
			c.Assert(tree.Hash.String(), Equals, expected.Object)
			k++
		}
	}
}

func (s *SuiteTag) TestBlob(c *C) {
	for _, t := range tagTests {
		r, ok := s.repos[t.repo]
		c.Assert(ok, Equals, true)
		k := 0
		for hashString, expected := range t.tags {
			if expected.Type != core.BlobObject {
				continue
			}
			hash := core.NewHash(hashString)
			tag, err := r.Tag(hash)
			c.Assert(err, IsNil)
			testTagExpected(c, tag, hash, expected, "")
			blob, err := tag.Blob()
			c.Assert(err, IsNil)
			c.Assert(blob.Type(), Equals, core.BlobObject)
			c.Assert(blob.Hash.String(), Equals, expected.Object)
			k++
		}
	}
}

func (s *SuiteTag) TestObject(c *C) {
	for _, t := range tagTests {
		r, ok := s.repos[t.repo]
		c.Assert(ok, Equals, true)
		k := 0
		for hashString, expected := range t.tags {
			if expected.Type != core.BlobObject {
				continue
			}
			hash := core.NewHash(hashString)
			tag, err := r.Tag(hash)
			c.Assert(err, IsNil)
			testTagExpected(c, tag, hash, expected, "")
			obj, err := tag.Object()
			c.Assert(err, IsNil)
			c.Assert(obj.Type(), Equals, expected.Type)
			c.Assert(obj.ID().String(), Equals, expected.Object)
			k++
		}
	}
}

func testTagExpected(c *C, tag *Tag, hash core.Hash, expected expectedTag, comment string) {
	when, err := time.Parse(time.RFC3339, expected.When)
	c.Assert(err, IsNil)
	c.Assert(tag, NotNil)
	c.Assert(tag.Hash.IsZero(), Equals, false)
	c.Assert(tag.Hash, Equals, tag.ID())
	c.Assert(tag.Hash, Equals, hash)
	c.Assert(tag.Type(), Equals, core.TagObject)
	c.Assert(tag.TargetType, Equals, expected.Type, Commentf("%stype=%v, expected=%v", comment, tag.TargetType, expected.Type))
	c.Assert(tag.Target.String(), Equals, expected.Object, Commentf("%sobject=%v, expected=%s", comment, tag.Target, expected.Object))
	c.Assert(tag.Name, Equals, expected.Tag, Commentf("subtest %d, iter %d, tag=%s, expected=%s", comment, tag.Name, expected.Tag))
	c.Assert(tag.Tagger.Name, Equals, expected.TaggerName, Commentf("subtest %d, iter %d, tagger.name=%s, expected=%s", comment, tag.Tagger.Name, expected.TaggerName))
	c.Assert(tag.Tagger.Email, Equals, expected.TaggerEmail, Commentf("subtest %d, iter %d, tagger.email=%s, expected=%s", comment, tag.Tagger.Email, expected.TaggerEmail))
	c.Assert(tag.Tagger.When.Equal(when), Equals, true, Commentf("subtest %d, iter %d, tagger.when=%s, expected=%s", comment, tag.Tagger.When, when))
	c.Assert(tag.Message, Equals, expected.Message, Commentf("subtest %d, iter %d, message=\"%s\", expected=\"%s\"", comment, tag.Message, expected.Message))
}

func testTagIter(c *C, iter *TagIter, tags map[string]expectedTag, comment string) {
	for k := 0; k < len(tags); k++ {
		comment = fmt.Sprintf("%siter %d: ", comment, k)
		tag, err := iter.Next()
		c.Assert(err, IsNil)
		c.Assert(tag, NotNil)

		c.Assert(tag.Hash.IsZero(), Equals, false)

		expected, ok := tags[tag.Hash.String()]
		c.Assert(ok, Equals, true, Commentf("%sunexpected tag hash=%v", comment, tag.Hash))

		testTagExpected(c, tag, tag.Hash, expected, comment)
	}
	_, err := iter.Next()
	c.Assert(err, Equals, io.EOF)
}
