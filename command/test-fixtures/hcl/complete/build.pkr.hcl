
build {
  source "source.null.base" {
    name  = "tiramisu"
  }
}

build {
  name = "recipes"
  source "source.null.base" {
    name   = "spaghetti_carbonara"
  }
  source "source.null.base" {
    name   = "lasagna"
  }

  provisioner "shell-local" {
    name = "add_spaghetti"
    inline = [ "echo spaghetti > ${upper(build.ID)}.${source.name}.txt" ]
    only = ["null.spaghetti_carbonara"]
  }

  post-processor "shell-local" {
    name = "carbonara_it"
    inline = [ "echo carbonara >> ${upper(build.ID)}.${source.name}.txt" ]
    except = ["null.lasagna"]
  }


  provisioner "shell-local" {
    name = "add_lasagna"
    inline = [ "echo lasagna > ${upper(build.ID)}.${source.name}.txt" ]
    only = ["null.lasagna"]
  }

  provisioner "shell-local" {
    name = "add_tomato"
    inline = [ "echo tomato >> ${upper(build.ID)}.${source.name}.txt" ]
    except = ["null.spaghetti_carbonara"]
  }

  provisioner "shell-local" {
    name = "add_mozza"
    inline = [ "echo mozza >> ${upper(build.ID)}.${source.name}.txt" ]
    except = ["null.spaghetti_carbonara"]
  }

  post-processor "shell-local" {
    name = "cook"
    inline = [ "echo 'cooking...' >> ${upper(build.ID)}.${source.name}.txt" ]
    except = ["null.spaghetti_carbonara"]
  }

}
