
source "file" "hello-world" {
    source = "${path.root}/../../../test-fixtures/hello.txt"
    target = "${path.root}/here/there/world.txt"
}

build {
    sources = ["source.file.hello-world"]
}
