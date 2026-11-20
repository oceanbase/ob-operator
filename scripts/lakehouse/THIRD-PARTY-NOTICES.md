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

Internal r3 source build / 内部 r3 源码构建：

MinIO Notice

This project optionally uses MinIO as an object storage backend and MinIO Client (mc) as a supporting tool. MinIO and mc are distributed under the GNU AGPL v3 (version 3 or later, as stated in their source headers).

For community convenience, this project builds and packages container images from the upstream open-source MinIO and mc source code. The upstream program source is unmodified. These are independently built OceanBase images, not official release artifacts or images published by MinIO, Inc., and they are not MinIO AIStor.

- This project is not affiliated with or endorsed by MinIO, Inc.
- Users are solely responsible for reviewing and complying with the MinIO and mc license terms applicable to their use case. The AGPL v3 may impose obligations that differ from this project's Apache 2.0 license.
- This project and its contributors make no warranty or representation regarding MinIO license compliance. Users should consult their own legal counsel if they have concerns.
- If the AGPL v3 terms are a concern, consider alternative storage backends.

Full upstream license texts:
MinIO: https://github.com/minio/minio/blob/9e49d5e7a648f00e26f2246f4dc28e6b07f8c84a/LICENSE
mc: https://github.com/minio/mc/blob/7394ce0dd2a80935aded936b09fa12cbb3cb8096/LICENSE
Copies of LICENSE and CREDITS are included in /licenses/minio (server image) and /licenses/mc (both images).

MinIO 说明

本项目可选使用 MinIO 作为对象存储后端，并使用 MinIO Client（mc）作为配套工具。MinIO 和 mc 以 GNU AGPL v3 许可协议分发（根据其源文件声明，为第 3 版或更新版本）。

为方便社区用户，本项目基于 MinIO 和 mc 的上游开源源码自行构建并封装容器镜像，未修改上游程序源码。这些镜像是 OceanBase 独立构建的产物，并非 MinIO, Inc. 官方发布的产物或镜像，也不是 MinIO AIStor。

- 本项目与 MinIO, Inc. 不存在关联或背书关系。
- 用户应自行审阅并遵守 MinIO 和 mc 许可协议中适用于其使用场景的条款。AGPL v3 的义务可能与本项目的 Apache 2.0 许可协议有所不同。
- 本项目及其贡献者不对 MinIO 的许可合规性作任何担保或声明。如有疑虑，用户应咨询其法律顾问。
- 如 AGPL v3 条款对您构成顾虑，可考虑替代存储方案。

上游完整许可协议文本：
MinIO：https://github.com/minio/minio/blob/9e49d5e7a648f00e26f2246f4dc28e6b07f8c84a/LICENSE
mc：https://github.com/minio/mc/blob/7394ce0dd2a80935aded936b09fa12cbb3cb8096/LICENSE
镜像内同时提供 LICENSE 和 CREDITS：/licenses/minio（服务端镜像）及 /licenses/mc（两个镜像均有）。

Build and redistribution details / 构建与再分发说明

Container packaging revision: r3 / 容器打包修订版本：r3
MinIO source tag: RELEASE.2025-10-15T17-29-55Z
MinIO source commit: 9e49d5e7a648f00e26f2246f4dc28e6b07f8c84a
mc source tag: RELEASE.2025-08-13T08-35-41Z
mc source commit: 7394ce0dd2a80935aded936b09fa12cbb3cb8096

Packaging uses Debian bookworm, curl and CA certificates. The binaries were built with Go 1.24.8 and are marked DEVELOPMENT.*.sourcebuild. Revision r3 updates documentation and image metadata on the digest-pinned r2 images; the program binaries, runtime dependencies and upstream license files are unchanged.
本镜像采用 Debian bookworm，包含 curl 和 CA 证书。程序使用 Go 1.24.8 构建，版本标记为 DEVELOPMENT.*.sourcebuild。r3 在按摘要固定的 r2 镜像基础上更新文档和镜像元数据；程序二进制、运行时依赖及上游许可证文件保持不变。

Source and build materials are described in /usr/share/doc/oceanbase-minio/SOURCE-CODE.txt. Additional Go dependency notices are in /licenses/distribution/go-dependencies. Debian component notices remain in /usr/share/doc and /licenses/runtime-doc; common licenses are in /licenses/common.
源码及构建资料见 /usr/share/doc/oceanbase-minio/SOURCE-CODE.txt。Go 依赖声明位于 /licenses/distribution/go-dependencies；Debian 组件声明位于 /usr/share/doc 和 /licenses/runtime-doc，通用许可证位于 /licenses/common。

Third-party components retain their own licenses. The AGPL image label describes the main application; it does not relicense every image component. Preserve all applicable notices when redistributing. This notice does not waive the distributor's obligations or add restrictions on recipients' rights under the applicable open-source licenses. Copyright remains with the respective authors; software and packaging are provided without warranty, subject to their applicable license terms.
第三方组件保留各自的许可协议。镜像的 AGPL 标签描述的是主程序，不会改变镜像内其他组件的许可协议。再分发时应保留适用的声明。本说明不免除分发者应承担的义务，也不对接收者依据适用开源许可享有的权利附加限制。版权归各自作者所有；软件及封装按其适用许可条款提供，不附带担保。

Source and build materials / 源码和构建资料：
Obtain the matching SOURCE-CODE.txt and source archive from your image distributor.
