# Packer Website

This subdirectory contains the entire source for the [Packer website](http://www.packer.io).
This is a [Middleman](http://middlemanapp.com) project, which builds a static
site from these source files.

## Contributions Welcome!

If you find a typo or you feel like you can improve the HTML, CSS, or
JavaScript, we welcome contributions. Feel free to open issues or pull
requests like any normal GitHub project, and we'll merge it in.

## Running the Site Locally

Running the site locally is simple. Clone this repo and run the following
commands:

```
make dev
```

Then open up `localhost:4567`. Note that some URLs you may need to append
".html" to make them work (in the navigation and such).

## Keeping Tidy

To keep the source code nicely formatted, there is a `make format` target. This
runs `htmlbeautify` and `pandoc` to reformat the source code so it's nicely formatted.

    make format

Note that you will need to install pandoc yourself. `make format` will skip it
if you don't have it installed.