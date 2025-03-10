# APScheduler任务调度管理系统

## 1. 项目概述

APScheduler任务调度管理系统是一个基于gin构建的高效任务调度与管理平台。该系统采用领域驱动设计(DDD)架构，通过依赖注入(DI)实现模块间的低耦合，具备任务创建、调度、执行、结果处理与外部系统集成等完整功能。

### 核心特性

- **灵活的任务调度**：支持即时任务(IMMEDIATE)和基于cron表达式的定时任务(SCHEDULED)
- **强大的任务管理**：包含优先级队列、依赖管理、超时处理、重试机制
- **外部系统集成**：支持与Jira、Confluence等系统对接，实现数据处理自动化
- **结果汇总与报告**：任务执行结果可自动汇总并更新至Confluence页面
- **RESTful API接口**：提供完整的任务管理接口，支持查询、过滤、监控
- **高可扩展性**：模块化设计，便于扩展新的任务类型和外部系统集成

## 2. 系统架构

系统基于清晰的分层架构设计：

### 架构分层

1. **表示层**（gin应用）
   - 提供RESTful API接口，处理用户交互

2. **应用层**（Application）
   - 编排核心业务流程
   - 协调领域服务与基础设施
   - 任务调度服务

3. **领域层**（Domain）
   - 核心业务规则与实体模型
   - 领域服务（Jira处理、Confluence处理等）

4. **基础设施层**（Infrastructure）
   - 外部系统集成（Jira API、Confluence API）
   - 配置管理、日志服务
   - 持久化实现

### 核心组件

- **任务调度器**（SchedulerService）：系统核心，协调各管理器完成任务调度
- **任务执行器**（TaskExecutor）：执行具体任务，调用相应领域服务
- **任务仓库**（TaskRepository）：管理任务的存储和检索
- **结果报告服务**（ResultReportingService）：处理任务结果并生成报告

## 3. 功能详解

### 3.1 任务调度与执行

#### 任务类型

- **定时任务（SCHEDULED）**：通过cron表达式配置执行时间
- **即时任务（IMMEDIATE）**：创建后立即加入执行队列

#### 任务状态流转

```
PENDING -> QUEUED -> RUNNING -> DONE/FAILED
             ^                    |
             |                    v
             +---- RETRY <---- TIMEOUT
```

#### 任务优先级

- HIGH：高优先级任务，优先执行
- MEDIUM：中等优先级（默认）
- LOW：低优先级任务

#### 任务依赖管理

- 支持任务间依赖关系配置
- 依赖任务完成后，才会执行后续任务

#### 任务超时与重试

- 支持配置任务执行超时时间
- 支持自定义重试策略（最大重试次数、重试延迟、退避因子）

### 3.2 外部系统集成

#### Jira集成

- 支持根据项目或根问题提取Issue信息
- 支持Issue状态更新、字段修改
- 支持生成Excel格式报表导出

#### Confluence集成

- 支持任务结果汇总至Confluence页面
- 支持表格形式更新与展示
- 支持自动创建/更新内容页面

#### 未来可扩展集成

- Mattermost通知（已有基础实现）
- 邮件通知
- 其他第三方系统API

### 3.3 API接口

| 端点 | 方法 | 描述 |
|------|------|------|
| `/tasks` | GET | 获取所有任务列表 |
| `/tasks/status/{status}` | GET | 根据状态过滤任务 |
| `/task_history` | GET | 获取已执行完成的任务历史 |

## 4. 技术规范

### 4.1 开发环境

- **Golang版本**：1.23.1
- **主要依赖**：
  - cron 0.115.8+

### 4.2 代码组织结构

```

```

### 4.3 数据模型

#### Task模型

```python
class Task(BaseModel):
    id: UUID                            # 任务唯一标识
    name: str                           # 任务名称
    task_type: TaskType                 # 任务类型(SCHEDULED/IMMEDIATE)
    cron_expr: Optional[str]            # cron表达式(定时任务)
    status: TaskStatus                  # 当前状态
    created_at: datetime                # 创建时间
    updated_at: datetime                # 更新时间
    priority: TaskPriority              # 优先级(HIGH/MEDIUM/LOW)
    metadata: Dict[str, Any]            # 元数据
    tags: List[str]                     # 标签列表
    owner: Optional[str]                # 创建者
    dependencies: List[UUID]            # 依赖任务ID列表
    timeout_seconds: Optional[int]      # 超时时间(秒)
    retry_policy: Optional[RetryPolicy] # 重试策略
    parameters: Dict[str, Any]          # 任务参数
```

### 4.4 配置项说明

config.yaml文件包含以下配置项：

#### Confluence配置

```yaml
confluence:
  url: "https://confluence.example.com"
  username: "confluence_user"
  password: "confluence_password"
  main_page_id: "123456"
  task_result_page_id: "789012"
```

#### Jira配置

```yaml
jira:
  url: "https://jira.example.com"
  username: "jira_user"
  password: "jira_password"
```

#### 调度器配置

```yaml
scheduler:
  poll_interval: 30        # 轮询间隔(秒)
  concurrency: 5           # 最大并发任务数
  coalesce: false          # 是否合并延迟任务
  max_instances: 5         # 最大实例数
```

#### 日志配置

```yaml
log:
  level: INFO              # 日志级别
  filename: logs/app.log   # 日志文件路径
  max_bytes: 10485760      # 单个日志文件大小上限(10MB)
  backup_count: 5          # 保留日志文件数量
  format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
```

#### 存储配置

```yaml
storage:
  path: "task_storage"     # 任务存储路径
```

#### 报告配置

```yaml
reporting:
  interval: 30             # 报告生成间隔(秒)
  report_types:
    - "confluence"
    - "mattermost"
  confluence:
    page_id: "789012"
    template: "task_results_template"
```

## 5. 安装与部署

### 5.1 环境准备


### 5.2 配置修改

1. 复制并修改配置文件
   ```bash
   cp config.yaml.example config.yaml
   # 编辑config.yaml，填入实际环境的配置信息
   ```

2. 创建必要的目录
   ```bash
   mkdir -p logs task_storage jira_reports
   ```

### 5.3 运行应用


### 5.4 访问API

应用启动后，可通过以下URL访问：

- API文档：http://localhost:8000/docs
- 任务列表：http://localhost:8000/tasks
- 任务历史：http://localhost:8000/task_history

## 6. 使用指南

### 6.1 创建任务示例

在程序中中已包含两个示例任务：

```python
# JIRA根问题提取任务
root_ticket_task = {
    "name": "JIRA Extraction - Root Ticket",
    "task_type": TaskType.IMMEDIATE,
    "tags": ["JIRA_TASK_EXP"],
    "parameters": {
        "jira_envs": ["env1.jira.com", "env2.jira.com"],
        "key_type": "root_ticket",
        "key_value": "PROJ-123",
        "user": "johndoe"
    }
}

# JIRA项目提取任务(每日执行)
project_task = {
    "name": "JIRA Extraction - Project",
    "task_type": TaskType.IMMEDIATE,
    "cron_expr": "0 0 * * *",  # 每天午夜执行
    "tags": ["JIRA_TASK_EXP"],
    "parameters": {
        "jira_envs": ["env1.jira.com"],
        "key_type": "project",
        "key_value": "PROJ",
        "user": "johndoe"
    }
}
```

### 6.2 任务执行流程

1. 任务创建并存入TaskRepository
2. SchedulerService周期性轮询待执行任务
3. 根据任务类型(即时/定时)加入执行队列或创建定时作业
4. TaskExecutor执行任务，调用相应的领域服务
5. 执行结果存入TaskResultRepository
6. ResultReportingService定期汇总结果并更新到Confluence

### 6.3 任务参数说明

任务参数(parameters)根据任务标签和类型而异：

#### JIRA_TASK_EXP标签任务

```json
{
  "jira_envs": ["jira环境URL列表"],
  "key_type": "root_ticket或project",
  "key_value": "问题键值或项目键值",
  "user": "执行用户"
}
```

## 7. 扩展与定制

### 7.1 添加新的任务处理器

1. 在domain/services/下创建新的领域服务
2. 在application/use_cases/executor.py中注册新的处理器
3. 参考现有标签(如JIRA_TASK_EXP)添加新的任务标签常量

### 7.2 集成新的外部系统

1. 在infrastructure/integration/下创建新的服务连接器
2. 在domain/services/下创建对应的领域服务
3. 在application/di_container.py中注册新的服务

### 7.3 自定义结果报告

1. 修改application/services/result_reporting_service.py
2. 添加新的报告类型和模板

## 8. 常见问题与解决方案

### 8.1 任务执行超时

- 检查任务timeout_seconds配置是否合理
- 确认外部系统(如Jira)响应是否正常
- 检查日志中的具体错误信息

### 8.2 定时任务未执行

- 验证cron表达式是否正确
- 检查系统时区设置
- 查看scheduler启动日志是否正常

### 8.3 结果未更新到Confluence

- 检查Confluence配置(URL、用户名、密码)
- 验证page_id是否正确
- 检查用户权限是否足够

## 9. 开发规范

### 9.1 代码风格

- 遵循PEP 8规范
- 使用类型注解
- 关键函数添加文档注释

### 9.2 异常处理

- 所有异常继承自BaseAppException
- 领域层异常应当有明确的业务含义
- 避免在领域层捕获基础设施异常

### 9.3 日志规范

- ERROR级别：影响系统运行的错误
- WARNING级别：需要注意但不影响主流程的问题
- INFO级别：重要操作节点信息
- DEBUG级别：详细调试信息

## 10. 未来规划

### 10.1 功能增强

- 支持更多外部系统集成
- 添加Web管理界面
- 实现分布式调度

### 10.2 性能优化

- 使用Redis存储任务状态
- 实现任务执行结果的异步处理
- 优化大量任务场景下的性能

### 10.3 安全增强

- API认证与授权
- 敏感信息加密存储
- 完善的审计日志

## 11. 贡献指南

1. Fork项目仓库
2. 创建特性分支
3. 提交变更
4. 提交Pull Request

## 12. 许可证

本项目采用MIT许可证。

---

*文档更新日期: 2024-03-10*