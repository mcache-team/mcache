# 需求文档

## 简介

本文档描述 mcache 缓存服务的功能需求。mcache 是一个支持多级 key 路径的键值缓存服务，数据存储于 key 路径的末端节点。系统支持两种集成方式：作为 Go module 直接嵌入宿主程序内存，或作为独立 HTTP 服务对外提供 REST 接口。

## 词汇表

- **Cache_Service**：缓存服务整体，负责数据的存储、检索、更新与删除。
- **Storage**：底层存储引擎，当前实现为内存存储（Memory Storage）。
- **Prefix**：由 `/` 分隔的多级 key 路径，例如 `/user/profile/name`，数据存储于路径末端节点。
- **Item**：缓存条目，包含 Prefix、Data、过期时间、创建时间、更新时间等字段。
- **PrefixTree**：前缀树，用于管理多级 key 路径的层级关系。
- **TTL**：Time To Live，缓存条目的存活时长，超时后条目失效。
- **HTTP_Server**：对外提供 REST API 的 HTTP 服务，基于 Gin 框架实现。
- **Module_Client**：以 Go module 形式嵌入宿主程序的客户端接口。
- **REST_API**：通过 HTTP 协议对外暴露的增删读改接口。

---

## 需求

### 需求 1：键值数据存储

**用户故事：** 作为开发者，我希望以键值对形式存储任意数据，以便在后续请求中快速检索。

#### 验收标准

1. THE Cache_Service SHALL 以 Prefix 作为唯一标识存储 Item。
2. WHEN 插入请求携带相同 Prefix 时，THE Cache_Service SHALL 返回 `prefix already existed` 错误，拒绝重复插入。
3. WHEN 插入成功时，THE Storage SHALL 记录 Item 的 `createdAt` 与 `updatedAt` 时间戳为当前时间。
4. THE Cache_Service SHALL 支持存储任意 JSON 可序列化的数据类型作为 Item 的 Data 字段。

---

### 需求 2：多级 key 路径

**用户故事：** 作为开发者，我希望使用 `/` 分隔的多级路径作为 key，以便对数据进行层级化组织与管理。

#### 验收标准

1. THE Cache_Service SHALL 将 Prefix 按 `/` 分隔解析为多级路径，并在 PrefixTree 中维护层级关系。
2. WHEN Prefix 为空或格式非法时，THE Cache_Service SHALL 返回描述性错误信息，拒绝操作。
3. THE Cache_Service SHALL 仅在路径末端节点存储实际数据，中间节点仅作路径索引。
4. WHEN 查询某一路径前缀时，THE Cache_Service SHALL 返回该路径下所有直接子节点的 Item 列表。
5. FOR ALL 合法的多级 Prefix，THE PrefixTree SHALL 保证路径解析后再重新拼接与原始 Prefix 等价（往返一致性）。

---

### 需求 3：TTL 过期机制

**用户故事：** 作为开发者，我希望为缓存条目设置失效时间，以便数据在指定时间后自动失效，避免脏数据长期占用内存。

#### 验收标准

1. WHERE TTL 选项被设置，THE Cache_Service SHALL 在 Item 中记录 `expireTime` 为当前时间加上 TTL 时长。
2. WHEN 读取 Item 时，IF Item 的 `expireTime` 早于当前时间，THEN THE Cache_Service SHALL 返回 `no data` 错误，并将该 Item 从 Storage 中删除。
3. WHERE TTL 选项未设置，THE Cache_Service SHALL 将 Item 视为永不过期。
4. WHEN 更新 Item 时，WHERE 新的 TTL 选项被提供，THE Cache_Service SHALL 以新的 TTL 重新计算并覆盖 `expireTime`。

---

### 需求 4：数据读取

**用户故事：** 作为开发者，我希望通过精确 Prefix 或路径前缀查询缓存数据，以便灵活检索单条或批量数据。

#### 验收标准

1. WHEN 以精确 Prefix 查询时，THE Cache_Service SHALL 返回对应的单条 Item。
2. WHEN 以精确 Prefix 查询但数据不存在时，THE Cache_Service SHALL 返回 `no data` 错误。
3. WHEN 以路径前缀查询时，THE Cache_Service SHALL 返回该前缀路径下所有直接子节点的 Item 列表。
4. WHEN 路径前缀下无任何子节点数据时，THE Cache_Service SHALL 返回空列表而非错误。

---

### 需求 5：数据更新

**用户故事：** 作为开发者，我希望更新已存在的缓存条目的数据内容，以便修正或刷新缓存值。

#### 验收标准

1. WHEN 更新请求携带已存在的 Prefix 时，THE Cache_Service SHALL 将该 Item 的 Data 替换为新值，并更新 `updatedAt` 时间戳。
2. WHEN 更新请求携带不存在的 Prefix 时，THE Cache_Service SHALL 返回 `no data` 错误，拒绝更新。
3. WHERE 更新请求未携带新的 TTL 选项，THE Cache_Service SHALL 保留 Item 原有的 `expireTime` 不变。

---

### 需求 6：数据删除

**用户故事：** 作为开发者，我希望删除指定 Prefix 的缓存条目，以便主动清理不再需要的数据。

#### 验收标准

1. WHEN 删除请求携带已存在的 Prefix 时，THE Cache_Service SHALL 从 Storage 中移除该 Item，并在 PrefixTree 中将对应节点标记为无数据。
2. WHEN 删除请求携带不存在的 Prefix 时，THE Cache_Service SHALL 返回 `prefix not existed` 错误。
3. WHEN 删除末端节点后，IF 父路径节点下不再有任何含数据的子节点，THEN THE PrefixTree SHALL 保留路径结构，不自动删除中间节点。

---

### 需求 7：HTTP 服务模式

**用户故事：** 作为运维人员，我希望将缓存服务作为独立进程部署，并通过 HTTP REST 接口进行数据操作，以便与任意语言的客户端集成。

#### 验收标准

1. THE HTTP_Server SHALL 在 `0.0.0.0:8080` 监听并提供以下 REST 接口：
   - `PUT /v1/data` — 创建缓存条目
   - `GET /v1/data/:prefix` — 读取单条缓存条目
   - `POST /v1/data/:prefix` — 更新缓存条目
   - `DELETE /v1/data/:prefix` — 删除缓存条目
   - `GET /v1/data/listByPrefix?prefix=<path>` — 按路径前缀列举子节点
2. WHEN HTTP 请求体格式非法时，THE HTTP_Server SHALL 返回 HTTP 400 状态码及描述性错误信息。
3. WHEN 请求的资源不存在时，THE HTTP_Server SHALL 返回 HTTP 404 状态码。
4. WHEN 服务内部发生错误时，THE HTTP_Server SHALL 返回 HTTP 500 状态码及错误描述，且不暴露内部堆栈信息。
5. THE HTTP_Server SHALL 提供 `GET /healthz` 健康检查接口，正常时返回 HTTP 200。
6. WHEN HTTP_Server 启动失败时，THE HTTP_Server SHALL 记录错误日志并终止进程。

---

### 需求 8：Module 嵌入模式

**用户故事：** 作为 Go 开发者，我希望将缓存服务以 Go module 形式直接嵌入到我的程序中，以便在不启动独立服务的情况下使用缓存能力。

#### 验收标准

1. THE Module_Client SHALL 暴露与 HTTP 服务等价的增删读改及前缀查询接口，供宿主程序直接调用。
2. THE Module_Client SHALL 与 HTTP_Server 共享同一套 Storage 与 PrefixTree 实现，保证行为一致性。
3. WHERE Module_Client 被使用，THE Cache_Service SHALL 不启动任何 HTTP 监听端口。
4. THE Module_Client SHALL 支持通过函数选项（Option 模式）配置 TTL 等参数。

---

### 需求 9：内存存储引擎

**用户故事：** 作为开发者，我希望缓存数据默认存储于进程内存，以便获得最低延迟的读写性能。

#### 验收标准

1. THE Storage SHALL 使用 `sync.Map` 保证并发读写安全。
2. THE Storage SHALL 维护 prefixList 以支持按前缀快速枚举 key。
3. WHEN 并发写入相同 Prefix 时，THE Storage SHALL 保证只有一次写入成功，其余返回 `prefix already existed` 错误。
4. THE Storage SHALL 实现 `storage.Storage` 接口，以便未来替换为其他存储后端（如 Redis）而无需修改上层逻辑。

---

### 需求 10：Prefix 序列化与反序列化

**用户故事：** 作为开发者，我希望 Prefix 路径能够在字符串与结构化对象之间可靠转换，以便系统内部正确处理多级路径。

#### 验收标准

1. WHEN 解析 Prefix 字符串时，THE Cache_Service SHALL 将其拆分为 RootPrefix、中间路径段和 NodePrefix 组成的 PrefixGroup 结构。
2. THE Cache_Service SHALL 提供将 PrefixGroup 重新拼接为完整 Prefix 字符串的能力。
3. FOR ALL 合法的 Prefix 字符串，解析后再拼接 SHALL 产生与原始字符串等价的结果（往返一致性）。
4. WHEN Prefix 字符串以 `/` 开头时，THE Cache_Service SHALL 自动去除前导 `/` 后再进行解析，结果与不含前导 `/` 的等价路径一致。
