# sub2api 文件引用化根治方案设计

## 背景

当前 `sub2api` 的 `/responses`、`/chat/completions` 等推理接口会直接接收包含图片 `base64 data URL` 的大请求体。真实部署中，即使应用层 `gateway.max_body_size` 默认较大，前置 `Nginx`、请求重试、长上下文叠加、图片编码膨胀等因素仍会导致：

- `413 Payload Too Large`
- 大请求占用更多代理层和应用层内存
- 文件上传失败与模型推理失败耦合
- 多实例部署时难以稳定扩展

本次设计的目标不是“继续增大 `/responses` 可承受的包体”，而是把“大文件传输”从“推理请求”里剥离出去。

## 目标

1. 让图片和附件不再作为推理请求主载荷进入 `/responses`
2. 为自家前端/面板提供成熟稳定的“先上传、后引用”主路径
3. 对第三方兼容客户端保留旧路径，但降级成兼容模式
4. 同时支持 `Docker Compose` 和 `systemd + 二进制安装` 的真实部署与验收
5. 补齐仓库文档、配置模板、测试和本地 Git 变更记录

## 非目标

1. 不追求让所有第三方 OpenAI 兼容客户端在完全不改行为的前提下永久发送超大 `base64`
2. 不在本次改造中引入复杂分片上传、多区域复制、CDN 优化
3. 不重构现有整个网关架构，只在现有 `handler -> service -> repository` 结构上新增能力

## 方案选择

### 方案 A：继续放大反代和应用层请求体限制

优点：

- 改动最小

缺点：

- 只是止血，不是根治
- 仍然让推理接口承担大文件传输
- 长期会继续遭遇大包、重试、内存和稳定性问题

### 方案 B：本地磁盘临时文件

优点：

- 依赖最少
- 上手快

缺点：

- 多实例和迁移不自然
- `Docker Compose` 与 `systemd` 下的目录、权限、挂载策略需要分别处理
- 长期扩展性差

### 方案 C：S3 兼容对象存储 + 预签名直传 + 推理引用

优点：

- 业界成熟路径
- 本地可用 `MinIO`，线上可切换 `S3/R2`
- 上传和推理解耦，适配多实例和长期演进

缺点：

- 需要新增文件元数据模型、上传接口和存储抽象

### 结论

采用 **方案 C**，默认优先支持 **MinIO**，通过 **S3 兼容接口**抽象统一接入。

## 总体架构

新增 3 个能力层：

1. **文件元数据层**
   - 负责记录文件状态、所属用户、大小、哈希、存储位置、过期时间
2. **对象存储层**
   - 负责生成预签名上传地址、读取文件引用、删除过期文件
3. **推理引用解析层**
   - 负责在 `/responses` 转发前把 `file_id` 解析成上游兼容输入

整体流程：

1. 客户端调用 `POST /v1/files` 申请上传
2. 服务端创建文件记录并返回 `file_id` 与预签名上传参数
3. 客户端直传对象存储
4. 客户端调用 `POST /v1/files/:id/complete`
5. 客户端在 `/responses` 中只传 `file_id`
6. 网关在发往上游前把 `file_id` 转换成 `image_url` 或上游可接受的输入结构

## 数据模型

新增 `files` 表，建议字段：

- `id`
- `owner_user_id`
- `api_key_id`
- `purpose`
- `storage_provider`
- `bucket`
- `object_key`
- `mime_type`
- `size_bytes`
- `sha256`
- `status`
- `expires_at`
- `created_at`
- `updated_at`

建议状态：

- `pending`
- `uploaded`
- `failed`
- `expired`

## API 设计

### 1. 创建上传会话

`POST /v1/files`

请求：

- 文件名
- MIME 类型
- 文件大小
- 用途

响应：

- `file_id`
- 预签名上传 URL
- 所需 headers 或 form fields
- 过期时间

### 2. 上传完成确认

`POST /v1/files/:id/complete`

职责：

- 校验对象是否存在
- 校验大小、MIME、哈希
- 将状态更新为 `uploaded`

### 3. 查询文件状态

`GET /v1/files/:id`

职责：

- 返回文件状态和可用于推理的引用信息

## 推理请求改造

### 新主路径

自家前端/客户端：

- 不再内联图片 `base64`
- 改为在 `/responses` 中传 `file_id`

### 兼容路径

第三方客户端继续发 `data:image/...;base64,...` 时：

- 小图兼容
- 大图拒绝
- 返回清晰错误，提示先上传文件再引用

### 处理位置

建议在现有网关转发链路里加入“文件引用解析”：

- `gateway_handler_responses.go`
- `gateway_forward_as_responses.go`
- 相关 OpenAI / Anthropic / Gemini 转换层

## 配置设计

新增配置段，例如：

```yaml
storage:
  backend: s3
  endpoint: http://minio:9000
  region: us-east-1
  bucket: sub2api-files
  access_key: xxx
  secret_key: xxx
  use_path_style: true
  presign_expire_seconds: 900
  max_file_size_bytes: 10485760
  allowed_mime_types:
    - image/png
    - image/jpeg
    - image/webp
```

同时要求：

- `nginx client_max_body_size >= gateway.max_body_size`
- 但新主路径下 `/responses` 实际包体将显著缩小

## 部署方案

### Docker Compose

新增 `MinIO` 服务：

- 与 `sub2api`、`PostgreSQL`、`Redis` 一起编排
- 通过 `.env` 和 `config.yaml` 注入存储配置

### systemd + 二进制

要求：

- 文档中给出 `MinIO` 外部部署方式
- `config.yaml` 中补齐 `storage` 配置
- systemd 服务仍保持 `sub2api` 单进程，文件存储由外部 `MinIO` 提供

## 测试策略

### 单元测试

- 文件服务
- 预签名参数生成
- 文件状态流转
- `file_id` 解析逻辑
- 超大 `base64` 判定逻辑

### 集成测试

- `POST /v1/files` -> 上传完成 -> `/responses`
- 文件不存在 / 未完成 / 已过期
- 兼容模式下大小阈值行为

### 部署验收

#### Docker Compose

- 启动 `sub2api + postgres + redis + minio`
- 从创建上传到推理完整跑通

#### systemd + 二进制

- 本地或测试机运行 `MinIO`
- 按 `config.yaml` 启动 `sub2api`
- 同样跑通完整链路

## 风险与回滚

风险：

1. 自家前端需要同步改造上传流程
2. 第三方兼容客户端短期仍可能发送大 `base64`
3. 新增对象存储配置后，部署复杂度略有上升

回滚策略：

1. 保留旧兼容路径
2. 新功能通过配置开关控制
3. 先灰度启用自家前端新路径

## 文档更新范围

需要同步更新：

- 部署文档
- Nginx 反代说明
- Docker Compose 配置说明
- 新增文件上传接口文档
- 资料夹中的中文总结与修改计划

## 验收标准

满足以下条件才视为可上线：

1. 本地 Git 变更清晰可追踪
2. 后端单元测试与集成测试通过
3. 前端相关测试通过
4. Docker Compose 真实编排链路跑通
5. systemd + 二进制配置链路跑通
6. 兼容模式下超大 `base64` 有明确错误提示
7. 文档与配置模板同步完成
