
behavior "regexp_issue_labeler" "panic_label" {
    regexp = "panic:"
    labels = ["crash", "bug"]
}

behavior "remove_labels_on_reply" "remove_stale" {
    labels = ["waiting-reply", "stale"]
    only_non_maintainers = true
}

poll "label_issue_migrater" "remote_plugin_migrater" {
  schedule                = "0 20 * * * *"
  new_owner               = "hashicorp"
  repo_prefix             = "packer-plugin-"
  label_prefix            = "remote-plugin/"
  excluded_label_prefixes  = ["communicator/"]
  excluded_labels         = ["build", "core", "new-plugin-contribution", "website"]

  issue_header     = <<-EOF
    _This issue was originally opened by @${var.user} as ${var.repository}#${var.issue_number}. It was migrated here as a result of the [Packer plugin split](https://github.com/hashicorp/packer/issues/8610#issuecomment-770034737). The original body of the issue is below._

    <hr>

    EOF
  migrated_comment = "This issue has been automatically migrated to ${var.repository}#${var.issue_number} because it looks like an issue with that plugin. If you believe this is _not_ an issue with the plugin, please reply to ${var.repository}#${var.issue_number}."
}

