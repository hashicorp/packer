
communicator "ssh" "vagrant" {
    string   = "string"
    int      = 42
    int64    = 43
    bool     = true
    trilean  = true
    duration = "10s"
    map_string_string {
        a = "b"
        c = "d"
    }
    slice_string = [
        "a",
        "b",
        "c",
    ]

    nested {
        string   = "string"
        int      = 42
        int64    = 43
        bool     = true
        trilean  = true
        duration = "10s"
        map_string_string {
            a = "b"
            c = "d"
        }
        slice_string = [
            "a",
            "b",
            "c",
        ]
    }

    nested_slice {
    }
}
