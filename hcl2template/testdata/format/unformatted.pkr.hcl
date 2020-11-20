
// starts resources to provision them.
build {
    sources = [
        "source.amazon-ebs.ubuntu-1604",
        "source.virtualbox-iso.ubuntu-1204",
    ]

    provisioner "shell" {
        string  = coalesce(null, "", "string")
        int     = "${41 + 1}"
        int64   = "${42 + 1}"
        bool    = "true"
        trilean = true
        duration = "${9 + 1}s"
        map_string_string = {
            a = "b"
            c = "d"
        }
        slice_string = [
            "a",
            "b",
            "c",
        ]
        slice_slice_string = [
            ["a","b"],
            ["c","d"]
        ]

        nested {
            string  = "string"
            int     = 42
            int64   = 43
            bool    = true
            trilean = true
            duration = "10s"
            map_string_string = {
                a = "b"
                c = "d"
            }
            slice_string = [
                "a",
                "b",
                "c",
            ]
            slice_slice_string = [
                ["a","b"],
                ["c","d"]
            ]
        }

        nested_slice {
        }
    }

    provisioner "file" {
        string  = "string"
        int     = 42
        int64   = 43
        bool    = true
        trilean = true
        duration          = "10s"
        map_string_string = {
            a = "b"
            c = "d"
        }
        slice_string = [
            "a",
            "b",
            "c",
        ]
        slice_slice_string = [
            ["a","b"],
            ["c","d"]
        ]

        nested {
            string   = "string"
            int      = 42
            int64    = 43
            bool     = true
            trilean  = true
            duration = "10s"
            map_string_string = {
                a = "b"
                c = "d"
            }
            slice_string = [
                "a",
                "b",
                "c",
            ]
            slice_slice_string = [
                ["a","b"],
                ["c","d"]
            ]
        }

        nested_slice {
        }
    }

    post-processor "amazon-import" {
        string   = "string"
        int      = 42
        int64    = 43
        bool     = true
        trilean  = true
        duration = "10s"
        map_string_string = {
            a = "b"
            c = "d"
        }
        slice_string = [
            "a",
            "b",
            "c",
        ]
        slice_slice_string = [
            ["a","b"],
            ["c","d"]
        ]

        nested {
            string   = "string"
            int      = 42
            int64    = 43
            bool     = true
            trilean  = true
            duration = "10s"
            map_string_string = {
                a = "b"
                c = "d"
            }
            slice_string = [
                "a",
                "b",
                "c",
            ]
            slice_slice_string = [
                ["a","b"],
                ["c","d"]
            ]
        }

        nested_slice {
        }
    }
}
