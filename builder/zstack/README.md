#### environment to work:
ZStack is an IaaS platform which is management private cloud, so we need to deploy a ZStack cloud which version is greater than or equal to 3.4.0

#### basic usage:
1. go build -o bin/packer
2. bin/packer build examples/zstack/quick_start.json

#### premium usage:
1. if use old version ZStack(before 3.7.0), you must add "skip_packer_systemtag = true" into your json template
2. if you want to check if the template is validate, then you can use "bin/packer validate examples/zstack/quick_start.json", see: https://www.packer.io/docs/commands/validate.html
3. if you have some user variables, you can use "bin/packer build -var "access_key=tKYwaNrMbARP4vF15eCs" -var "key_secret=gZ0PGTIZcOAPBnnec5tLD5TXYO3R7Ml8Gk6niX2r" examples/zstack/quick_start.json" see: https://www.packer.io/docs/templates/user-variables.html
4. if you want to ssh to vm and execute provision, please choose public not private for 'l3network_uuid', see: https://www.zstack.io/help/product_manuals/user_guide/6.html#c6_4_3_1 (chinese)   http://en.zstack.io/blog/virtual-router.html (english for old version)

#### wordpress usage:
bin/packer build -var "basedir=/root/go/src/github.com/hashicorp/packer/examples/zstack/" examples/zstack/wordpress/data.json  
PS. it will build 1 node wordpress with datavolumes, and both master-slave mysql configs
PPS. basedir is the parent of wordpress dir
