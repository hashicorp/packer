
variables {
    key = "value"
    my_secret = "foo"
    image_name = "foo-image-{{user `my_secret`}}"
}
