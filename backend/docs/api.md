# 基于Golang的救援物资管理系统 API 文档

以下文档采用Swagger/OpenAPI风格编写，按模块组织接口信息，包含URL路径、HTTP方法、参数说明、请求/响应示例及权限说明，方便前后端协作与系统联调。所有接口均使用JWT令牌进行身份验证[1]。

## 认证与多角色权限模块

说明：本系统采用基于角色的访问控制（RBAC），用户拥有如 `Admin`、`WarehouseManager`、`Dispatcher` 角色。所有接口均需在HTTP头中携带有效的JWT令牌进行认证[1]。不同接口会根据用户角色限制访问权限。

### 用户登录 (Login)

- **URL Path**：`POST /api/v1/auth/login`
- **描述**：用户使用用户名和密码登录，成功后返回JWT令牌。
- **请求参数**（JSON Body）：
  - `username` (string, 必填)：用户名
  - `password` (string, 必填)：密码
- **请求示例**（JSON）：
  ```json
  {
    "username": "admin",
    "password": "admin"
  }
  ```
- **响应**：
  - **成功**：HTTP 200，返回JSON，包含token（JWT令牌）和用户信息。
    ```json
    {
      "token": "eyJhbGciOiJI...（JWT）",
      "user": {
        "id": 1,
        "username": "admin",
        "roles": ["Admin"]
      }
    }
    ```
  - **失败**：HTTP 401（Unauthorized），返回错误信息。
- **权限**：公开（任何用户均可访问，无需预先授权）。

### 用户登出 (Logout)

- **URL Path**：`POST /api/v1/auth/logout`
- **描述**：用户登出接口，可用于客户端丢弃JWT。
- **请求头**：`Authorization: Bearer <JWT>`
- **请求参数**：无
- **响应**：
  - **成功**：HTTP 200，`{ "message": "Logged out" }`。
  - **失败**：HTTP 401（Unauthorized）或400（Bad Request）。
- **权限**：已登录用户。

## 用户管理接口 (User Management)

- **URL Path**：`GET /api/v1/users`
- **描述**：查询所有用户列表，支持分页（可选）。
- **请求参数**（Query）：
  - `page` (integer, 可选)：页码，默认1
  - `size` (integer, 可选)：每页数量，默认20
- **响应**：HTTP 200，返回用户数组及分页信息：
  ```json
  {
    "page": 1,
    "size": 20,
    "total": 5,
    "users": [
      { "id": 1, "username": "admin", "roles": ["Admin"] },
      { "id": 2, "username": "manager", "roles": ["WarehouseManager"] }
      // ...
    ]
  }
  ```
- **权限**：仅限 `Admin`。

- **URL Path**：`POST /api/v1/users`
- **描述**：创建新用户，设置用户名、密码、角色等信息。
- **请求参数**（JSON Body）：
  - `username` (string, 必填)
  - `password` (string, 必填)
  - `roles` (array[string], 可选) - 角色列表，如 `["WarehouseManager"]`
- **请求示例**：
  ```json
  {
    "username": "newuser",
    "password": "pass123",
    "roles": ["Dispatcher"]
  }
  ```
- **响应**：HTTP 201，返回创建的用户信息：
  ```json
  {
    "id": 6,
    "username": "newuser",
    "roles": ["Dispatcher"]
  }
  ```
- **权限**：仅限 `Admin`。

- **URL Path**：`GET /api/v1/users/{id}`
- **描述**：获取指定用户详细信息。
- **路径参数**：
  - `id` (integer, 必填)：用户ID
- **响应**：HTTP 200，返回用户对象：
  ```json
  {
    "id": 2,
    "username": "manager",
    "roles": ["WarehouseManager"]
  }
  ```
- **权限**：仅限 `Admin`。

- **URL Path**：`PUT /api/v1/users/{id}`
- **描述**：更新用户信息（用户名、密码或角色）。
- **路径参数**：同上
- **请求参数**（JSON Body，可选提供以下字段）：
  - `password` (string)：新密码
  - `roles` (array[string])：新角色列表
- **请求示例**：
  ```json
  {
    "password": "newpass",
    "roles": ["Admin"]
  }
  ```
- **响应**：HTTP 200，返回更新后的用户信息。
- **权限**：仅限 `Admin`。

- **URL Path**：`DELETE /api/v1/users/{id}`
- **描述**：删除指定用户。
- **路径参数**：同上
- **响应**：HTTP 204（No Content）。
- **权限**：仅限 `Admin`。

## 角色与权限分配 (Role Assignment)

- **URL Path**：`GET /api/v1/roles`
- **描述**：列出所有系统角色及其描述。
- **响应**：HTTP 200，返回角色列表：
  ```json
  [
    { "name": "Admin", "description": "系统管理员, 拥有所有权限" },
    { "name": "WarehouseManager", "description": "仓库管理员, 管理物资出入库，查看报表和统计数据" },
    { "name": "Dispatcher", "description": "救援人员, 负责物资调度的申请" }
  ]
  ```
- **权限**：`Admin` 可见（或公开均可）。

- **URL Path**：`PUT /api/v1/users/{id}/roles`
- **描述**：更新用户角色。
- **路径参数**：同上
- **请求参数**（JSON Body）：
  - `roles` (array[string], 必填)：角色列表
- **请求示例**：
  ```json
  {
    "roles": ["WarehouseManager","Dispatcher"]
  }
  ```
- **响应**：HTTP 200，返回用户新角色信息。
- **权限**：仅限 `Admin`。

## 物资库存与仓储管理模块

说明：负责物资种类管理、库存增减、仓库管理等功能。仓库管理员（`WarehouseManager`）和管理员（`Admin`）可执行库存相关操作。

### 查询物资列表 (List Items)

- **URL Path**：`GET /api/v1/items`
- **描述**：查询所有物资种类列表，支持分页和模糊搜索。
- **请求参数**（Query，可选）：
  - `page` (integer): 页码，默认1
  - `size` (integer): 每页数量，默认20
  - `query` (string): 搜索关键字（按名称搜索）
- **响应**：HTTP 200，返回物资数组和分页信息，例如：
  ```json
  {
    "page": 1,
    "size": 20,
    "total": 50,
    "items": [
      { "id": 101, "name": "帐篷", "unit": "顶", "description": "双人帐篷" },
      { "id": 102, "name": "睡袋", "unit": "个", "description": "保暖睡袋" }
      // ...
    ]
  }
  ```
- **权限**：`WarehouseManager`、`Admin`。

### 创建物资 (Create Item)

- **URL Path**：`POST /api/v1/items`
- **描述**：添加新的物资种类。
- **请求参数**（JSON Body）：
  - `name` (string, 必填)：物资名称
  - `unit` (string, 必填)：计量单位（如件、箱等）
  - `description` (string, 可选)：描述
- **请求示例**：
  ```json
  {
    "name": "食物包",
    "unit": "箱",
    "description": "应急食品包"
  }
  ```
- **响应**：HTTP 201，返回创建的物资信息：
  ```json
  {
    "id": 103,
    "name": "食物包",
    "unit": "箱",
    "description": "应急食品包"
  }
  ```
- **权限**：`WarehouseManager`、`Admin`。

### 获取物资详情 (Get Item)

- **URL Path**：`GET /api/v1/items/{itemId}`
- **描述**：获取指定物资的详细信息。
- **路径参数**：
  - `itemId` (integer, 必填)：物资ID
- **响应**：HTTP 200，返回物资对象：
  ```json
  {
    "id": 101,
    "name": "帐篷",
    "unit": "顶",
    "description": "双人帐篷"
  }
  ```
- **权限**：`WarehouseManager`、`Admin`。

### 更新物资 (Update Item)

- **URL Path**：`PUT /api/v1/items/{itemId}`
- **描述**：更新指定物资的信息。
- **路径参数**：同上
- **请求参数**（JSON Body，可选字段）：
  - `name` (string)
  - `unit` (string)
  - `description` (string)
- **请求示例**：
  ```json
  {
    "unit": "顶",
    "description": "加厚型双人帐篷"
  }
  ```
- **响应**：HTTP 200，返回更新后的物资信息。
- **权限**：`WarehouseManager`、`Admin`。

### 删除物资 (Delete Item)

- **URL Path**：`DELETE /api/v1/items/{itemId}`
- **描述**：删除指定物资种类。
- **路径参数**：同上
- **响应**：HTTP 204（No Content）。
- **权限**：`Admin`（谨慎操作）。

## 仓库管理接口 (Warehouses)

- **URL Path**：`GET /api/v1/warehouses`
- **描述**：查询所有仓库信息列表。
- **响应**：HTTP 200，返回仓库数组：
  ```json
  [
    { "id": 1, "name": "主仓库", "location": "北京市" },
    { "id": 2, "name": "西部仓库", "location": "陕西省" }
  ]
  ```
- **权限**：`WarehouseManager`、`Admin`。

- **URL Path**：`POST /api/v1/warehouses`
- **描述**：添加新仓库。
- **请求参数**（JSON Body）：
  - `name` (string, 必填)：仓库名称
  - `location` (string, 可选)：仓库地址
- **请求示例**：
  ```json
  {
    "name": "东部仓库",
    "location": "上海市"
  }
  ```
- **响应**：HTTP 201，返回创建的仓库信息。
- **权限**：`Admin`。

- **URL Path**：`GET /api/v1/warehouses/{id}`
- **描述**：获取指定仓库详情。
- **路径参数**：
  - `id` (integer, 必填)：仓库ID
- **响应**：HTTP 200，返回仓库对象：
  ```json
  {
    "id": 1,
    "name": "主仓库",
    "location": "北京市"
  }
  ```
- **权限**：`WarehouseManager`、`Admin`。

- **URL Path**：`PUT /api/v1/warehouses/{id}`
- **描述**：更新仓库信息。
- **路径参数**：同上
- **请求参数**（JSON Body，可选字段）：
  - `name` (string)
  - `location` (string)
- **请求示例**：
  ```json
  {
    "location": "北京市朝阳区"
  }
  ```
- **响应**：HTTP 200，返回更新后的仓库信息。
- **权限**：`Admin`。

- **URL Path**：`DELETE /api/v1/warehouses/{id}`
- **描述**：删除指定仓库。
- **路径参数**：同上
- **响应**：HTTP 204。
- **权限**：`Admin`。

## 库存调整 (Inventory Adjustment)

- **URL Path**：`GET /api/v1/warehouses/{id}/inventory`
- **描述**：获取指定仓库的库存列表（各物资数量）。
- **路径参数**：同上
- **响应**：HTTP 200，返回库存详情数组：
  ```json
  [
    { "itemId": 101, "name": "帐篷", "quantity": 20 },
    { "itemId": 102, "name": "睡袋", "quantity": 50 }
  ]
  ```
- **权限**：`WarehouseManager`、`Admin`。

- **URL Path**：`POST /api/v1/warehouses/{id}/inventory/add`
- **描述**：入库增加库存。
- **路径参数**：同上
- **请求参数**（JSON Body）：
  - `itemId` (integer, 必填)：物资ID
  - `amount` (integer, 必填)：增加数量
- **请求示例**：
  ```json
  {
    "itemId": 101,
    "amount": 10
  }
  ```
- **响应**：HTTP 200，返回更新后的该物资库存数：
  ```json
  { "itemId": 101, "quantity": 30 }
  ```
- **权限**：`WarehouseManager`、`Admin`。

- **URL Path**：`POST /api/v1/warehouses/{id}/inventory/remove`
- **描述**：出库减少库存（只有在可满足时）。
- **路径参数**：同上
- **请求参数**（JSON Body）：
  - `itemId` (integer, 必填)
  - `amount` (integer, 必填)：减少数量
- **请求示例**：
  ```json
  {
    "itemId": 102,
    "amount": 5
  }
  ```
- **响应**：
  - **成功**：HTTP 200，返回更新后的库存数。
  - **库存不足**：HTTP 400，返回错误信息。
- **权限**：`WarehouseManager`、`Admin`。

## 智能需求调度模块

说明：管理救援需求（需求单）并进行资源调度分配。调度员（`Dispatcher`）或管理员可操作相关接口。

### 创建需求单 (Create Request)

- **URL Path**：`POST /api/v1/requests`
- **描述**：前端或管理员创建新的救援物资需求单。
- **请求参数**（JSON Body）：
  - `title` (string, 必填)：需求标题
  - `location` (string, 必填)：需求地点
  - `items` (array, 必填)：所需物资列表，格式例如：
    ```json
    [
      { "itemId": 101, "quantity": 2 },
      { "itemId": 103, "quantity": 1 }
    ]
    ```
- **请求示例**：
  ```json
  {
    "title": "地震救援物资需求",
    "location": "成都",
    "items": [
      { "itemId": 101, "quantity": 5 },
      { "itemId": 102, "quantity": 10 }
    ]
  }
  ```
- **响应**：HTTP 201，返回创建的需求单信息：
  ```json
  {
    "id": 1001,
    "title": "地震救援物资需求",
    "location": "成都",
    "items": [
      { "itemId": 101, "quantity": 5 },
      { "itemId": 102, "quantity": 10 }
    ],
    "status": "PENDING"  // PENDING, ASSIGNED, COMPLETED 等
  }
  ```
- **权限**：`Dispatcher`、`Admin`。

### 查询需求单 (List Requests)

- **URL Path**：`GET /api/v1/requests`
- **描述**：查询需求单列表，支持筛选和分页。
- **请求参数**（Query，可选）：
  - `page`, `size`：分页
  - `status` (string)：按状态过滤，如 `PENDING`、`ASSIGNED`
- **响应**：HTTP 200，返回需求单列表，例如：
  ```json
  {
    "page": 1,
    "size": 10,
    "total": 3,
    "requests": [
      { "id": 1001, "title": "地震救援物资需求", "location": "成都", "status": "PENDING" },
      { "id": 1002, "title": "洪涝救援需求", "location": "武汉", "status": "ASSIGNED" }
      // ...
    ]
  }
  ```
- **权限**：`Dispatcher`、`Admin`。

### 需求单详情 (Get Request)

- **URL Path**：`GET /api/v1/requests/{id}`
- **描述**：获取指定需求单详情，包括物资列表和当前状态。
- **路径参数**：
  - `id` (integer, 必填)：需求单ID
- **响应**：HTTP 200，返回需求单对象：
  ```json
  {
    "id": 1001,
    "title": "地震救援物资需求",
    "location": "成都",
    "items": [
      { "itemId": 101, "quantity": 5 },
      { "itemId": 102, "quantity": 10 }
    ],
    "status": "PENDING",
    "assignedTo": null
  }
  ```
- **权限**：`Dispatcher`、`Admin`。

### 更新需求单 (Update Request)

- **URL Path**：`PUT /api/v1/requests/{id}`
- **描述**：更新需求单信息或状态（例如指派调度员）。
- **路径参数**：同上
- **请求参数**（JSON Body，可选字段）：
  - `status` (string)：更新状态，如 `ASSIGNED` 或 `COMPLETED`
  - `assignedTo` (integer)：指派的用户ID（调度员）
- **请求示例**：
  ```json
  {
    "status": "ASSIGNED",
    "assignedTo": 2
  }
  ```
- **响应**：HTTP 200，返回更新后的需求单信息。
- **权限**：`Dispatcher`、`Admin`。

### 删除需求单 (Delete Request)

- **URL Path**：`DELETE /api/v1/requests/{id}`
- **描述**：删除指定需求单（仅当未指派时可删除）。
- **路径参数**：同上
- **响应**：HTTP 204。
- **权限**：`Admin`。

## 物流轨迹追踪模块

说明：管理运输过程中的物流信息和轨迹。可创建运输任务、更新状态并跟踪位置。

### 创建运输任务 (Create Shipment)

- **URL Path**：`POST /api/v1/shipments`
- **描述**：为已指派的需求单创建运输任务，记录发货仓库和目的地。
- **请求参数**（JSON Body）：
  - `requestId` (integer, 必填)：对应的需求单ID
  - `fromWarehouseId` (integer, 必填)：出发仓库ID
  - `toLocation` (string, 必填)：目的地
  - `items` (array, 必填)：运输物资列表，与需求单物资一致。
- **请求示例**：
  ```json
  {
    "requestId": 1001,
    "fromWarehouseId": 1,
    "toLocation": "成都",
    "items": [
      { "itemId": 101, "quantity": 5 },
      { "itemId": 102, "quantity": 10 }
    ]
  }
  ```
- **响应**：HTTP 201，返回运输任务信息：
  ```json
  {
    "shipmentId": 5001,
    "requestId": 1001,
    "fromWarehouseId": 1,
    "toLocation": "成都",
    "status": "IN_TRANSIT",  // NEW, IN_TRANSIT, DELIVERED 等
    "tracking": []
  }
  ```
- **权限**：`Dispatcher`、`Admin`。

### 更新运输状态 (Update Shipment)

- **URL Path**：`PUT /api/v1/shipments/{shipmentId}/status`
- **描述**：更新运输任务的状态（例如记录已发货、在途、已送达）。
- **路径参数**：
  - `shipmentId` (integer, 必填)：运输任务ID
- **请求参数**（JSON Body）：
  - `status` (string, 必填)：更新状态，如 `IN_TRANSIT`、`DELIVERED`
  - `location` (string, 可选)：当前所在地点
  - `timestamp` (string, 可选)：更新时间戳（ISO 8601格式）
- **请求示例**：
  ```json
  {
    "status": "DELIVERED",
    "location": "成都",
    "timestamp": "2025-12-01T14:30:00Z"
  }
  ```
- **响应**：HTTP 200，返回更新后的运输任务状态及轨迹记录：
  ```json
  {
    "shipmentId": 5001,
    "status": "DELIVERED",
    "tracking": [
      { "status": "NEW", "timestamp": "2025-11-30T08:00:00Z", "location": "北京" },
      { "status": "IN_TRANSIT", "timestamp": "2025-12-01T10:00:00Z", "location": "西安" },
      { "status": "DELIVERED", "timestamp": "2025-12-01T14:30:00Z", "location": "成都" }
    ]
  }
  ```
- **权限**：`Dispatcher`、`Admin`。

### 查询运输任务 (List Shipments)

- **URL Path**：`GET /api/v1/shipments`
- **描述**：查询所有运输任务列表，支持状态筛选。
- **请求参数**（Query，可选）：
  - `page`, `size`：分页
  - `status` (string)：筛选状态
- **响应**：HTTP 200，返回运输任务列表：
  ```json
  {
    "page": 1,
    "size": 10,
    "total": 2,
    "shipments": [
      { "shipmentId": 5001, "requestId": 1001, "status": "IN_TRANSIT" },
      { "shipmentId": 5002, "requestId": 1002, "status": "DELIVERED" }
    ]
  }
  ```
- **权限**：`Dispatcher`、`Admin`。

### 运输详情 (Get Shipment)

- **URL Path**：`GET /api/v1/shipments/{shipmentId}`
- **描述**：获取指定运输任务的详细信息，包括当前状态和轨迹。
- **路径参数**：
  - `shipmentId` (integer)：运输任务ID
- **响应**：HTTP 200，返回运输任务对象：
  ```json
  {
    "shipmentId": 5001,
    "requestId": 1001,
    "fromWarehouseId": 1,
    "toLocation": "成都",
    "status": "IN_TRANSIT",
    "tracking": [
      { "status": "NEW", "timestamp": "2025-11-30T08:00:00Z", "location": "北京" },
      { "status": "IN_TRANSIT", "timestamp": "2025-12-01T10:00:00Z", "location": "西安" }
    ]
  }
  ```
- **权限**：`Dispatcher`、`Admin`。

## 库存预警模块

说明：系统对库存量进行监控，当低于预设阈值时生成预警。管理员和仓库管理员可查看和配置预警阈值。

### 设置预警阈值 (Set Threshold)

- **URL Path**：`PUT /api/v1/alerts/threshold`
- **描述**：设置某物资的库存预警阈值。
- **请求参数**（JSON Body）：
  - `itemId` (integer, 必填)：物资ID
  - `threshold` (integer, 必填)：预警阈值数量
- **请求示例**：
  ```json
  {
    "itemId": 101,
    "threshold": 10
  }
  ```
- **响应**：HTTP 200，返回设置结果：
  ```json
  { "itemId": 101, "threshold": 10 }
  ```
- **权限**：`Admin`、`WarehouseManager`。

### 查询库存预警 (List Alerts)

- **URL Path**：`GET /api/v1/alerts`
- **描述**：获取当前所有库存低于阈值的物资列表。
- **响应**：HTTP 200，返回警告列表，例如：
  ```json
  [
    { "itemId": 102, "name": "睡袋", "quantity": 5, "threshold": 10 }
  ]
  ```
- **权限**：`Admin`、`WarehouseManager`。

## 统计分析模块

说明：提供各类统计报表和数据分析接口，仅限管理员（`Admin`）使用。

### 库存统计 (Inventory Statistics)

- **URL Path**：`GET /api/v1/stats/inventory`
- **描述**：获取当前库存情况统计（按物资分类汇总）。
- **响应**：HTTP 200，返回各物资总库存，例如：
  ```json
  [
    { "itemId": 101, "name": "帐篷", "totalQuantity": 50 },
    { "itemId": 102, "name": "睡袋", "totalQuantity": 25 }
  ]
  ```
- **权限**：`Admin`。

### 需求统计 (Request Statistics)

- **URL Path**：`GET /api/v1/stats/requests`
- **描述**：获取一段时间内各状态需求单数量统计（可按日期区间筛选）。
- **请求参数**（Query）：
  - `startDate`, `endDate` (string)：开始和结束日期（YYYY-MM-DD）
- **响应**：HTTP 200，返回统计结果：
  ```json
  {
    "startDate": "2025-01-01",
    "endDate": "2025-12-31",
    "totalRequests": 100,
    "byStatus": {
      "PENDING": 20,
      "ASSIGNED": 50,
      "COMPLETED": 30
    }
  }
  ```
- **权限**：`Admin`。

### 运输统计 (Shipment Statistics)

- **URL Path**：`GET /api/v1/stats/shipments`
- **描述**：统计运输任务完成情况（如按周/月运输量）。
- **请求参数**（Query）：
  - `period` (string)：统计周期，如 `weekly`、`monthly`
- **响应**：HTTP 200，返回统计数据，例如每周运输数量：
  ```json
  {
    "period": "weekly",
    "data": [
      { "week": "2025-W01", "count": 5 },
      { "week": "2025-W02", "count": 8 }
      // ...
    ]
  }
  ```
- **权限**：`Admin`。

## 微服务 gRPC 接口说明

系统采用微服务架构，各服务通过 gRPC 通信，并在 `.proto` 文件中定义接口和消息[2]。以下示例展示各服务的 gRPC 方法、请求和响应消息结构，均使用Protobuf语法。

说明：使用 gRPC 可以使 API 定义独立于实现，前后端均可通过 `.proto` 协议文件自动生成代码[2]。

### 用户服务 (UserService)

```protobuf
service UserService {
  // 创建用户
  rpc CreateUser(CreateUserRequest) returns (UserResponse);
  // 根据ID获取用户
  rpc GetUser(GetUserRequest) returns (UserResponse);
  // 列出所有用户
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
  // 更新用户
  rpc UpdateUser(UpdateUserRequest) returns (UserResponse);
  // 删除用户
  rpc DeleteUser(DeleteUserRequest) returns (Empty);
}

// 消息定义
message CreateUserRequest {
  string username = 1;
  string password = 2;
  repeated string roles = 3;
}
message GetUserRequest { int32 id = 1; }
message ListUsersRequest { int32 page = 1; int32 size = 2; }
message UpdateUserRequest {
  int32 id = 1;
  string password = 2;
  repeated string roles = 3;
}
message DeleteUserRequest { int32 id = 1; }

message UserResponse {
  int32 id = 1;
  string username = 2;
  repeated string roles = 3;
}
message ListUsersResponse {
  int32 total = 1;
  repeated UserResponse users = 2;
}
```

### 仓储服务 (WarehouseService)

```protobuf
service WarehouseService {
  rpc ListItems(ListItemsRequest) returns (ListItemsResponse);
  rpc GetItem(GetItemRequest) returns (ItemResponse);
  rpc CreateItem(CreateItemRequest) returns (ItemResponse);
  rpc UpdateItem(UpdateItemRequest) returns (ItemResponse);
  rpc DeleteItem(DeleteItemRequest) returns (Empty);

  rpc ListWarehouses(ListWarehousesRequest) returns (ListWarehousesResponse);
  rpc GetWarehouse(GetWarehouseRequest) returns (WarehouseResponse);
  rpc CreateWarehouse(CreateWarehouseRequest) returns (WarehouseResponse);
  rpc UpdateWarehouse(UpdateWarehouseRequest) returns (WarehouseResponse);
  rpc DeleteWarehouse(DeleteWarehouseRequest) returns (Empty);

  rpc AdjustInventory(AdjustInventoryRequest) returns (InventoryResponse);
}

message ListItemsRequest { int32 page = 1; int32 size = 2; string query = 3; }
message ListItemsResponse { int32 total = 1; repeated ItemResponse items = 2; }
message GetItemRequest { int32 itemId = 1; }
message CreateItemRequest { string name = 1; string unit = 2; string description = 3; }
message UpdateItemRequest { int32 itemId = 1; string name = 2; string unit = 3; string description = 4; }
message DeleteItemRequest { int32 itemId = 1; }

message ItemResponse { int32 id = 1; string name = 2; string unit = 3; string description = 4; }

message ListWarehousesRequest {}
message ListWarehousesResponse { repeated WarehouseResponse warehouses = 1; }
message GetWarehouseRequest { int32 id = 1; }
message CreateWarehouseRequest { string name = 1; string location = 2; }
message UpdateWarehouseRequest { int32 id = 1; string name = 2; string location = 3; }
message DeleteWarehouseRequest { int32 id = 1; }

message WarehouseResponse { int32 id = 1; string name = 2; string location = 3; }

message AdjustInventoryRequest {
  int32 warehouseId = 1;
  int32 itemId = 2;
  int32 amount = 3; // 正数为入库，负数为出库
}
message InventoryResponse { int32 itemId = 1; int32 quantity = 2; }
```

### 调度服务 (SchedulingService)

```protobuf
service SchedulingService {
  rpc CreateRequest(CreateRequestProto) returns (RequestResponse);
  rpc ListRequests(ListRequestsProto) returns (ListRequestsResponse);
  rpc GetRequest(GetRequestProto) returns (RequestResponse);
  rpc UpdateRequest(UpdateRequestProto) returns (RequestResponse);
  rpc DeleteRequest(DeleteRequestProto) returns (Empty);

  rpc CreateShipment(CreateShipmentProto) returns (ShipmentResponse);
  rpc UpdateShipmentStatus(UpdateShipmentProto) returns (ShipmentResponse);
  rpc ListShipments(ListShipmentsProto) returns (ListShipmentsResponse);
  rpc GetShipment(GetShipmentProto) returns (ShipmentResponse);
}

message CreateRequestProto {
  string title = 1;
  string location = 2;
  repeated ItemQuantity items = 3;
}
message RequestResponse {
  int32 id = 1;
  string title = 2;
  string location = 3;
  repeated ItemQuantity items = 4;
  string status = 5;
  int32 assignedTo = 6;
}
message ListRequestsProto { int32 page = 1; int32 size = 2; string status = 3; }
message ListRequestsResponse { int32 total = 1; repeated RequestResponse requests = 2; }
message GetRequestProto { int32 id = 1; }
message UpdateRequestProto { int32 id = 1; string status = 2; int32 assignedTo = 3; }
message DeleteRequestProto { int32 id = 1; }

message CreateShipmentProto {
  int32 requestId = 1;
  int32 fromWarehouseId = 2;
  string toLocation = 3;
  repeated ItemQuantity items = 4;
}
message ShipmentResponse {
  int32 shipmentId = 1;
  int32 requestId = 2;
  int32 fromWarehouseId = 3;
  string toLocation = 4;
  string status = 5;
  repeated TrackingInfo tracking = 6;
}
message ListShipmentsProto { int32 page = 1; int32 size = 2; string status = 3; }
message ListShipmentsResponse { int32 total = 1; repeated ShipmentResponse shipments = 2; }
message GetShipmentProto { int32 shipmentId = 1; }
message UpdateShipmentProto {
  int32 shipmentId = 1;
  string status = 2;
  string location = 3;
  string timestamp = 4;
}

message ItemQuantity { int32 itemId = 1; int32 quantity = 2; }
message TrackingInfo {
  string status = 1;
  string location = 2;
  string timestamp = 3;
}
```

### 统计服务 (StatisticsService)

```protobuf
service StatisticsService {
  rpc GetInventoryStats(Empty) returns (InventoryStatsResponse);
  rpc GetRequestStats(StatsRequest) returns (RequestStatsResponse);
  rpc GetShipmentStats(StatsRequest) returns (ShipmentStatsResponse);
}

message StatsRequest {
  string startDate = 1;
  string endDate = 2;
}
message InventoryStatsResponse {
  repeated ItemStock items = 1;
}
message ItemStock { int32 itemId = 1; string name = 2; int32 totalQuantity = 3; }

message RequestStatsResponse {
  string startDate = 1;
  string endDate = 2;
  int32 totalRequests = 3;
  map<string,int32> byStatus = 4;
}
message ShipmentStatsResponse {
  string period = 1;
  repeated ShipmentCount data = 2;
}
message ShipmentCount { string periodLabel = 1; int32 count = 2; }
```

## 引用文献

本文档设计参考了RESTful API和gRPC接口设计的最佳实践[1][2]。各接口设计符合Swagger/OpenAPI规范，可直接导入前端(Vue)项目中使用。

[1] REST API 常用的安全认证方式
https://apifox.com/apiskills/common-security-authentication-methods-for-rest-api/

[2] gRPC API详解：从实现原理到使用实例
https://apifox.com/apiskills/the-complete-guide-to-grpc-api/
