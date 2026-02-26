<script setup lang="ts">
/**
 * 调度管理：
 * - 需求（Requests）：创建/查询/更新/删除
 * - 运输任务（Shipments）：创建任务、更新运输状态、查看详情
 * - 地图追踪（Map）：基于高德地图展示运输路径与节点
 *
 * 说明：对后端返回字段做 normalize，兼容不同命名风格。
 */
import { computed, nextTick, onMounted, reactive, ref, watch } from "vue";
import { List, Map as MapIcon, Plus, RefreshCw } from "lucide-vue-next";
import AMapLoader from "@amap/amap-jsapi-loader";
import { useAuthStore } from "@/stores/auth";
import request from "@/api/request";
import { ElMessage, ElMessageBox } from "element-plus";

type Item = { id: number; name: string; unit: string; description?: string };
type ItemQuantity = { itemId: number; quantity: number };
type RescueRequest = {
  id: number;
  title: string;
  location: string;
  status: string;
  items?: ItemQuantity[];
  assignedTo?: number | null;
  createdAt?: string;
};
type ShipmentTracking = {
  status: string;
  location?: string;
  timestamp?: string;
};
type Shipment = {
  shipmentId: number;
  requestId: number;
  fromWarehouseId: number;
  toLocation: string;
  status: string;
  tracking?: ShipmentTracking[];
  items?: ItemQuantity[];
  createdAt?: string;
};
type Warehouse = { id: number; name: string; location?: string };

// 统一规范后端数据字段，避免大小写/命名不一致导致渲染失败
const normalizeItem = (raw: any): Item => ({
  id: Number(raw?.id ?? raw?.ID ?? raw?.itemId ?? raw?.ItemID ?? 0),
  name: String(raw?.name ?? raw?.Name ?? ""),
  unit: String(raw?.unit ?? raw?.Unit ?? ""),
  description: raw?.description ?? raw?.Description ?? undefined,
});

const normalizeItemQuantity = (raw: any): ItemQuantity => ({
  itemId: Number(raw?.itemId ?? raw?.ItemID ?? raw?.itemID ?? 0),
  quantity: Number(raw?.quantity ?? raw?.Quantity ?? 0),
});

const normalizeRequest = (raw: any): RescueRequest => ({
  id: Number(raw?.id ?? raw?.ID ?? 0),
  title: String(raw?.title ?? raw?.Title ?? ""),
  location: String(raw?.location ?? raw?.Location ?? ""),
  status: String(raw?.status ?? raw?.Status ?? ""),
  assignedTo:
    raw?.assignedTo ??
    raw?.AssignedTo ??
    raw?.assigned_to ??
    raw?.Assigned_To ??
    undefined,
  items: Array.isArray(raw?.items ?? raw?.Items)
    ? (raw?.items ?? raw?.Items).map(normalizeItemQuantity)
    : undefined,
  createdAt:
    raw?.createdAt ??
    raw?.CreatedAt ??
    raw?.created_at ??
    raw?.Created_At ??
    undefined,
});

const normalizeTracking = (raw: any): ShipmentTracking => ({
  status: String(raw?.status ?? raw?.Status ?? ""),
  location: raw?.location ?? raw?.Location ?? undefined,
  timestamp:
    raw?.timestamp ??
    raw?.Timestamp ??
    raw?.recordedAt ??
    raw?.RecordedAt ??
    undefined,
});

const normalizeWarehouse = (raw: any): Warehouse => ({
  id: Number(raw?.id ?? raw?.ID ?? 0),
  name: String(raw?.name ?? raw?.Name ?? ""),
  location: raw?.location ?? raw?.Location ?? undefined,
});

// 运输任务字段兼容处理：保证 shipmentId/requestId/fromWarehouseId 有值
const normalizeShipment = (raw: any): Shipment => ({
  shipmentId: Number(
    raw?.shipmentId ?? raw?.ShipmentId ?? raw?.id ?? raw?.ID ?? 0
  ),
  requestId: Number(raw?.requestId ?? raw?.RequestId ?? raw?.RequestID ?? 0),
  fromWarehouseId: Number(
    raw?.fromWarehouseId ?? raw?.FromWarehouseId ?? raw?.FromWarehouseID ?? 0
  ),
  toLocation: String(raw?.toLocation ?? raw?.ToLocation ?? ""),
  status: String(raw?.status ?? raw?.Status ?? ""),
  tracking: Array.isArray(raw?.tracking ?? raw?.Tracking)
    ? (raw?.tracking ?? raw?.Tracking).map(normalizeTracking)
    : undefined,
  items: Array.isArray(raw?.items ?? raw?.Items)
    ? (raw?.items ?? raw?.Items).map(normalizeItemQuantity)
    : undefined,
  createdAt:
    raw?.createdAt ??
    raw?.CreatedAt ??
    raw?.created_at ??
    raw?.Created_At ??
    undefined,
});

const authStore = useAuthStore();
const activeTab = ref<"requests" | "shipments" | "map">("requests");

// 状态码到中文标签的映射
const statusLabelMap: Record<string, string> = {
  PENDING: "待处理",
  ASSIGNED: "已批准",
  COMPLETED: "已完成",
  CANCELLED: "已取消",
  NEW: "新建",
  IN_TRANSIT: "运输中",
  DELIVERED: "已送达",
};

const items = ref<Item[]>([]);
// 物资列表：用于创建需求单/运输任务的选择项
const loadItems = async () => {
  const res: any = await request.get("/api/v1/items", {
    params: { page: 1, size: 1000 },
  });
  const list = res?.items ?? res?.Items ?? res ?? [];
  items.value = Array.isArray(list) ? list.map(normalizeItem) : [];
};

const warehouses = ref<Warehouse[]>([]);
// 仓库列表：地图轨迹与运输任务创建使用
const loadWarehouses = async () => {
  const res: any = await request.get("/api/v1/warehouses");
  const list = res?.warehouses ?? res?.Warehouses ?? res ?? [];
  warehouses.value = Array.isArray(list) ? list.map(normalizeWarehouse) : [];
};

const requestsLoading = ref(false);
const requestsData = ref<RescueRequest[]>([]);
const requestsTotal = ref(0);
const requestQuery = reactive({ page: 1, size: 10, status: "" });

// 需求单列表：分页 + 状态筛选
const loadRequests = async () => {
  requestsLoading.value = true;
  try {
    const res: any = await request.get("/api/v1/requests", {
      params: { ...requestQuery, status: requestQuery.status || undefined },
    });
    const list = res?.requests ?? res?.Requests ?? res ?? [];
    requestsData.value = Array.isArray(list) ? list.map(normalizeRequest) : [];
    requestsTotal.value =
      Number(res?.total ?? res?.Total) || requestsData.value.length;
  } finally {
    requestsLoading.value = false;
  }
};

const shipmentsLoading = ref(false);
const shipmentsData = ref<Shipment[]>([]);
const shipmentsTotal = ref(0);
const shipmentQuery = reactive({ page: 1, size: 10, status: "" });

// 运输任务列表：分页 + 状态筛选
const loadShipments = async () => {
  shipmentsLoading.value = true;
  try {
    const res: any = await request.get("/api/v1/shipments", {
      params: { ...shipmentQuery, status: shipmentQuery.status || undefined },
    });
    const list = res?.shipments ?? res?.Shipments ?? res ?? [];
    shipmentsData.value = Array.isArray(list)
      ? list.map(normalizeShipment)
      : [];
    shipmentsTotal.value =
      Number(res?.total ?? res?.Total) || shipmentsData.value.length;
  } finally {
    shipmentsLoading.value = false;
  }
};

const requestDialogVisible = ref(false);
const requestDialogMode = ref<"create" | "edit">("create");
const requestForm = reactive({
  id: 0,
  title: "",
  location: "",
  status: "",
  assignedTo: undefined as number | undefined,
  items: [] as { itemId: number | null; quantity: number }[],
});

// 打开新建需求单：初始化表单
const openCreateRequest = () => {
  requestDialogMode.value = "create";
  requestForm.id = 0;
  requestForm.title = "";
  requestForm.location = "";
  requestForm.status = "";
  requestForm.assignedTo = undefined;
  requestForm.items = [{ itemId: null, quantity: 1 }];
  requestDialogVisible.value = true;
};

// 打开编辑需求单：加载详情并填充表单
const openEditRequest = async (row: RescueRequest) => {
  requestDialogMode.value = "edit";
  const res: any = await request.get(`/api/v1/requests/${row.id}`);
  const normalized = normalizeRequest(res ?? row);
  requestForm.id = normalized.id;
  requestForm.title = normalized.title;
  requestForm.location = normalized.location;
  requestForm.status = normalized.status;
  requestForm.assignedTo =
    normalized.assignedTo == null ? undefined : Number(normalized.assignedTo);
  requestForm.items = (normalized.items ?? []).map((it: any) => ({
    itemId: Number(it.itemId),
    quantity: Number(it.quantity),
  }));
  if (requestForm.items.length === 0)
    requestForm.items = [{ itemId: null, quantity: 1 }];
  requestDialogVisible.value = true;
};

// 保存需求单：创建 or 更新（状态/指派）
const saveRequest = async () => {
  const payloadItems = requestForm.items
    .filter((it) => it.itemId != null && it.quantity > 0)
    .map((it) => ({
      itemId: Number(it.itemId),
      quantity: Number(it.quantity),
    }));

  if (requestDialogMode.value === "create") {
    if (
      !requestForm.title ||
      !requestForm.location ||
      payloadItems.length === 0
    ) {
      ElMessage.warning("请填写标题、地点和物资明细");
      return;
    }
    await request.post("/api/v1/requests", {
      title: requestForm.title,
      location: requestForm.location,
      items: payloadItems,
    });
    ElMessage.success("需求单已创建");
  } else {
    const payload: any = {};
    if (requestForm.status) payload.status = requestForm.status;
    if (requestForm.assignedTo != null)
      payload.assignedTo = requestForm.assignedTo;
    await request.put(`/api/v1/requests/${requestForm.id}`, payload);
    ElMessage.success("需求单已更新");
  }
  requestDialogVisible.value = false;
  await loadRequests();
};

// 批准需求单：将状态更新为 ASSIGNED，并引导创建运输任务
const approveRequest = async (row: RescueRequest) => {
  shipmentDialogMode.value = "create";
  shipmentForm.shipmentId = 0;
  shipmentForm.requestId = row.id;
  shipmentForm.fromWarehouseId = undefined;
  shipmentForm.toLocation = row.location;
  shipmentForm.items = (row.items ?? []).map((it) => ({
    itemId: Number(it.itemId),
    quantity: Number(it.quantity),
  }));
  if (shipmentForm.items.length === 0)
    shipmentForm.items = [{ itemId: null, quantity: 1 }];

  shipmentDialogVisible.value = true;
  ElMessage.info("请选择发货仓库以完成批准和派单");
};

// 驳回需求单：将状态更新为 CANCELLED
const rejectRequest = async (row: RescueRequest) => {
  await ElMessageBox.prompt("请输入驳回原因", "驳回需求单", {
    confirmButtonText: "确认驳回",
    cancelButtonText: "取消",
    inputPattern: /\S+/,
    inputErrorMessage: "驳回原因不能为空",
  });
  // 暂未在后端存储原因字段，仅更新状态
  await request.put(`/api/v1/requests/${row.id}`, { status: "CANCELLED" });
  ElMessage.warning("需求单已驳回");
  await loadRequests();
};

// 删除需求单（仅管理员可见）
const deleteRequest = async (row: RescueRequest) => {
  await ElMessageBox.confirm(
    `确认删除需求单 #${row.id}？仅未指派状态可删除。`,
    "提示",
    { type: "warning" }
  );
  await request.delete(`/api/v1/requests/${row.id}`);
  ElMessage.success("已删除");
  await loadRequests();
};

const shipmentDialogVisible = ref(false);
const shipmentDialogMode = ref<"create" | "status">("create");
const shipmentForm = reactive({
  shipmentId: 0,
  requestId: undefined as number | undefined,
  fromWarehouseId: undefined as number | undefined,
  toLocation: "",
  status: "IN_TRANSIT",
  location: "",
  timestamp: "",
  items: [] as { itemId: number | null; quantity: number }[],
});

// 打开新建运输任务弹窗
const openCreateShipment = () => {
  shipmentDialogMode.value = "create";
  shipmentForm.shipmentId = 0;
  shipmentForm.requestId = undefined;
  shipmentForm.fromWarehouseId = undefined;
  shipmentForm.toLocation = "";
  shipmentForm.items = [{ itemId: null, quantity: 1 }];
  shipmentDialogVisible.value = true;
};

// 加载运输任务对应的需求单明细
const loadRequestForShipment = async () => {
  if (!shipmentForm.requestId) {
    ElMessage.warning("请先输入需求单 ID");
    return;
  }
  const res: any = await request.get(
    `/api/v1/requests/${shipmentForm.requestId}`
  );
  const normalized = normalizeRequest(res ?? {});
  shipmentForm.toLocation = normalized.location ?? "";
  shipmentForm.items = (normalized.items ?? []).map((it: any) => ({
    itemId: Number(it.itemId),
    quantity: Number(it.quantity),
  }));
  if (shipmentForm.items.length === 0)
    shipmentForm.items = [{ itemId: null, quantity: 1 }];
  ElMessage.success("已加载需求单明细");
};

// 打开更新运输状态：只更新状态/位置/时间
const openUpdateShipmentStatus = (row: Shipment) => {
  shipmentDialogMode.value = "status";
  shipmentForm.shipmentId = row.shipmentId;
  shipmentForm.status = row.status;
  shipmentForm.location = "";
  shipmentForm.timestamp = "";
  shipmentDialogVisible.value = true;
};

// 保存运输任务：创建任务或更新运输状态
const saveShipment = async () => {
  if (shipmentDialogMode.value === "create") {
    const payloadItems = shipmentForm.items
      .filter((it) => it.itemId != null && it.quantity > 0)
      .map((it) => ({
        itemId: Number(it.itemId),
        quantity: Number(it.quantity),
      }));

    if (
      !shipmentForm.requestId ||
      !shipmentForm.fromWarehouseId ||
      !shipmentForm.toLocation ||
      payloadItems.length === 0
    ) {
      ElMessage.warning("请填写需求单ID、出发仓库、目的地和物资明细");
      return;
    }
    // 创建时默认状态设为 IN_TRANSIT（运输中），而不是 NEW
    await request.post("/api/v1/shipments", {
      requestId: shipmentForm.requestId,
      fromWarehouseId: shipmentForm.fromWarehouseId,
      toLocation: shipmentForm.toLocation,
      status: "IN_TRANSIT",
      items: payloadItems,
    });
    ElMessage.success("运输任务已创建并开始派送");
  } else {
    if (!shipmentForm.status) {
      ElMessage.warning("请选择状态");
      return;
    }
    const payload: any = { status: shipmentForm.status };
    if (shipmentForm.location) payload.location = shipmentForm.location;
    if (shipmentForm.timestamp) payload.timestamp = shipmentForm.timestamp;
    await request.put(
      `/api/v1/shipments/${shipmentForm.shipmentId}/status`,
      payload
    );
    ElMessage.success("运输状态已更新");
  }
  shipmentDialogVisible.value = false;
  await loadShipments();
  if (activeTab.value === "map" && selectedShipmentId.value) {
    await loadSelectedShipment();
    await renderShipmentOnMap();
  }
  if (activeTab.value === "requests") await loadRequests();
};

// 签收运输任务
const completeShipment = async (row: Shipment) => {
  await ElMessageBox.confirm(`确认签收运输任务 #${row.shipmentId}？`, "提示", {
    type: "success",
  });
  await request.put(`/api/v1/shipments/${row.shipmentId}/status`, {
    status: "DELIVERED",
    location: row.toLocation, // 默认使用目的地作为签收地点
    timestamp: new Date().toISOString(),
  });
  ElMessage.success("已签收");
  await loadShipments();
};

// 拒绝签收运输任务
const rejectShipment = async (row: Shipment) => {
  await ElMessageBox.confirm(
    `确认拒绝签收运输任务 #${row.shipmentId}？`,
    "拒绝签收",
    { type: "warning", confirmButtonText: "确认拒绝", cancelButtonText: "取消" }
  );
  await request.put(`/api/v1/shipments/${row.shipmentId}/status`, {
    status: "CANCELLED",
    location: row.toLocation,
    timestamp: new Date().toISOString(),
  });
  ElMessage.warning("已拒绝签收");
  await loadShipments();
};

// 当前地图中选择的运输任务与详情
const selectedShipmentId = ref<number | null>(null);
const selectedShipment = ref<Shipment | null>(null);

// 高德地图实例与工具
const mapContainer = ref<HTMLElement | null>(null);
let map: any = null;
let amapApi: any = null;
let geocoder: any = null;
let driving: any = null;
const drivingSummary = ref<{ distance: number; time: number } | null>(null);
// 高德地图 Web 端 key 与安全密钥（前端可见）：如需替换请在此处修改或改为环境变量注入
const amapKey = "0282382759b77d9371ab6f78e022bfeb";
const amapSecurityCode = "3512fe5d4078e94b9dca5b2f2f8cb6eb";

// 拉取运输任务详情（包含 tracking）
const loadSelectedShipment = async () => {
  if (!selectedShipmentId.value) return;
  const res: any = await request.get(
    `/api/v1/shipments/${selectedShipmentId.value}`
  );
  selectedShipment.value = normalizeShipment(res ?? {});
};

// 懒加载高德 JSAPI，并初始化地理编码/驾车插件
const ensureAmapApi = async () => {
  if (amapApi) return amapApi;
  try {
    if (amapSecurityCode) {
      (window as any)._AMapSecurityConfig = {
        securityJsCode: amapSecurityCode,
      };
    }
    const AMap = await (AMapLoader as any).load({
      key: amapKey,
      securityJsCode: amapSecurityCode,
      version: "2.0",
      plugins: ["AMap.Geocoder", "AMap.Driving"],
    });
    amapApi = AMap;
    geocoder = new AMap.Geocoder();
    return amapApi;
  } catch {
    ElMessage.error("地图加载失败，请检查高德 Key 与安全密钥配置");
    return null;
  }
};

const getWarehouseLocation = (warehouseId: number) => {
  const warehouse = warehouses.value.find((w) => w.id === warehouseId);
  return warehouse?.location ? String(warehouse.location) : "";
};

const getWarehouseName = (warehouseId: number) => {
  const warehouse = warehouses.value.find((w) => w.id === warehouseId);
  return warehouse ? warehouse.name : String(warehouseId);
};

const formatItems = (rowItems?: ItemQuantity[]) => {
  if (!rowItems || rowItems.length === 0) return "-";
  return rowItems
    .map((it) => {
      const item = items.value.find((i) => i.id === it.itemId);
      return `${item?.name || it.itemId} x${it.quantity}`;
    })
    .join(", ");
};

const formatTime = (time?: string) => {
  if (!time) return "-";
  return new Date(time).toLocaleString();
};

// 运输轨迹显示：优先使用 tracking，若为空则用起点/终点兜底
const buildTrackingDisplay = (shipment: Shipment) => {
  const tracking = shipment.tracking ?? [];
  const list = tracking.map((t) => ({
    status: t.status,
    location: t.location || "",
    timestamp: t.timestamp || "",
  }));
  if (list.length) return list;
  const fallback: { status: string; location: string; timestamp: string }[] =
    [];
  const fromLocation = getWarehouseLocation(shipment.fromWarehouseId);
  const toLocation = shipment.toLocation;
  if (fromLocation)
    fallback.push({ status: "起点", location: fromLocation, timestamp: "" });
  if (toLocation)
    fallback.push({ status: "终点", location: toLocation, timestamp: "" });
  return fallback;
};

// 轨迹路线规划：起点 + tracking + 终点（去重）
const buildTrackingLocations = (shipment: Shipment) => {
  const fromLocation = getWarehouseLocation(shipment.fromWarehouseId);
  const toLocation = shipment.toLocation;
  const locations: string[] = [];
  const pushLocation = (loc?: string) => {
    if (!loc) return;
    const val = String(loc);
    if (!val) return;
    if (locations.length === 0 || locations[locations.length - 1] !== val)
      locations.push(val);
  };

  pushLocation(fromLocation);
  (shipment.tracking ?? []).forEach((t) => {
    if (!t.location) return;
    const loc = t.location === "Warehouse" ? fromLocation || "" : t.location;
    if (!loc) return;
    pushLocation(loc);
  });
  pushLocation(toLocation);
  return locations;
};

const trackingDisplay = computed(() =>
  selectedShipment.value ? buildTrackingDisplay(selectedShipment.value) : []
);

// 初始化地图容器与驾车实例
const initMap = async () => {
  if (!mapContainer.value) return;
  const AMap = await ensureAmapApi();
  if (!AMap) return;

  map = new AMap.Map(mapContainer.value, {
    viewMode: "3D",
    zoom: 5,
    center: [104.195397, 35.86166],
    mapStyle: "amap://styles/dark",
    resizeEnable: true,
  });
  driving = new AMap.Driving({
    map,
    showTraffic: true,
    hideMarkers: true,
  });
};

// 解析 "lng,lat" 字符串为坐标点
const parseLocationToPoint = (loc: string) => {
  const match = loc.match(/(-?\d+(?:\.\d+)?)[,\s]+(-?\d+(?:\.\d+)?)/);
  if (!match) return null;
  const lng = Number(match[1]);
  const lat = Number(match[2]);
  if (Number.isNaN(lng) || Number.isNaN(lat)) return null;
  return [lng, lat];
};

// 统一转为 AMap 的 LngLat 实例
const toLngLat = (AMap: any, point: any) => {
  if (Array.isArray(point)) return new AMap.LngLat(point[0], point[1]);
  return point;
};

// 驾车距离格式化（m/km）
const formatDistanceMeters = (meters: number) => {
  if (!Number.isFinite(meters)) return "-";
  if (meters < 1000) return `${Math.round(meters)} m`;
  return `${(meters / 1000).toFixed(1)} km`;
};

// 驾车时间格式化（秒 -> 小时/分钟）
const formatDurationSeconds = (seconds: number) => {
  if (!Number.isFinite(seconds)) return "-";
  const s = Math.max(0, Math.round(seconds));
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  if (h > 0 && m > 0) return `${h} 小时 ${m} 分钟`;
  if (h > 0) return `${h} 小时`;
  return `${m} 分钟`;
};

// 地图渲染：定位轨迹 -> 驾车规划 -> marker/折线渲染
const renderShipmentOnMap = async () => {
  if (!selectedShipment.value) return;
  if (!map) {
    await nextTick();
    await initMap();
  }
  if (!map) return;
  if (warehouses.value.length === 0) await loadWarehouses();
  map.resize();
  map.clearMap();
  if (driving?.clear) driving.clear();
  drivingSummary.value = null;

  const AMap = amapApi || (await ensureAmapApi());
  if (!AMap) return;
  const localGeocoder = geocoder || new AMap.Geocoder();

  const locations = buildTrackingLocations(selectedShipment.value);
  if (locations.length === 0) {
    ElMessage.info("该运输任务暂无可用位置轨迹");
    return;
  }

  const drivingNodes: any[] = [];
  const points: any[] = [];
  for (const loc of locations) {
    const direct = parseLocationToPoint(loc);
    if (direct) {
      drivingNodes.push(toLngLat(AMap, direct));
      points.push(direct);
      continue;
    }
    drivingNodes.push({ keyword: loc });
    const p = await new Promise<any | null>((resolve) => {
      localGeocoder.getLocation(loc, (status: string, result: any) => {
        if (status === "complete" && result?.geocodes?.[0]?.location)
          resolve(result.geocodes[0].location);
        else resolve(null);
      });
    });
    if (p) points.push(p);
  }

  if (drivingNodes.length < 2) {
    ElMessage.warning("轨迹地点不足，无法进行驾车规划");
    return;
  }

  let drivingOk = false;
  if (driving) {
    await new Promise<void>((resolve) => {
      driving.search(drivingNodes, (status: string, result: any) => {
        if (status === "complete" && result?.routes?.[0]) {
          const route = result.routes[0];
          if (
            Number.isFinite(route?.distance) &&
            Number.isFinite(route?.time)
          ) {
            drivingSummary.value = {
              distance: Number(route.distance),
              time: Number(route.time),
            };
          }
          drivingOk = true;
        } else {
          ElMessage.warning("驾车路径规划失败，已回退直线路径");
        }
        resolve();
      });
    });
  }

  if (!drivingOk) {
    if (points.length === 0) {
      ElMessage.warning("无法将轨迹地点解析为坐标，请检查地点文本");
      return;
    }
    if (points.length >= 2) {
      const polyline = new AMap.Polyline({
        path: points,
        strokeColor: "#3b82f6",
        strokeOpacity: 1,
        strokeWeight: 6,
        lineJoin: "round",
        lineCap: "round",
      });
      map.add(polyline);
    }
  }

  if (points.length > 0) {
    points.forEach((p, idx) => {
      new AMap.Marker({
        position: p,
        map,
        title:
          idx === 0
            ? "起点"
            : idx === points.length - 1
            ? "终点"
            : `节点${idx + 1}`,
      });
    });
  }

  map.setFitView();
};

const canDeleteRequest = computed(() => authStore.isAdmin);
const canManageRequest = computed(() => authStore.isAdmin);
const isDispatcher = computed(() => authStore.isDispatcher);

// Tab 切换时按需加载数据并刷新地图
watch(activeTab, async (tab) => {
  if (tab === "requests") await loadRequests();
  if (tab === "shipments") await loadShipments();
  if (tab === "map") {
    await nextTick();
    if (!map) await initMap();
    if (map) map.resize();
    if (warehouses.value.length === 0) await loadWarehouses();
    if (selectedShipmentId.value) {
      await loadSelectedShipment();
      await renderShipmentOnMap();
    }
  }
});

// 初次进入页面：准备基础数据与需求单列表
onMounted(async () => {
  await loadItems();
  await loadWarehouses();
  await loadRequests();
});
</script>

<template>
  <div class="scheduling-container">
    <div class="tabs-header">
      <div
        v-for="tab in ['requests', 'shipments', 'map']"
        :key="tab"
        :class="['tab-item', { active: activeTab === tab }]"
        @click="activeTab = tab as any"
      >
        <component :is="tab === 'map' ? MapIcon : List" :size="18" />
        {{
          tab === "requests"
            ? "需求单"
            : tab === "shipments"
            ? "运输任务"
            : "地图追踪"
        }}
      </div>
    </div>

    <div v-if="activeTab === 'requests'" class="panel">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-select
            v-model="requestQuery.status"
            placeholder="状态筛选"
            clearable
            style="width: 180px"
            @change="loadRequests"
          >
            <el-option label="待处理" value="PENDING" />
            <el-option label="已批准" value="ASSIGNED" />
            <el-option label="已完成" value="COMPLETED" />
            <el-option label="已取消" value="CANCELLED" />
          </el-select>
          <el-button :icon="RefreshCw" @click="loadRequests">刷新</el-button>
        </div>
        <el-button type="primary" :icon="Plus" @click="openCreateRequest"
          >新建需求单</el-button
        >
      </div>

      <el-card class="table-card" shadow="never">
        <el-table
          :data="requestsData"
          v-loading="requestsLoading"
          style="width: 100%"
        >
          <el-table-column prop="id" label="ID" width="90" />
          <el-table-column prop="title" label="标题" min-width="120" />
          <el-table-column label="物资" min-width="160">
            <template #default="{ row }">
              {{ formatItems(row.items) }}
            </template>
          </el-table-column>
          <el-table-column prop="location" label="地点" min-width="120" />
          <el-table-column label="创建时间" width="180">
            <template #default="{ row }">
              {{ formatTime(row.createdAt) }}
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="140">
            <template #default="{ row }">
              <el-tag
                :type="
                  row.status === 'PENDING'
                    ? 'warning'
                    : row.status === 'IN_TRANSIT' || row.status === 'ASSIGNED'
                    ? 'primary'
                    : row.status === 'COMPLETED'
                    ? 'success'
                    : row.status === 'CANCELLED'
                    ? 'info'
                    : ''
                "
                effect="plain"
                >{{ statusLabelMap[row.status] || row.status }}</el-tag
              >
            </template>
          </el-table-column>
          <el-table-column
            label="操作"
            width="240"
            fixed="right"
            v-if="canManageRequest"
          >
            <template #default="{ row }">
              <el-button link type="primary" @click="openEditRequest(row)"
                >查看/更新</el-button
              >
              <el-button
                v-if="row.status === 'PENDING'"
                link
                type="success"
                @click="approveRequest(row)"
                >批准</el-button
              >
              <el-button
                v-if="row.status === 'PENDING'"
                link
                type="warning"
                @click="rejectRequest(row)"
                >驳回</el-button
              >
              <el-button
                v-if="canDeleteRequest && row.status === 'PENDING'"
                link
                type="danger"
                @click="deleteRequest(row)"
                >删除</el-button
              >
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <div class="pagination">
        <el-pagination
          background
          layout="prev, pager, next"
          :page-size="requestQuery.size"
          :current-page="requestQuery.page"
          :total="requestsTotal"
          @current-change="(p:number) => { requestQuery.page = p; loadRequests() }"
        />
      </div>
    </div>

    <div v-else-if="activeTab === 'shipments'" class="panel">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-select
            v-model="shipmentQuery.status"
            placeholder="状态筛选"
            clearable
            style="width: 180px"
            @change="loadShipments"
          >
            <!-- <el-option label="新建" value="NEW" /> -->
            <el-option label="运输中" value="IN_TRANSIT" />
            <el-option label="已送达" value="DELIVERED" />
            <el-option label="已取消" value="CANCELLED" />
          </el-select>
          <el-button :icon="RefreshCw" @click="loadShipments">刷新</el-button>
        </div>
        <el-button
          v-if="canManageRequest"
          type="primary"
          :icon="Plus"
          @click="openCreateShipment"
          >新建运输任务</el-button
        >
      </div>

      <el-card class="table-card" shadow="never">
        <el-table
          :data="shipmentsData"
          v-loading="shipmentsLoading"
          style="width: 100%"
        >
          <el-table-column prop="shipmentId" label="运输ID" width="90" />
          <el-table-column prop="requestId" label="需求ID" width="90" />
          <el-table-column label="出发仓库" min-width="120">
            <template #default="{ row }">
              {{ getWarehouseName(row.fromWarehouseId) }}
            </template>
          </el-table-column>
          <el-table-column prop="toLocation" label="目的地" min-width="120" />
          <el-table-column label="物资" min-width="160">
            <template #default="{ row }">
              {{ formatItems(row.items) }}
            </template>
          </el-table-column>
          <el-table-column label="创建时间" width="180">
            <template #default="{ row }">
              {{ formatTime(row.createdAt) }}
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="140">
            <template #default="{ row }">
              <el-tag
                :type="
                  row.status === 'NEW'
                    ? 'info'
                    : row.status === 'IN_TRANSIT'
                    ? 'primary'
                    : row.status === 'DELIVERED'
                    ? 'success'
                    : row.status === 'CANCELLED'
                    ? 'info'
                    : ''
                "
                effect="plain"
                >{{ statusLabelMap[row.status] || row.status }}</el-tag
              >
            </template>
          </el-table-column>
          <el-table-column label="操作" width="220" fixed="right">
            <template #default="{ row }">
              <el-button
                v-if="
                  !canManageRequest &&
                  !isDispatcher &&
                  row.status !== 'DELIVERED' &&
                  row.status !== 'CANCELLED'
                "
                link
                type="primary"
                @click="openUpdateShipmentStatus(row)"
                >更新状态</el-button
              >
              <el-button
                v-if="isDispatcher && row.status === 'IN_TRANSIT'"
                link
                type="success"
                @click="completeShipment(row)"
                >确认收到</el-button
              >
              <el-button
                v-if="isDispatcher && row.status === 'IN_TRANSIT'"
                link
                type="danger"
                @click="rejectShipment(row)"
                >拒绝签收</el-button
              >
              <el-button
                link
                type="primary"
                @click="
                  async () => {
                    selectedShipmentId = row.shipmentId;
                    activeTab = 'map';
                    await loadSelectedShipment();
                    await renderShipmentOnMap();
                  }
                "
              >
                查看轨迹
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <div class="pagination">
        <el-pagination
          background
          layout="prev, pager, next"
          :page-size="shipmentQuery.size"
          :current-page="shipmentQuery.page"
          :total="shipmentsTotal"
          @current-change="(p:number) => { shipmentQuery.page = p; loadShipments() }"
        />
      </div>
    </div>

    <div v-else class="map-view">
      <div class="map-toolbar">
        <el-input-number
          v-model="selectedShipmentId"
          :min="1"
          controls-position="right"
          placeholder="运输ID"
        />
        <el-button
          type="primary"
          @click="
            async () => {
              await loadSelectedShipment();
              await renderShipmentOnMap();
            }
          "
          :disabled="!selectedShipmentId"
        >
          加载并渲染
        </el-button>
      </div>

      <div class="map-body">
        <div ref="mapContainer" class="map-container"></div>
        <div class="map-overlay glass-panel">
          <h3>轨迹信息</h3>
          <p v-if="selectedShipment">
            运输任务：#{{ selectedShipment.shipmentId }}
          </p>
          <p v-if="selectedShipment && drivingSummary">
            驾车规划：{{ formatDistanceMeters(drivingSummary.distance) }} ·
            {{ formatDurationSeconds(drivingSummary.time) }}
          </p>
          <p v-if="!selectedShipment">请选择运输任务</p>
          <div v-if="selectedShipment" class="tracking-list">
            <div
              v-for="(t, idx) in trackingDisplay"
              :key="idx"
              class="tracking-item"
            >
              <div class="tracking-title">
                {{ statusLabelMap[t.status] || t.status }}
              </div>
              <div class="tracking-sub">
                {{ t.location || "-" }} · {{ t.timestamp || "-" }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <el-dialog
      v-model="requestDialogVisible"
      :title="requestDialogMode === 'create' ? '新建需求单' : '查看/更新需求单'"
      width="720px"
    >
      <el-form label-width="90px">
        <el-form-item label="标题" v-if="requestDialogMode === 'create'">
          <el-input v-model="requestForm.title" placeholder="请输入需求标题" />
        </el-form-item>
        <el-form-item label="地点" v-if="requestDialogMode === 'create'">
          <el-input
            v-model="requestForm.location"
            placeholder="请输入需求地点"
          />
        </el-form-item>

        <el-form-item label="状态" v-if="requestDialogMode === 'edit'">
          <el-select
            v-model="requestForm.status"
            placeholder="选择状态"
            clearable
            style="width: 220px"
          >
            <el-option label="待处理 (PENDING)" value="PENDING" />
            <el-option label="已指派 (ASSIGNED)" value="ASSIGNED" />
            <el-option label="已完成 (COMPLETED)" value="COMPLETED" />
            <el-option label="已取消 (CANCELLED)" value="CANCELLED" />
          </el-select>
        </el-form-item>
        <el-form-item label="指派给" v-if="requestDialogMode === 'edit'">
          <el-input-number
            v-model="requestForm.assignedTo"
            :min="1"
            controls-position="right"
            placeholder="用户ID"
          />
        </el-form-item>

        <el-form-item label="物资明细" v-if="requestDialogMode === 'create'">
          <div class="items-editor">
            <div
              v-for="(it, idx) in requestForm.items"
              :key="idx"
              class="items-row"
            >
              <el-select
                v-model="it.itemId"
                placeholder="选择物资"
                filterable
                style="width: 360px"
              >
                <el-option
                  v-for="opt in items"
                  :key="opt.id"
                  :label="`${opt.name}（${opt.unit}）`"
                  :value="opt.id"
                />
              </el-select>
              <el-input-number
                v-model="it.quantity"
                :min="1"
                controls-position="right"
              />
              <el-button
                link
                type="danger"
                :disabled="requestForm.items.length <= 1"
                @click="requestForm.items.splice(idx, 1)"
                >移除</el-button
              >
            </div>
            <el-button
              link
              type="primary"
              @click="requestForm.items.push({ itemId: null, quantity: 1 })"
              >添加一行</el-button
            >
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="requestDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveRequest">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="shipmentDialogVisible"
      :title="shipmentDialogMode === 'create' ? '新建运输任务' : '更新运输状态'"
      width="760px"
    >
      <el-form label-width="110px">
        <template v-if="shipmentDialogMode === 'create'">
          <el-form-item label="需求单ID">
            <el-input-number
              v-model="shipmentForm.requestId"
              :min="1"
              controls-position="right"
            />
            <el-button
              link
              type="primary"
              @click="loadRequestForShipment"
              style="margin-left: 8px"
              >加载需求明细</el-button
            >
          </el-form-item>
          <el-form-item label="出发仓库ID">
            <el-input-number
              v-model="shipmentForm.fromWarehouseId"
              :min="1"
              controls-position="right"
            />
          </el-form-item>
          <el-form-item label="目的地">
            <el-input
              v-model="shipmentForm.toLocation"
              placeholder="请输入目的地"
            />
          </el-form-item>
          <el-form-item label="物资明细">
            <div class="items-editor">
              <div
                v-for="(it, idx) in shipmentForm.items"
                :key="idx"
                class="items-row"
              >
                <el-select
                  v-model="it.itemId"
                  placeholder="选择物资"
                  filterable
                  style="width: 360px"
                >
                  <el-option
                    v-for="opt in items"
                    :key="opt.id"
                    :label="`${opt.name}（${opt.unit}）`"
                    :value="opt.id"
                  />
                </el-select>
                <el-input-number
                  v-model="it.quantity"
                  :min="1"
                  controls-position="right"
                />
                <el-button
                  link
                  type="danger"
                  :disabled="shipmentForm.items.length <= 1"
                  @click="shipmentForm.items.splice(idx, 1)"
                  >移除</el-button
                >
              </div>
              <el-button
                link
                type="primary"
                @click="shipmentForm.items.push({ itemId: null, quantity: 1 })"
                >添加一行</el-button
              >
            </div>
          </el-form-item>
        </template>
        <template v-else>
          <el-form-item label="状态">
            <el-select
              v-model="shipmentForm.status"
              placeholder="选择状态"
              style="width: 240px"
            >
              <el-option label="新建 (NEW)" value="NEW" />
              <el-option label="运输中 (IN_TRANSIT)" value="IN_TRANSIT" />
              <el-option label="已送达 (DELIVERED)" value="DELIVERED" />
              <el-option label="已取消 (CANCELLED)" value="CANCELLED" />
            </el-select>
          </el-form-item>
          <el-form-item label="当前位置">
            <el-input
              v-model="shipmentForm.location"
              placeholder="可选：例如 成都/西安/北京"
            />
          </el-form-item>
          <el-form-item label="时间戳">
            <el-input
              v-model="shipmentForm.timestamp"
              placeholder="可选：ISO8601，例如 2025-12-01T14:30:00Z"
            />
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="shipmentDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveShipment">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.scheduling-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 120px);
  gap: 20px;
}

.tabs-header {
  display: flex;
  gap: 4px;
  background: #0d1117;
  padding: 4px;
  border-radius: 8px;
  width: fit-content;
}

.tab-item {
  padding: 8px 16px;
  display: flex;
  align-items: center;
  gap: 8px;
  border-radius: 6px;
  cursor: pointer;
  color: #8b949e;
  transition: all 0.2s;
}

.tab-item:hover {
  color: #e6edf3;
}
.tab-item.active {
  background: #21262d;
  color: #58a6ff;
}

.panel {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.table-card {
  border: 1px solid #30363d;
  background: #161b22;
}

.pagination {
  display: flex;
  justify-content: flex-end;
}

.map-view {
  flex: 1;
  position: relative;
  border-radius: 12px;
  overflow: hidden;
  border: 1px solid #30363d;
}

.map-toolbar {
  position: absolute;
  top: 16px;
  left: 16px;
  z-index: 5;
  display: flex;
  gap: 10px;
  align-items: center;
}

.map-body {
  width: 100%;
  height: 100%;
}

.map-container {
  width: 100%;
  height: 100%;
}

.map-overlay {
  position: absolute;
  top: 20px;
  right: 20px;
  width: 320px;
  max-height: calc(100% - 40px);
  overflow: auto;
  padding: 20px;
  border-radius: 12px;
  color: white;
}

.map-overlay h3 {
  margin-top: 0;
  font-size: 1.1rem;
}
.map-overlay p {
  margin: 8px 0;
  color: #8b949e;
}

.tracking-list {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.tracking-item {
  padding: 10px 12px;
  border-radius: 10px;
  background: rgba(13, 17, 23, 0.45);
  border: 1px solid rgba(48, 54, 61, 0.8);
}

.tracking-title {
  font-weight: 600;
  color: #e6edf3;
}

.tracking-sub {
  margin-top: 4px;
  font-size: 0.85rem;
  color: #8b949e;
}

.items-editor {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.items-row {
  display: flex;
  align-items: center;
  gap: 10px;
}
</style>
