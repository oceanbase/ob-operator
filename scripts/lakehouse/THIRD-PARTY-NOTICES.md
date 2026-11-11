# Third-party components / 第三方组件说明

This configuration bundle optionally deploys MinIO as object storage and uses mc
for storage initialization. MinIO and mc are distributed under GNU AGPL v3.
They are separate third-party components and are not relicensed under this
project's Apache 2.0 license. This project is not affiliated with or endorsed by
MinIO, Inc. Review the applicable upstream terms for your use and redistribution.

本配置包可选部署 MinIO 对象存储，并使用 mc 初始化存储。MinIO 和 mc 按
GNU AGPL v3 分发，是独立的第三方组件，不因本项目采用 Apache 2.0 而改变许可。
本项目与 MinIO, Inc. 不存在关联或背书关系。请审阅适用于使用和再分发的上游条款。

The configuration archives contain no MinIO/mc binaries or image layers.
The offline tool exports selected images without rebuilding them; exact tags,
config IDs and available registry digests are recorded in the image inventory.
This does not assert that a particular cached image is an unmodified official build.

配置压缩包不包含 MinIO/mc 二进制或镜像层。离线工具导出选定镜像，不重建镜像，
并记录版本、config ID 及可取得的仓库 digest；这不表示任意缓存镜像都是原封不动的官方构建。

Upstream sources and license texts / 上游源码及许可文本：

- MinIO: https://github.com/minio/minio ; https://github.com/minio/minio/blob/master/LICENSE
- mc: https://github.com/minio/mc ; https://github.com/minio/mc/blob/master/LICENSE

Use the source revision corresponding to the selected release when preparing
redistribution materials. These links and this notice alone are not a complete
redistribution compliance assessment.
