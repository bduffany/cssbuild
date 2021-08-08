# cssbuild

A fast CSS compiler supporting a limited subset of
[CSS modules](https://github.com/css-modules/css-modules) functionality.

## Features

- [x] Generated class selectors and animation names are suffixed with
      a random token to effectively make them locally scoped by default
- [x] Local scoping can be switched off via a `:global` mode selector or
      `:global()` function
- [x] Generates JS files with class name and animation name mappings

## What's not supported

- Local `@import` and `@url`
- Class composition

## Installation

```bash
go get -u github.com/bduffany/cssbuild
```

## Running

```bash
# Convert a CSS module stylesheet to vanilla CSS.
cssbuild -in src/styles.module.css -out dist/styles.css

# The above command writes JS to `dist/styles.module.css.js` by default.
# You can control this with the `-js_out` flag:
cssbuild -in src/styles.module.css -out dist/styles.css -js_out dist/styles.js

# See all options with documentation:
cssbuild -help
```

## More details

- The map keys in the generated JS file can optionally be made camelCase,
  using the `-camel_case_js_keys` flag, even if class names are kebab-case.
  This makes it easier to migrate to CSS modules (no need to waste time
  rewriting existing CSS or break with convention by requiring all CSS
  classes to be camelCase).
- The `:global` mode selector applies to the rules block, which allows
  referencing global animation names.
- Animation scoping supports `-webkit-` and `-moz-` prefixes.

## Thanks to

- [tdewolff/parse](https://github.com/tdewolff/parse) for the excellent
  CSS parsing library, which made it possible to implement this with
  minimal custom parsing.
- [evanw/esbuild](https://github.com/evanw/esbuild) for inspiration.
