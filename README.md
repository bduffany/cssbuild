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

## More details

- The map keys in the generated JS file can optionally be camelCase, even
  if class names are kebab-case. This makes it easier to migrate to
  CSS modules (no need to waste time rewriting existing CSS or break with
  convention by requiring all CSS classes to be camelCase).
- `:global` mode selector applies to the rules block, which allows
  referencing global animation names
- Animation scoping supports legacy `-webkit-` prefixes

## Thanks to

- [tdewolff/parse](https://github.com/tdewolff/parse) for the excellent
  CSS parsing library, which made it possible to implement this with
  minimal custom parsing.
