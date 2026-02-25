## 调度服务（scheduling）库存校验/扣减逻辑检查结论

### 结论（当前实现）

- 创建运输任务（CreateShipment）时：未通过 RPC 调用仓储/库存服务进行“库存是否足够”的校验，也未做“预占/扣减库存”。
- 更新运输状态（UpdateShipmentStatus）时：仅更新本服务数据库中的 `shipments.status` 并追加一条 `shipment_tracking` 记录，不会扣减任何库存。

### 证据（代码位置）

**1）创建运输任务**

- 业务层：CreateShipment 仅写入 Shipment + Tracking，并把 Request 状态置为 ASSIGNED  
  - [scheduling_service.go](file:///root/ems/backend/internal/scheduling/service/scheduling_service.go#L99-L143)
    - 写入 shipment：调用 `repo.CreateShipment(...)`
    - 写入初始 tracking：`Location: "Warehouse"` 只是字符串标记，不是库存动作
    - 更新 request 状态：`req.Status = ASSIGNED` 并 `repo.UpdateRequest(...)`
- 持久层：CreateShipment 仅 `db.Create(shipment)`  
  - [scheduling_repository.go](file:///root/ems/backend/internal/scheduling/repository/scheduling_repository.go#L83-L86)

**2）更新运输状态**

- 业务层：UpdateShipmentStatus 仅调用 repo 更新状态并返回最新 Shipment  
  - [scheduling_service.go](file:///root/ems/backend/internal/scheduling/service/scheduling_service.go#L150-L156)
- 持久层：事务内仅更新 `shipments.status` + 插入 `shipment_tracking`  
  - [scheduling_repository.go](file:///root/ems/backend/internal/scheduling/repository/scheduling_repository.go#L97-L112)

**3）目录内未发现库存/RPC 调用痕迹**

在 `backend/internal/scheduling` 目录中未发现与 `warehouse/inventory/stock` 相关的 gRPC client 初始化或调用点（例如 `NewWarehouseServiceClient`、`Reserve/Deduct` 等语义），因此调度服务当前对库存没有任何跨服务校验/扣减逻辑。

### 风险与可能 Bug 表现

- 多个运输任务可从同一仓库针对同一物资“超额下单”，不会被拦截。
- 运输任务从 NEW→IN_TRANSIT→DELIVERED 的过程中，即使业务语义上“应扣减库存”，系统也不会产生库存变化，导致仓储库存与实际发放不一致。

### 建议（如何补齐）

建议在“创建运输任务”时引入仓储服务（warehouse）库存校验与扣减，并保证幂等：

- **校验时机**：CreateShipment（创建时）做“是否足够”；如需要更严格一致性，可在发货（状态切到 IN_TRANSIT）时做最终扣减。
- **接口设计**：
  - 预占/释放/确认扣减三段式（推荐）：`ReserveStock` / `ReleaseReservation` / `CommitReservation`
  - 或者简化为 `CheckAndDeduct`（实现容易，但回滚/超时补偿复杂）
- **幂等**：以 `shipment_id`（或 reservation_id）作为幂等键，避免重试导致重复扣减。
- **一致性策略**：跨服务事务建议用 Saga（补偿）或 Outbox；不要依赖分布式事务。
