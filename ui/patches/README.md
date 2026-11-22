# OceanBase UI style callback patch

`@oceanbase/ui` is pinned to 0.2.13 while this patch is required. Its style factory read `ConfigProvider.ConfigContext` inside the callback passed to `useStyle`. CSS-in-JS can invoke that callback again from an insertion effect when rebuilding a discarded cache. Calling `useContext` there causes React #321.

The ES and CommonJS patches read context in the returned hook's render instead, then capture the values for the style callback. Installation and build both apply the patch with `--error-on-fail`; a dependency upgrade must explicitly review or remove it, rather than silently building without the fix.

Run `node --test tests/style-hook-render-scope.test.cjs` on a pooled test machine after installing dependencies. It rebuilds the style cache after render and checks that separate renders retain their own prefix and icon context. Browser coverage must also exercise topology → cluster overview → global overview, including first-login password reset.
