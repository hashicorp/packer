
behavior "regexp_issue_labeler" "panic_label" {
    regexp = "panic:"
    labels = ["crash", "bug"]
}

behavior "remove_labels_on_reply" "remove_stale" {
    labels = ["waiting-reply", "stale"]
    only_non_maintainers = true
}

poll "closed_issue_locker" "locker" {
  schedule             = "0 50 1 * * *"
  closed_for           = "720h" # 30 days
  max_issues           = 500
  sleep_between_issues = "5s"
  no_comment_if_no_activity_for = "4320h" # 180 days

  message = <<-EOF
    I'm going to lock this issue because it has been closed for _30 days_ ⏳. This helps our maintainers find and focus on the active issues.

    If you have found a problem that seems similar to this, please open a new issue and complete the issue template so we can capture all the details necessary to investigate further.
  EOF
}

poll "label_issue_migrater" "remote_plugin_migrater" {
  schedule                = "0 20 * * * *"
  new_owner               = "hashicorp"
  repo_prefix             = "packer-plugin-"
  label_prefix            = "remote-plugin/"
  excluded_label_prefixes  = ["communicator/"]
  excluded_labels         = ["build", "core", "new-plugin-contribution", "website"]

  issue_header     = <<-EOF
    _This issue was originally opened by @${var.user} as ${var.repository}#${var.issue_number}. It was migrated here as a result of the [Packer plugin split](###blog-post-url###). The original body of the issue is below._

    <hr>

    EOF
  migrated_comment = "This issue has been automatically migrated to ${var.repository}#${var.issue_number} because it looks like an issue with that plugin. If you believe this is _not_ an issue with the plugin, please reply to ${var.repository}#${var.issue_number}."
}

