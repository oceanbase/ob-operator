# Third-party components / 第三方组件说明

OceanSeek can use MinIO for object storage and MinIO Client (mc) to initialize storage.
This repository uses the Mulan Permissive Software License, Version 2. MinIO and mc
are licensed under the GNU Affero General Public License, version 3 (AGPLv3). [Repository license](https://github.com/oceanbase/ob-operator/blob/master/LICENSE)

OceanSeek 可使用 MinIO 存储对象，使用 MinIO Client（mc）初始化存储。本仓库采用木兰宽松许可证第 2 版。
MinIO 和 mc 沿用各自的 GNU Affero General Public License 第 3 版（AGPLv3）许可条款。

The Compose and Helm configuration archives contain deployment files. The offline
exporter can separately copy the actual image layers. For redistribution of those
images, retain their license texts, copyright notices, corresponding source and
build materials, including any modifications for the exact selected version.
The image lock records the selected references and content identities.

Compose 和 Helm 配置包包含部署文件。离线导出工具可另行复制镜像层。再分发镜像时，请保留所选版本的许可文本、
版权声明、对应源码和构建材料，并包含该版本使用的修改。镜像清单记录所选引用和内容摘要。

| Component / 组件 | Source and license for the indicated upstream release / 对应上游版本源码及许可证 |
| --- | --- |
| MinIO RELEASE.2025-10-15T17-29-55Z | [Source](https://github.com/minio/minio/tree/9e49d5e7a648f00e26f2246f4dc28e6b07f8c84a) · [License](https://github.com/minio/minio/blob/9e49d5e7a648f00e26f2246f4dc28e6b07f8c84a/LICENSE) |
| mc RELEASE.2025-08-13T08-35-41Z | [Source](https://github.com/minio/mc/tree/7394ce0dd2a80935aded936b09fa12cbb3cb8096) · [License](https://github.com/minio/mc/blob/7394ce0dd2a80935aded936b09fa12cbb3cb8096/LICENSE) |

For the OceanBase-packaged images of these releases, consult the notices supplied
by the image distributor. The November 2026 packages include the distributor’s
bilingual notice at `/usr/share/doc/oceanbase-minio/NOTICE.txt` in both images. Other tags
and custom builds require the source and build materials corresponding to those images.

上述 OceanBase 封装镜像的具体声明请见分发方随附的文件。2026 年 11 月的两个镜像均在
`/usr/share/doc/oceanbase-minio/NOTICE.txt` 中提供分发方的双语声明。
其他版本及定制构建须使用与其镜像内容相对应的源码和构建材料。
