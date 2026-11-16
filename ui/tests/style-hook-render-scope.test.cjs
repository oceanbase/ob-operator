const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');
const vm = require('node:vm');

// CSS-in-JS may rebuild a discarded cache during an insertion effect. That
// callback is outside React render, so it must never read React context itself.
for (const format of ['es', 'lib']) {
  test(`${format}: style cache reconstruction does not invoke React hooks`, () => {
    const filename = process.env.OB_UI_STYLE_FIXTURE_DIR
      ? path.join(process.env.OB_UI_STYLE_FIXTURE_DIR, format + '.js')
      : require.resolve(
          `@oceanbase/ui/${format}/_util/genComponentStyleHook.js`
        );
    let source = fs.readFileSync(filename, 'utf8');
    let rendering = false;
    let hookCalls = 0;
    let context = { getPrefixCls: () => 'first', iconPrefixCls: 'first-icon' };
    const callbacks = [];
    const React = {
      useContext() {
        if (!rendering) throw new Error('Invalid hook call outside render');
        hookCalls++;
        return context;
      },
    };
    const ConfigProvider = { ConfigContext: {} };
    const useStyle = (name, generate) => {
      assert.equal(name, 'TestComponent');
      callbacks.push(generate);
      return { wrapSSR: (node) => node, hashId: 'test' };
    };
    const module = { exports: {} };
    const sandbox = {
      React,
      ConfigProvider,
      useStyle,
      module,
      exports: module.exports,
      require(name) {
        if (name === 'react') return React;
        if (name === '@oceanbase/design') return { ConfigProvider };
        if (name === '@ant-design/pro-components') return { useStyle };
        throw new Error(`Unexpected dependency: ${name}`);
      },
    };
    if (format === 'es') {
      source = source
        .replace(/^import .*;\r?\n/gm, '')
        .replace('export function', 'function');
    }
    vm.runInNewContext(source, sandbox, { filename });
    const factory =
      format === 'es'
        ? sandbox.genComponentStyleHook
        : module.exports.genComponentStyleHook;
    const useComponentStyle = factory('TestComponent', (token) => token);
    for (const prefix of ['first', 'second']) {
      context = { getPrefixCls: () => prefix, iconPrefixCls: prefix + '-icon' };
      rendering = true;
      useComponentStyle(prefix + '-component');
      rendering = false;
    }
    for (let i = 0; i < callbacks.length; i++) {
      const prefix = i === 0 ? 'first' : 'second';
      // Rebuild multiple times after render, as cache lifetimes change on routes.
      for (let repeat = 0; repeat < 2; repeat++) {
        let result;
        assert.doesNotThrow(() => {
          result = callbacks[i]({ color: 'red' })[0];
        });
        assert.deepEqual(JSON.parse(JSON.stringify(result)), {
          color: 'red',
          componentCls: '.' + prefix + '-component',
          prefixCls: prefix + '-component',
          iconCls: '.' + prefix + '-icon',
          antCls: '.' + prefix,
        });
      }
    }
    assert.equal(hookCalls, 2, 'exactly one context read per hook render');
  });
}
