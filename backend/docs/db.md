下面给你一份**面向微服务拆分 + 高并发库存场景**的数据库表设计方案。

设计目标：

* 满足 RBAC 权限模型
* 支持多仓库、多物资
* 库存强一致（可加行级锁 / 乐观锁）
* 支持需求单 → 运输单 → 轨迹完整链路
* 支持统计与库存预警
* 易于拆分为独立微服务数据库

默认数据库：**PostgreSQL / MySQL 8+**
主键：`BIGINT`
时间字段：`TIMESTAMP`
字符集：`utf8mb4`

---

# 一、认证与RBAC模型

---

## 1️⃣ users

```sql
CREATE TABLE users (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    username        VARCHAR(64) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

说明：

* 密码必须存 hash（bcrypt/argon2）
* 不建议明文

---

## 2️⃣ roles

```sql
CREATE TABLE roles (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    name        VARCHAR(64) NOT NULL UNIQUE,
    description VARCHAR(255)
);
```

---

## 3️⃣ user_roles（多对多）

```sql
CREATE TABLE user_roles (
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);
```

---

## 4️⃣ （可选）refresh_tokens

如果你使用 Refresh Token：

```sql
CREATE TABLE refresh_tokens (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id     BIGINT NOT NULL,
    token       VARCHAR(512) NOT NULL,
    expires_at  TIMESTAMP NOT NULL,
    revoked     BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

---

# 二、仓储服务（WarehouseService）

---

## 5️⃣ items（物资表）

```sql
CREATE TABLE items (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    name        VARCHAR(128) NOT NULL,
    unit        VARCHAR(32) NOT NULL,
    description TEXT,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_item_name (name)
);
```

---

## 6️⃣ warehouses

```sql
CREATE TABLE warehouses (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    name        VARCHAR(128) NOT NULL,
    location    VARCHAR(255),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

---

## 7️⃣ inventory（核心库存表）

这是最关键的表。

```sql
CREATE TABLE inventory (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    warehouse_id    BIGINT NOT NULL,
    item_id         BIGINT NOT NULL,
    quantity        INT NOT NULL DEFAULT 0,
    version         INT NOT NULL DEFAULT 0,  -- 乐观锁版本号
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_warehouse_item (warehouse_id, item_id),

    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
);
```

说明：

* `UNIQUE(warehouse_id, item_id)` 保证唯一库存记录
* `version` 用于乐观锁更新：

  ```sql
  UPDATE inventory
  SET quantity = quantity - 5, version = version + 1
  WHERE id = ? AND version = ?
  ```

---

## 8️⃣ inventory_logs（库存流水表 🔥 强烈建议）

```sql
CREATE TABLE inventory_logs (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    warehouse_id    BIGINT NOT NULL,
    item_id         BIGINT NOT NULL,
    change_amount   INT NOT NULL,  -- 正数入库 负数出库
    before_qty      INT NOT NULL,
    after_qty       INT NOT NULL,
    reference_type  VARCHAR(32),   -- REQUEST / SHIPMENT / MANUAL
    reference_id    BIGINT,
    operator_id     BIGINT,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (warehouse_id) REFERENCES warehouses(id),
    FOREIGN KEY (item_id) REFERENCES items(id)
);
```

这是审计与统计的核心表。

---

# 三、库存预警模块

---

## 9️⃣ item_thresholds

```sql
CREATE TABLE item_thresholds (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    item_id     BIGINT NOT NULL UNIQUE,
    threshold   INT NOT NULL,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
);
```

库存预警通过：

```sql
SELECT i.id, i.name, inv.quantity, t.threshold
FROM inventory inv
JOIN items i ON i.id = inv.item_id
JOIN item_thresholds t ON t.item_id = i.id
WHERE inv.quantity < t.threshold;
```

---

# 四、调度服务（SchedulingService）

---

## 🔟 requests（需求单主表）

```sql
CREATE TABLE requests (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    title           VARCHAR(255) NOT NULL,
    location        VARCHAR(255) NOT NULL,
    status          VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    assigned_to     BIGINT NULL,
    created_by      BIGINT NOT NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (assigned_to) REFERENCES users(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);
```

状态建议：

```
PENDING
ASSIGNED
COMPLETED
CANCELLED
```

---

## 1️⃣1️⃣ request_items

```sql
CREATE TABLE request_items (
    id          BIGINT PRIMARY KEY AUTO_INCREMENT,
    request_id  BIGINT NOT NULL,
    item_id     BIGINT NOT NULL,
    quantity    INT NOT NULL,

    FOREIGN KEY (request_id) REFERENCES requests(id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items(id)
);
```

---

# 五、物流服务（Shipment）

---

## 1️⃣2️⃣ shipments

```sql
CREATE TABLE shipments (
    id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
    request_id          BIGINT NOT NULL,
    from_warehouse_id   BIGINT NOT NULL,
    to_location         VARCHAR(255) NOT NULL,
    status              VARCHAR(32) NOT NULL DEFAULT 'NEW',
    created_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (request_id) REFERENCES requests(id),
    FOREIGN KEY (from_warehouse_id) REFERENCES warehouses(id)
);
```

状态：

```
NEW
IN_TRANSIT
DELIVERED
CANCELLED
```

---

## 1️⃣3️⃣ shipment_items

```sql
CREATE TABLE shipment_items (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    shipment_id     BIGINT NOT NULL,
    item_id         BIGINT NOT NULL,
    quantity        INT NOT NULL,

    FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items(id)
);
```

---

## 1️⃣4️⃣ shipment_tracking（轨迹表）

```sql
CREATE TABLE shipment_tracking (
    id              BIGINT PRIMARY KEY AUTO_INCREMENT,
    shipment_id     BIGINT NOT NULL,
    status          VARCHAR(32) NOT NULL,
    location        VARCHAR(255),
    recorded_at     TIMESTAMP NOT NULL,

    FOREIGN KEY (shipment_id) REFERENCES shipments(id) ON DELETE CASCADE
);
```

建议索引：

```sql
CREATE INDEX idx_tracking_shipment ON shipment_tracking(shipment_id);
```

---

# 六、统计模块设计说明

统计接口不要额外建表，直接基于：

* `inventory`
* `requests`
* `shipments`
* `inventory_logs`

做聚合查询。

例如：

### 库存统计

```sql
SELECT item_id, SUM(quantity) AS total_quantity
FROM inventory
GROUP BY item_id;
```

---

### 需求统计

```sql
SELECT status, COUNT(*)
FROM requests
WHERE created_at BETWEEN ? AND ?
GROUP BY status;
```

---

### 运输统计（按周）

```sql
SELECT YEARWEEK(created_at), COUNT(*)
FROM shipments
GROUP BY YEARWEEK(created_at);
```

---

# 七、微服务数据库拆分建议

如果你做真正的微服务：

| 服务                 | 数据库                                                          |
| ------------------ | ------------------------------------------------------------ |
| user-service       | users / roles / user_roles                                   |
| warehouse-service  | items / warehouses / inventory / inventory_logs / thresholds |
| scheduling-service | requests / request_items                                     |
| shipment-service   | shipments / shipment_items / shipment_tracking               |
| statistics-service | 只读，不持久化                                                      |

---

# 八、关键架构建议（非常重要）

### 1️⃣ 库存扣减必须使用事务

```
BEGIN;
SELECT quantity FROM inventory WHERE ... FOR UPDATE;
UPDATE inventory SET quantity = quantity - ?;
INSERT inventory_logs ...
COMMIT;
```

---

### 2️⃣ 不要直接改库存

永远通过：

```
inventory_logs + inventory
```

---

### 3️⃣ request 与 shipment 分离

不要让 request 直接操作 inventory。

流程应为：

```
request → assigned → shipment → 扣库存 → 生成库存流水
```

---
