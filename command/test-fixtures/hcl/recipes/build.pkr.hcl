
build {
  source "source.null.base" {
    name  = "tiramisu"
    // pull me up !
  }

  provisioner "shell-local" {
    name = "whipped_york"
    inline = [ "echo whip_york > ${upper(build.ID)}.${source.name}.txt" ]
  }
  provisioner "shell-local" {
    name = "mascarpone"
    inline = [ "echo mascarpone >> ${upper(build.ID)}.${source.name}.txt" ]
  }
  post-processor "shell-local" {
    name = "whipped_egg_white"
    inline = [ "echo whipped_egg_white >> ${upper(build.ID)}.${source.name}.txt" ]
  }
  post-processor "shell-local" {
    name = "dress_with_coffeed_boudoirs"
    inline = [ "echo dress >> ${upper(build.ID)}.${source.name}.txt" ]
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
    inline = [ "echo cooking... >> ${upper(build.ID)}.${source.name}.txt" ]
    except = ["null.spaghetti_carbonara"]
  }

}
