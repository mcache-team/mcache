# 实现计划：cache-service

## 概述

基于需求文档与技术设计文档，将 mcache 缓存服务的缺失功能逐步实现并集成。主要工作包括：修复内存存储并发安全问题、补全 TTL 过期机制、实现数据更新接口、完善 HTTP 路由错误映射、新增 Module 嵌入客户端，以及为各核心属性编写属性测试。

## 任务

- [x] 1. 修复内存存储层并发安全问题
  - 在 `pkg/storage/memory/memory.go` 的 `Memory` 结构体中添加 `sync.RWMutex` 字段
  - 对 `prefixList` 的所有读操作（`ListPrefix`、`CountPrefix`）加读锁，写操作（`Insert`、`Delete`）加写锁
  - 修复 `Insert` 中 `dataMap.Load` 与 `dataMap.Store` 之间的竞态：改用 `sync.Map.LoadOrStore` 保证原子性
  - _需求：9.1、9.3_

  - [ ]* 1.1 为并发写入安全性编写属性测试
    - **属性 10：并发写入安全性**
    - **验证：需求 9.1、9.3**
    - 在 `pkg/storage/memory/memory_test.go` 中，用 N 个 goroutine 并发插入同一 Prefix，断言恰好一次成功

- [x] 2. 实现 TTL 过期机制
  - 在 `pkg/apis/v1/item/types.go` 中新增 `WithTTL(d time.Duration) item.Option` 函数
  - 修改 `pkg/storage/memory/memory.go` 的 `GetOne`：若 `ExpireTime` 非零且早于当前时间，则调用 `Delete` 并返回 `NoDataError`
  - 修改 `Insert`：应用 `opt` 后若 `Timeout > 0` 则设置 `ExpireTime = CreatedAt + Timeout`
  - _需求：3.1、3.2、3.3_

  - [ ]* 2.1 为 TTL 过期行为编写属性测试
    - **属性 6：TTL 过期行为**
    - **验证：需求 3.1、3.2、3.3**
    - 在 `pkg/storage/memory/memory_test.go` 中，测试过期前可查询、过期后返回 `NoDataError`、无 TTL 永不过期

  - [ ]* 2.2 为插入时间戳正确性编写属性测试
    - **属性 3：插入时间戳正确性**
    - **验证：需求 1.3**
    - 在 `pkg/storage/memory/memory_test.go` 中，断言 `createdAt`/`updatedAt` 在操作前后时间范围内

- [x] 3. 实现数据更新接口
  - 修改 `pkg/storage/memory/memory.go` 的 `Update`：将 `data []byte` 改为 `data interface{}`，更新 `updatedAt` 时间戳，若未传入新 TTL 则保留原 `expireTime`
  - 同步修改 `pkg/apis/v1/storage/types.go` 中 `Storage` 接口的 `Update` 签名
  - 在 `pkg/services/rest/data.go` 的 `DataHandler` 中新增 `update` 方法，注册 `POST /v1/data/:prefix` 路由
  - _需求：5.1、5.2、5.3_

  - [ ]* 3.1 为更新后数据一致性编写属性测试
    - **属性 8：更新后数据一致性**
    - **验证：需求 5.1**
    - 在 `pkg/storage/memory/memory_test.go` 中，插入后更新，断言查询返回新值且 `updatedAt` 更晚

  - [ ]* 3.2 为更新保留原有 TTL 编写属性测试
    - **属性 7：更新保留原有 TTL**
    - **验证：需求 3.4、5.3**
    - 在 `pkg/storage/memory/memory_test.go` 中，带 TTL 插入后不传新 TTL 更新，断言 `expireTime` 不变

- [x] 4. 完善存储层基础读写行为
  - 修复 `pkg/storage/memory/memory.go` 的 `Insert`：检测到 `PrefixExisted` 时不写入 `prefixList`
  - 确认 `Delete` 在 Prefix 不存在时返回 `PrefixNotExisted`（已有，确认无误）
  - _需求：1.1、1.2、6.2_

  - [ ]* 4.1 为插入-查询往返一致性编写属性测试
    - **属性 1：插入-查询往返一致性**
    - **验证：需求 1.1、1.4、4.1**
    - 在 `pkg/storage/memory/memory_test.go` 中，随机 Prefix + 随机数据，插入后查询断言数据相同

  - [ ]* 4.2 为重复插入拒绝编写属性测试
    - **属性 2：重复插入拒绝**
    - **验证：需求 1.2**
    - 在 `pkg/storage/memory/memory_test.go` 中，同一 Prefix 插入两次，断言第二次返回 `PrefixExisted` 且数据不变

- [x] 5. 检查点 — 确保所有测试通过
  - 确保所有测试通过，如有疑问请向用户确认。

- [x] 6. 完善 PrefixTree Handler 与 HTTP 错误映射
  - 在 `pkg/handlers/data.go` 的 `InsertNode` 中：若 Prefix 为空则返回描述性错误；将 `storage.StorageClient.Insert` 的 `PrefixExisted` 错误向上透传
  - 在 `pkg/handlers/data.go` 的 `RemoveNode` 中：将 `PrefixNotExisted` 错误向上透传
  - 修改 `pkg/services/rest/data.go`：根据错误类型映射到正确 HTTP 状态码（`PrefixExisted` → 409，`NoDataError`/`PrefixNotExisted` → 404，其他 → 500）
  - 在 `pkg/services/response/response.go` 中新增 `ResponseConflict`（409）响应函数
  - _需求：7.2、7.3、7.4_

  - [ ]* 6.1 为 HTTP 接口各端点编写单元测试
    - 测试 PUT 创建（201）、重复创建（409）、GET 存在（200）、GET 不存在（404）、DELETE（200/404）、POST 更新（200/404）
    - 在 `pkg/services/rest/data_test.go` 中实现

- [x] 7. 完善 Prefix 解析与往返一致性
  - 检查 `pkg/apis/v1/prefix-tree/types.go` 的 `SplitPrefix`：确认前导 `/` 去除逻辑正确
  - 新增 `PrefixGroup.Rejoin() string` 方法，将 PrefixGroup 重新拼接为完整 Prefix 字符串
  - _需求：2.1、2.2、10.1、10.2、10.3、10.4_

  - [ ]* 7.1 为 Prefix 解析往返一致性编写属性测试
    - **属性 4：Prefix 解析往返一致性**
    - **验证：需求 2.5、10.1、10.2、10.3、10.4**
    - 在 `pkg/apis/v1/prefix-tree/types_test.go` 中，随机多级路径字符串，解析后拼接断言等价

- [x] 8. 完善前缀查询与 PrefixTree 节点操作
  - 修改 `pkg/handlers/data.go` 的 `ListNode`：当路径前缀下无子节点时返回空列表而非错误
  - 修改 `pkg/handlers/data.go` 的 `RemoveNode`：删除末端节点后将对应 `PrefixNode.HasData` 置为 `false`，保留节点结构
  - _需求：4.3、4.4、6.1、6.3_

  - [ ]* 8.1 为前缀查询完整性编写属性测试
    - **属性 5：前缀查询完整性**
    - **验证：需求 2.4、4.3**
    - 在 `pkg/handlers/data_test.go` 中，随机父路径下插入若干子节点，断言查询结果恰好包含所有子节点

  - [ ]* 8.2 为删除后不可查询编写属性测试
    - **属性 9：删除后不可查询**
    - **验证：需求 6.1、6.3**
    - 在 `pkg/handlers/data_test.go` 中，插入后删除，断言查询返回 `NoDataError` 且节点 `HasData` 为 `false`

  - [ ]* 8.3 为空路径前缀查询返回空列表编写属性测试
    - **属性 11：空路径前缀查询返回空列表**
    - **验证：需求 4.4**
    - 在 `pkg/handlers/data_test.go` 中，查询不存在子节点的路径，断言返回空列表

- [x] 9. 实现 Module 嵌入客户端
  - 新建 `pkg/client/cache.go`，定义 `CacheClient` 接口及其实现，包含 `Insert`、`Get`、`Update`、`Delete`、`ListByPrefix` 方法
  - 实现复用 `pkg/handlers` 与 `pkg/storage` 的现有逻辑，不启动 HTTP 监听
  - 支持通过 `item.Option`（如 `WithTTL`）传入配置参数
  - _需求：8.1、8.2、8.3、8.4_

- [x] 10. 检查点 — 确保所有测试通过
  - 确保所有测试通过，如有疑问请向用户确认。

- [x] 11. 集成收尾与路由注册验证
  - 确认 `POST /v1/data/:prefix` 已在 `DataHandler.RegisterRouter` 中注册
  - 确认 `GET /healthz` 正常响应（已有，确认无误）
  - 确认所有 `init()` 注册的 Controller 均已正确加载
  - _需求：7.1、7.5、7.6_

## 备注

- 标有 `*` 的子任务为可选项，可跳过以加快 MVP 交付
- 每个任务均引用具体需求条款以保证可追溯性
- 属性测试使用 [gopter](https://github.com/leanovate/gopter) 库，每个属性最少运行 100 次迭代
- 单元测试与属性测试互补：单元测试覆盖具体示例与边界条件，属性测试验证通用正确性
