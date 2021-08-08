(function (factory) {
  if (typeof module === 'object' && typeof module.exports === 'object') {
    var v = factory(require, exports);
    if (v !== undefined) module.exports = v;
  } else if (typeof define === 'function' && define.amd) {
    define('cssbuild/cssbuild/testdata/expected_output.module.css', [
      'require',
      'exports',
    ], factory);
  }
})(function (require, exports) {
  'use strict';
  Object.defineProperty(exports, '__esModule', { value: true });
  exports.animationNames = exports.classNames = void 0;
  exports.classNames = {
    bar: 'bar__SUFFIX__',
    baz: 'baz__SUFFIX__',
    foo: 'foo__SUFFIX__',
    fooBar: 'foo-bar__SUFFIX__',
  };
  exports.animationNames = {
    foo: 'foo__SUFFIX__',
  };
  exports.default = exports.classNames;
});
