---
layout: "docs"
---

# Debugging Packer

Packer strives to be stable and bug-free, but issues inevitably arise where
certain things may not work entirely correctly, or may not appear to work
correctly. In these cases, it is sometimes helpful to see more details about
what Packer is actually doing.

Packer has detailed logs which can be enabled by setting the `PACKER_LOG`
environmental variable to any value. This will cause detailed logs to appear
on stderr. The logs contain log messages from Packer as well as any plugins
that are being used. Log messages from plugins are prefixed by their application
name.

Note that because Packer is highly parallelized, log messages sometimes
appear out of order, especially with respect to plugins. In this case,
it is important to pay attention to the timestamp of the log messages
to determine order.

If you find a bug with Packer, please include the detailed log by using
a service such as [gist](http://gist.github.com).
