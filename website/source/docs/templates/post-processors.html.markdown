---
layout: "docs"
---

# Templates: Post-Processors

The post-processor section within a template configures any post-processing
that will be done to images built by the builders. Examples of post-processing
would be compressing files, uploading artifacts, etc.

Post-processors are _optional_. If no post-processors are defined within a template,
then no post-processing will be done to the image. The resulting artifact of
a build is just the image outputted by the builder.

This documentation page will cover how to configure a post-processor in a
template. The specific configuration options available for each post-processor,
however, must be referenced from the documentation for that specific post-processor.

Within a template, a section of post-processor definitions looks like this:

<pre class="prettyprint">
{
  "post-processors": [
    ... one or more post-processor definitions here ...
  ]
}
</pre>

For each post-processor definition, Packer will take the result of each of the
defined builders and send it through the post-processors. This means that if you
have one post-processor defined and two builders defined in a template, the
post-processor will run twice (once for each builder), by default. There are
ways, which will be covered later, to control what builders post-processors
apply to, if you wish.

## Post-Processor Definition

Within the `post-processors` array in a template, there are three ways to
define a post-processor. There are _simple_ definitions, _detailed_ definitions,
and _sequence_ definitions. Don't worry, they're all very easy to understand,
and the "simple" and "detailed" definitions are simply shortcuts for the
"sequence" definition.

A **simple definition** is just a string; the name of the post-processor. An
example is shown below. Simple definitions are used when no additional configuration
is needed for the post-processor.

<pre class="prettyprint">
{
  "post-processors": ["compress"]
}
</pre>

A **detailed definition** is a JSON object. It is very similar to a builder
or provisioner definition. It contains a `type` field to denote the type of
the post-processor, but may also contain additional configuration for the
post-processor. A detailed definition is used when additional configuration
is needed beyond simply the type for the post-processor. An example is shown below.

<pre class="prettyprint">
{
  "post-processors": [
    {
      "type": "compress",
      "format": "tar.gz"
    }
  ]
}
</pre>

A **sequence definition** is a JSON array comprised of other **simple** or
**detailed** definitions. The post-processors defined in the array are run
in order, with the artifact of each feeding into the next, and any intermediary
artifacts being discarded. A sequence definition may not contain another
sequence definition. Sequnce definitions are used to chain together multiple
post-processors. An example is shown below, where the artifact of a build is
compressed then uploaded, but the compressed result is not kept.

<pre class="prettyprint">
{
  "post-processors": [
    [
      "compress",
      { "type": "upload", "endpoint": "http://fake.com" }
    ]
  ]
}
</pre>

As you may be able to imagine, the **simple** and **detailed** definitions
are simply shortcuts for a **sequence** definition of only one element.
