Describe your change.

Use the following checklist to see when you're ready to merge (for docs-only changes you can skip this).

- [ ] For bugs, your first commit should be a test that reproduces the bug.
  This test should fail.
- [ ] Implement the change
- [ ] Add unit tests as appropriate
      Example: builder/virtualbox/common/ssh_config_test.go:19
- [ ] Add an acceptance test with a template example using your feature
      Example: post-processor/compress/post-processor_test.go:153
- [ ] Verify that tests pass
- [ ] For features, update / add documentation under `website/`
- [ ] Format your code with `go fmt`
- [ ] Rebase onto master
- [ ] If your PR resolves other open issues, indicate them below

Closes #xxx
Closes #xxx
