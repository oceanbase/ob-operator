# LogService 详情导航回归（2026-09-22）

## 修复范围

此前 LogService 详情路由直接渲染 PageContainer，未经过 OB 集群等详情使用的 DetailLayout，因此缺少顶部栏和左侧导航。

- LogService 详情接入共用 DetailLayout，保留顶部栏、全局导航及资源子导航。
- 概览与拓扑、日志对象存储、LS 专项监控、启动参数、事件改为独立子路由；刷新或直接打开链接仍显示对应内容。
- 原 `/logservice/:ns/:name` 链接跳转至 `overview`，列表页、创建页、湖库关联入口继续可用。
- 共用详情布局增加受 LogService 读/写权限控制的全局入口；OB/租户/OBProxy 等详情可返回 LogService 列表。
- 未修改后端、Operator、OBD、数据库资源或原有扩缩容/删除权限。

## 自动化验证

```bash
cd ui
node --test tests/logservice-detail-layout.test.cjs
node --test tests/*.test.cjs
npm run build
```

新增测试 14/14、全体前端回归 39/39、生产构建通过。路由缺陷在修复前复现为 6 项失败；修复后通过。测试使用真实 React Router 匹配及组件服务端渲染，替换了网络/路由 hooks 和第三方布局容器，不将其宣称为已登录浏览器 E2E。

全项目 TypeScript 检查仍有既有错误，不能据此宣称全量类型检查通过。测试交付时应独立部署本分支前端再做下面的浏览器复测；本次代码提交本身不升级任何现有 Dashboard。

## 测试同学复测步骤

1. 正常登录，从 LogService 列表打开测试 LS：应有顶部栏、左侧全局导航、LS 名称/命名空间及五个子菜单。
2. 依次点击五个子菜单：内容与地址一致，菜单正确选中；刷新、浏览器前进/后退后仍保留当前页。
3. 直接打开旧详情链接及 `overview`、`storage`、`monitor`、`parameters`、`events` 深链接，验证导航不消失、不跳到别的资源。
4. 从湖库配置进入关联 LS，再通过左侧全局导航切回集群/LogService；验证目标资源正确。
5. 使用只读账号检查扩缩容/删除入口仍禁用；没有 LS 权限的账号不应看到全局 LS 入口。不要执行真实删除或扩缩容来验证布局。
6. 回归 OB 集群、租户和 OBProxy 详情：原子菜单/选中态不受影响。LogService 新建仍是原创建页面。

工单应停在 Fixed / 待验证，待测试确认后再关闭；构建通过不等于部署或页面验收通过。
