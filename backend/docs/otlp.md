# 后端系统集成 OpenTelemetry（OTLP / Collector / Jaeger / Prometheus / Elasticsearch）

本文档用于说明如何在本仓库后端系统中引入 OpenTelemetry，实现 traces、metrics、logs 的采集、处理、转发与可视化，目标是做到：

- 应用统一用 OTLP 上报（gRPC/HTTP）到 Collector
- Collector 作为中心枢纽（hub），进行处理与分流（fan-out）
- traces 可在 Jaeger 查询，并由 Elasticsearch 落盘保存
- metrics 以 Prometheus 方式暴露并由 Prometheus 抓取，Grafana 展示
- logs（可选）落地到 Elasticsearch，Grafana/Discover 查询

本仓库当前已有：

- Jaeger（all-in-one）容器定义：[deploy/docker-compose.yaml](file:///root/ems/deploy/docker-compose.yaml)
- Prometheus 容器定义与配置示例：[prometheus.yaml](file:///root/ems/backend/internal/common/observation/metrics/prometheus.yaml)
- Go 后端采用 Gin（REST）与 gRPC（内部调用），适合用 OpenTelemetry 的 Gin/gRPC 插件做自动埋点

---

## 1）架构总览

### 1.1 OpenTelemetry 组件与职责

- OpenTelemetry SDK（应用端）
  - 负责在代码中创建 spans/metrics/logs（或通过框架/库自动埋点）
  - 负责将遥测数据通过 OTLP exporter 发送到 Collector
  - 负责上下文传播（Propagation）：将 trace context 在 HTTP headers / gRPC metadata 中透传，形成跨服务一条链

- OpenTelemetry Collector（平台端）
  - receivers：接收应用上报的数据（OTLP gRPC/HTTP、Prometheus scrape 等）
  - processors：在内存中处理数据（批处理、限流、采样、属性增删改、补充资源信息等）
  - exporters：将数据导出到后端（Jaeger、Prometheus、Elasticsearch、OTLP、Kafka…）
  - pipelines：按 data type（traces/metrics/logs）分别组装 receiver+processor+exporter

- 后端存储与查询
  - Trace 查询与 UI：Jaeger（Query + UI）
  - Trace 存储：Elasticsearch（作为 Jaeger 的 span storage）
  - Metric 存储：Prometheus（抓取 Collector 暴露的 metrics endpoint）
  - Log 存储（可选）：Elasticsearch（OTLP logs 或日志采集器上送）

### 1.2 以 Collector 为中心枢纽的“接收-处理-分发”

推荐把 Collector 放在所有服务旁边（同一集群/同一 docker-compose 网络），所有服务仅配置一个 OTLP endpoint：

- 应用（SDK）→ Collector（OTLP）
- Collector 根据 pipeline 配置把 traces 分发到多个后端（Jaeger/ES），把 metrics 暴露给 Prometheus，把 logs 写入 Elasticsearch

优点：

- 应用侧配置极简：只要 OTLP endpoint + 资源属性 + 采样策略
- 后端可演进：随时替换/新增 exporter，不动应用
- 统一处理：batch、memory limiter、重试、队列、脱敏、采样等集中在 Collector

### 1.3 Trace / Metric / Log 的整体数据流（推荐形态）

**Trace（Jaeger + Elasticsearch 落盘）**

1. 应用 SDK 产生 spans → OTLP exporter 上报到 Collector
2. Collector processors 做批处理、属性治理、采样等
3. Collector exporter 把 traces 导出到 Jaeger Collector（或直接 OTLP 到 Jaeger OTLP 端口）
4. Jaeger 将 spans 写入 Elasticsearch（SPAN_STORAGE_TYPE=elasticsearch）
5. Grafana 或 Jaeger UI 通过 Jaeger Query 检索 traces

**Metric（Prometheus）**

两种常见方式（二选一或混用）：

- 方式 A：应用用 OTLP 导出 metrics → Collector → Prometheus exporter 暴露 `/metrics` → Prometheus 抓取
- 方式 B：应用直接暴露 `/metrics`（Prometheus client）→ Prometheus 抓取；Collector 主要负责 traces/logs

本文推荐方式 A，便于跨语言统一与与 traces/资源属性对齐。

**Log（Elasticsearch，可选）**

1. 应用产生结构化日志 → OTLP logs exporter → Collector
2. Collector 将 logs 写入 Elasticsearch（按索引策略滚动）
3. Grafana/Discover 查询日志，并通过 trace_id/span_id 与 traces 关联

---

## 2）Collector 配置

### 2.1 完整示例：otel-collector.yaml

下面给出一个“落地可用”的 Collector 配置示例，支持：

- receivers：OTLP（gRPC 4317 / HTTP 4318）
- traces：导出到 Jaeger（通过 OTLP）并通过 Jaeger 使用 Elasticsearch 存储
- metrics：导出为 Prometheus scrape endpoint（默认 9464）
- logs：导出到 Elasticsearch（可选）

```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  memory_limiter:
    check_interval: 1s
    limit_mib: 512
    spike_limit_mib: 128

  batch:
    send_batch_size: 8192
    send_batch_max_size: 16384
    timeout: 2s

  resource:
    attributes:
      - key: service.namespace
        value: ems
        action: upsert
      - key: deployment.environment
        value: dev
        action: upsert

  attributes/drop_sensitive:
    actions:
      - key: http.request.header.authorization
        action: delete
      - key: http.request.header.cookie
        action: delete

exporters:
  otlp/jaeger:
    endpoint: jaeger:4317
    tls:
      insecure: true
    sending_queue:
      enabled: true
      num_consumers: 4
      queue_size: 2048
    retry_on_failure:
      enabled: true
      initial_interval: 1s
      max_interval: 30s
      max_elapsed_time: 5m

  prometheus:
    endpoint: 0.0.0.0:9464
    namespace: ems
    send_timestamps: true
    enable_open_metrics: true
    resource_to_telemetry_conversion:
      enabled: true

  elasticsearch/logs:
    endpoints: [ "http://elasticsearch:9200" ]
    logs_index: "otel-logs-%{+yyyy.MM.dd}"
    pipeline: ""
    timeout: 10s
    retry:
      enabled: true
      max_requests: 5
      initial_interval: 1s
      max_interval: 30s
    sending_queue:
      enabled: true
      queue_size: 4096

service:
  telemetry:
    logs:
      level: info
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, resource, attributes/drop_sensitive, batch]
      exporters: [otlp/jaeger]

    metrics:
      receivers: [otlp]
      processors: [memory_limiter, resource, batch]
      exporters: [prometheus]

    logs:
      receivers: [otlp]
      processors: [memory_limiter, resource, batch]
      exporters: [elasticsearch/logs]
```

### 2.2 “Trace 同时到 Jaeger 和 Elasticsearch”的解释方式（推荐实现）

生产场景更推荐让 **Jaeger 使用 Elasticsearch 作为 trace 存储**，这样：

- traces 在 Jaeger UI 可查（查询入口）
- traces 实际落盘在 Elasticsearch（存储后端）

也就是说：**Collector 只需要把 traces 导出到 Jaeger**；“落 ES”由 Jaeger 的存储配置完成，避免 Collector 维护 span 存储格式与索引映射。

如果你确实需要“Collector 直接写 spans 到 Elasticsearch（独立索引）”，建议：

- 引入 Elastic APM Server（Collector OTLP → APM Server → ES），或
- 自行在 ES 维护 trace 索引与 mapping（复杂度显著更高，不建议作为首选）

### 2.3 配置项作用与调优建议

- memory_limiter
  - 作用：限制 Collector 内存占用，避免 OOM 被系统杀死导致数据“全丢”
  - 建议：limit_mib 与容器内存配额配合设置；对突发流量配 spike_limit_mib

- batch
  - 作用：减少 exporter 调用次数、提升吞吐、降低后端压力
  - 建议：
    - `send_batch_size`：吞吐优先可增大（如 8192/16384）；低延迟优先可减小
    - `timeout`：通常 1~5s，过小会增加请求数，过大增加尾延迟

- sending_queue + retry_on_failure（exporters）
  - 作用：后端短暂不可用时“先排队再重试”，降低数据丢失
  - 建议：
    - queue_size 根据峰值流量与可接受的缓冲时长配置
    - max_elapsed_time 控制最长重试窗口，避免无限重试拖垮内存

- attributes/resource processors
  - 作用：统一补齐或治理属性（脱敏、规范化、打环境标签）
  - 建议：在 Collector 做“平台级”一致性治理，避免每个服务各写一套

---

## 3）应用端 SDK 配置

### 3.1 通用推荐（适用于所有语言）

- 统一 OTLP endpoint 指向 Collector（例如 `otel-collector:4317`）
- 统一资源属性（resource attributes）：
  - `service.name`：服务名（auth/user/warehouse/scheduling/statistics）
  - `service.namespace`：系统域（ems）
  - `deployment.environment`：dev/staging/prod
  - `service.version`：版本号（git sha / tag）
- 传播协议建议：W3C TraceContext（默认）+ Baggage；需要兼容旧系统可加 B3

### 3.2 Go：SDK 初始化（traces + metrics，OTLP gRPC）

下面示例展示如何在 Go 服务中初始化 TracerProvider 与 MeterProvider，并把 traces/metrics 导出到 Collector：

```go
package otelinit

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Shutdown func(ctx context.Context) error

func Init(ctx context.Context, collectorEndpoint, serviceName, env, version string) (Shutdown, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			attribute.String("service.namespace", "ems"),
			attribute.String("deployment.environment", env),
			attribute.String("service.version", version),
		),
	)
	if err != nil {
		return nil, err
	}

	traceExp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(collectorEndpoint),
		otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1))),
	)

	metricExp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(collectorEndpoint),
		otlpmetricgrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		_ = tp.Shutdown(ctx)
		return nil, err
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp, sdkmetric.WithInterval(10*time.Second))),
	)

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_ = mp.Shutdown(ctx)
		return tp.Shutdown(ctx)
	}, nil
}
```

**Gin / gRPC 自动埋点建议**

- Gin：使用 `otelgin` 中间件，把入站 HTTP 自动生成 span，并把 trace context 注入 `context.Context`
- gRPC：使用 `otelgrpc` 的 stats handler 或 interceptor，把 client/server 的 RPC 自动生成 span

### 3.3 Java：SDK 初始化（traces + metrics，OTLP gRPC）

Java 推荐优先使用 OpenTelemetry Java Agent（零侵入），但此处给出“代码初始化”的可控版本：

```java
import io.opentelemetry.api.GlobalOpenTelemetry;
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.common.Attributes;
import io.opentelemetry.api.trace.propagation.W3CTraceContextPropagator;
import io.opentelemetry.context.propagation.ContextPropagators;
import io.opentelemetry.exporter.otlp.metrics.OtlpGrpcMetricExporter;
import io.opentelemetry.exporter.otlp.trace.OtlpGrpcSpanExporter;
import io.opentelemetry.sdk.OpenTelemetrySdk;
import io.opentelemetry.sdk.metrics.SdkMeterProvider;
import io.opentelemetry.sdk.metrics.export.PeriodicMetricReader;
import io.opentelemetry.sdk.resources.Resource;
import io.opentelemetry.sdk.trace.SdkTracerProvider;
import io.opentelemetry.sdk.trace.export.BatchSpanProcessor;
import io.opentelemetry.semconv.ResourceAttributes;

import java.time.Duration;

public class OtelInit {
  public static OpenTelemetry init(String collectorEndpoint, String serviceName, String env, String version) {
    Resource resource = Resource.getDefault().merge(
      Resource.create(Attributes.of(
        ResourceAttributes.SERVICE_NAME, serviceName,
        ResourceAttributes.SERVICE_NAMESPACE, "ems",
        ResourceAttributes.DEPLOYMENT_ENVIRONMENT, env,
        ResourceAttributes.SERVICE_VERSION, version
      ))
    );

    OtlpGrpcSpanExporter spanExporter = OtlpGrpcSpanExporter.builder()
      .setEndpoint("http://" + collectorEndpoint)
      .build();

    SdkTracerProvider tracerProvider = SdkTracerProvider.builder()
      .setResource(resource)
      .addSpanProcessor(
        BatchSpanProcessor.builder(spanExporter)
          .setScheduleDelay(Duration.ofMillis(800))
          .build()
      )
      .build();

    OtlpGrpcMetricExporter metricExporter = OtlpGrpcMetricExporter.builder()
      .setEndpoint("http://" + collectorEndpoint)
      .build();

    SdkMeterProvider meterProvider = SdkMeterProvider.builder()
      .setResource(resource)
      .registerMetricReader(
        PeriodicMetricReader.builder(metricExporter)
          .setInterval(Duration.ofSeconds(10))
          .build()
      )
      .build();

    OpenTelemetrySdk sdk = OpenTelemetrySdk.builder()
      .setTracerProvider(tracerProvider)
      .setMeterProvider(meterProvider)
      .setPropagators(ContextPropagators.create(W3CTraceContextPropagator.getInstance()))
      .build();

    GlobalOpenTelemetry.set(sdk);
    return sdk;
  }
}
```

### 3.4 导出到 Collector 的端点与资源属性建议

推荐固定 Collector：

- OTLP gRPC：`otel-collector:4317`
- OTLP HTTP：`otel-collector:4318`

资源属性建议最少包含：

- `service.name`：必填（Grafana/Jaeger 的服务维度依赖它）
- `deployment.environment`：区分环境（dev/staging/prod）
- `service.version`：版本追溯（与发布系统打通）

Prometheus 侧建议打开 `resource_to_telemetry_conversion`，让 `service.name` 等资源信息变成 labels，便于按服务聚合指标。

---

## 4）后端存储与服务设置

### 4.1 部署 Jaeger 并连接 Elasticsearch 作为存储后端

下面给出一个示例 docker-compose 片段，展示：

- Elasticsearch 作为 Jaeger span storage
- Jaeger 提供 UI 查询与 OTLP 接收端口

```yaml
services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.15.2
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - ES_JAVA_OPTS=-Xms1g -Xmx1g
    ports:
      - "9200:9200"

  jaeger:
    image: jaegertracing/all-in-one:latest
    environment:
      - COLLECTOR_OTLP_ENABLED=true
      - SPAN_STORAGE_TYPE=elasticsearch
      - ES_SERVER_URLS=http://elasticsearch:9200
      - ES_NUM_SHARDS=1
      - ES_NUM_REPLICAS=0
    ports:
      - "16686:16686"  # UI
      - "4317:4317"    # OTLP gRPC
      - "4318:4318"    # OTLP HTTP
```

说明：

- 生产建议为 Elasticsearch 配置持久化卷、资源限制、ILM 等
- Jaeger 依赖 ES 可用后才能正常落盘；Collector exporter 侧应开启队列与重试

### 4.2 Prometheus scrape 配置示例（抓取 Collector metrics）

Collector 中 Prometheus exporter 默认暴露在 `:9464/metrics`，Prometheus 需要配置抓取目标：

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: otel-collector
    scrape_interval: 10s
    metrics_path: /metrics
    static_configs:
      - targets: ["otel-collector:9464"]
```

如果你同时需要抓取应用自己暴露的 `/metrics`（非 OTLP metrics），可在 scrape_configs 中加对应 targets。

### 4.3 Elasticsearch 保存日志/指标的索引策略建议（可选）

如果使用 Elasticsearch 保存 OTLP logs（推荐），建议：

- 索引命名：按天滚动，例如 `otel-logs-YYYY.MM.DD`
- 时间字段：统一使用 `@timestamp` 或 `timestamp`，并确保 mapping 为 date
- 保留策略：使用 ILM（Hot/Warm/Delete），按环境/服务拆分 index pattern
- 脱敏：在 Collector processor 侧删除敏感 headers/token，避免写入 ES

---

## 5）Grafana 可视化

### 5.1 配置数据源

- Prometheus
  - URL：`http://prometheus:9090`
  - 开启 Exemplars（如果你的指标链路支持 exemplar trace_id，可实现指标跳 trace）

- Jaeger
  - URL：`http://jaeger:16686`
  - 数据源类型：Jaeger

- Elasticsearch（logs）
  - URL：`http://elasticsearch:9200`
  - Index pattern：`otel-logs-*`
  - Time field：`@timestamp`（或你写入的时间字段）

### 5.2 常用查询示例

PromQL（示例，具体指标名取决于你的应用/采集方式）：

- 请求量（按服务聚合）
  - `sum(rate(http_server_requests_total[5m])) by (service_name)`
- P95 延迟
  - `histogram_quantile(0.95, sum(rate(http_server_request_duration_seconds_bucket[5m])) by (le, service_name))`
- gRPC 错误率
  - `sum(rate(grpc_server_handled_total{grpc_code!="OK"}[5m])) by (service_name, grpc_method)`

Jaeger Trace 搜索：

- 按服务：service = `auth`
- 按 operation：operation = `POST /api/v1/auth/login`（取决于你在埋点中设置的 span name）
- 按 tag：例如 `http.status_code=500`、`rpc.grpc.status_code=13`

### 5.3 交叉关联（metrics → trace → logs）

推荐逐步实现“三件套关联”：

1. 指标到 Trace（Exemplars）
   - 在 Collector 中可使用 span→metrics 方案（例如 spanmetrics connector）或在应用侧把 trace_id 作为 exemplar 附在 metrics 上
   - Grafana 的 Prometheus datasource 支持配置 exemplar trace link，跳转到 Jaeger

2. Trace 到日志
   - 日志结构化字段包含 `trace_id`、`span_id`
   - 在 Grafana Explore 中可通过 trace_id 过滤 ES 日志

3. 统一命名与资源属性
   - `service.name`、`deployment.environment` 在 metrics/traces/logs 中一致，保证跨维度过滤结果一致

---

## 6）最佳实践与注意事项

### 6.1 采样策略建议

- 开发环境：可 100% 采样，便于调试
- 生产环境：
  - 入口服务使用 ParentBased + TraceIDRatioBased（例如 1%~10%）
  - 对错误请求与慢请求尽量保留（可在 Collector 使用 tail sampling：按 status_code、latency 等条件采样）

控制 Span 数量的建议：

- 限制高频循环/批处理内部的 span 创建
- 对低价值 span 使用更粗粒度命名或合并
- 避免把大 payload（body/SQL 参数）作为属性写入（会导致 UI 卡顿与存储膨胀）

### 6.2 标签（Labels/Attributes）命名规范

- Prometheus labels 建议：
  - 使用 snake_case：`service_name`, `deployment_environment`
  - 控制 label 基数（cardinality），避免把 user_id、request_id 这类高基数字段作为 label

- Trace attributes 建议：
  - 优先遵循语义约定（semantic conventions），例如：
    - HTTP：`http.method`, `http.route`, `http.status_code`
    - gRPC：`rpc.system=grpc`, `rpc.service`, `rpc.method`
  - 高基数字段（user_id、订单号）可以作为 trace attribute（用于排障），但要谨慎数量与脱敏

### 6.3 数据一致性（Grafana 图形/表格展示）

- 指标统计口径与 trace 采样会影响一致性：
  - 指标通常是全量（不采样）更可靠
  - trace 可能采样；因此“trace 数量”不等于“请求数量”
- 建议把“业务 KPI 指标”与“链路排障 trace”区分用途，避免误解

---

## 7）故障排查与调试

### 7.1 验证 Collector 是否正确接收数据

- 检查 Collector 日志是否有 receiver/exporter 报错（连接失败、队列满、重试耗尽）
- 确认端口可达：
  - OTLP gRPC：4317
  - OTLP HTTP：4318
  - Prometheus metrics：9464
- Prometheus Targets 页面应能看到 `otel-collector` 为 UP

### 7.2 常见连通性问题

- 应用上报 endpoint 写错
  - 容器内用服务名（`otel-collector:4317`），本机直跑用 `localhost:4317`
- Jaeger/Elasticsearch 未就绪
  - exporter 会重试并堆积队列；队列满可能导致丢数据
  - 建议开启 sending_queue + retry_on_failure，并设置合理 max_elapsed_time
- Elasticsearch 写入失败（mapping/权限/磁盘）
  - 检查 ES 健康：`/_cluster/health`
  - 检查索引是否创建与写入是否被拒绝（磁盘水位、只读索引等）

### 7.3 Jaeger / Grafana 常见查询问题

- Jaeger 查不到服务名
  - 检查应用是否设置 `service.name`（必填）
  - 检查 Collector traces pipeline 是否正确（receiver/processor/exporter）
  - 检查 Jaeger 是否落盘 ES（SPAN_STORAGE_TYPE 与 ES_SERVER_URLS）

- Grafana 查不到指标或维度不全
  - 检查 Prometheus 是否抓到 `otel-collector:9464/metrics`
  - 检查 `resource_to_telemetry_conversion.enabled` 是否开启（决定资源属性是否变 labels）
  - 检查 label 基数是否过高导致查询变慢/爆炸

---

## 附：与本仓库的落地建议

- 增加一个 otel-collector 容器（与后端服务同网络），统一接收 OTLP
- 后端 Go 服务引入 Gin/gRPC 自动埋点（入站与出站），并确保 context 传播
- 将 Jaeger 改为 Elasticsearch 存储后端（或新增 ES + Jaeger collector/query 组件）
- Prometheus 抓取 Collector 的 metrics endpoint，并在 Grafana 关联到 Jaeger datasource
