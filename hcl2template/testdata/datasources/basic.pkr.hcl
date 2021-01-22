data "amazon-ami" "test" {
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
}