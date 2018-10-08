builders = {
    type = "test"
  }
  
  provisioners = {
    "override" "test" {
      foo = "bar"
    }
  
    type = "test"
  }