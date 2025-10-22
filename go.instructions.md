# Go 项目开发指南

本文档基于 [Google Go Style Guide](https://google.github.io/styleguide/go/)，用于指导 AI 编写符合 Google 标准的 Go 代码。

## 核心原则（优先级顺序）

1. **清晰性（Clarity）** - 代码的目的和原理必须让读者易于理解
2. **简洁性（Simplicity）** - 用最直接的方式完成目标
3. **精炼性（Concision）** - 保持代码的高信噪比
4. **可维护性（Maintainability）** - 便于未来修改
5. **一致性（Consistency）** - 与更广泛的代码库模式保持一致

## 强制性格式规范

### 工具合规性
- 所有源文件必须符合 `gofmt` 的输出格式
- 使用 `go fmt` 自动格式化代码

### 命名约定

**基本规则：使用驼峰命名法（MixedCaps 或 mixedCaps），绝不使用下划线**

```go
// 导出的（公开的）- 首字母大写
type UserAccount struct{}
const MaxLength = 100
func ParseRequest() {}

// 未导出的（私有的）- 首字母小写
type internalCache struct{}
const maxRetries = 3
func validateInput() {}
```

**例外情况（可以使用下划线）：**
1. 仅由生成代码导入的包
2. `*_test.go` 文件中的测试函数名
3. 低级别的操作系统/cgo 库

### 包命名

```go
// ✅ 推荐 - 简洁、小写、无中断
package tabwriter
package httputil

// ❌ 避免 - 使用下划线
package tab_writer
package http_util

// ❌ 避免 - 过于通用的名称
package util
package common
package helpers
```

### 常量命名

```go
// ✅ 根据用途命名，使用驼峰
const MaxPacketSize = 512
const DefaultTimeout = 30

// ❌ 避免全大写加下划线
const MAX_PACKET_SIZE = 512
```

### 缩写词处理

保持缩写词内部大小写一致：

```go
// ✅ 正确
type XMLAPI struct{}      // 导出的
type xmlAPI struct{}      // 未导出的

type UserID int           // 导出的
type userID int           // 未导出的

// iOS 特殊处理
type IOSApp struct{}      // 导出的
type iOSApp struct{}      // 未导出的

// ❌ 错误 - 不一致的大小写
type XmlApi struct{}
type UserId int
```

### 接收器命名

```go
// ✅ 简短的缩写，1-2 个字母，保持一致
func (c *Client) Connect() {}
func (c *Client) Disconnect() {}

func (u *UserAccount) Validate() {}
func (u *UserAccount) Save() {}

// ❌ 避免使用完整类型名
func (client *Client) Connect() {}

// ❌ 避免不一致
func (c *Client) Connect() {}
func (cl *Client) Disconnect() {}
```

### Getter 命名

```go
// ✅ 省略 Get 前缀
func (c *Client) Counts() int {}
func (u *User) Name() string {}

// 昂贵操作使用 Compute 或 Fetch
func (s *Stats) ComputeTotal() int {}
func (d *Database) FetchUsers() []User {}

// ❌ 避免 Get 前缀
func (c *Client) GetCounts() int {}
```

### 变量命名

**作用域原则：**
- 名称长度应与作用域大小成正比
- 名称长度应与使用频率成反比

```go
// ✅ 短作用域使用短名称
for i := 0; i < 10; i++ {}
if err := doSomething(); err != nil {}

// ✅ 长作用域使用描述性名称
var authenticatedUserSessions map[string]*Session

// ✅ 不在名称中包含类型
users := []User{}           // 而不是 userSlice
counts := map[string]int{}  // 而不是 countMap

// ❌ 避免包名/导出名重复
// 在 user 包中
type UserManager struct{}  // ❌ 使用时: user.UserManager
type Manager struct{}      // ✅ 使用时: user.Manager
```

## 导入管理

### 导入分组

按以下顺序分为四组，组间用空行分隔：

```go
import (
    // 1. 标准库
    "context"
    "fmt"
    "os"

    // 2. 其他包
    "github.com/pkg/errors"
    "go.uber.org/zap"

    // 3. Protocol Buffers
    pb "myproject/gen/proto/go/myproject/v1"

    // 4. 副作用导入（仅导入执行 init）
    _ "embed"
)
```

### 导入规则

```go
// ❌ 永远不要使用点导入（除了特殊测试场景）
import . "fmt"

// ✅ 空白导入仅限于 main 包或测试
import _ "net/http/pprof"

// ✅ 重命名导入 - 解决冲突或遵循约定
import (
    neturl "net/url"
    pb "myproject/gen/proto/go/myproject/v1"
)
```

## 错误处理

### 返回模式

```go
// ✅ 错误应该是最后一个返回值
func Open(name string) (*File, error) {}
func Parse(data []byte) (Result, error) {}

// ✅ 多返回值 - 避免带内错误
func Lookup(key string) (value string, ok bool) {}

// ❌ 避免使用 -1、nil 或空字符串表示错误
func Find(key string) string {
    // 返回 "" 表示未找到 - 不好
}
```

### 错误字符串

```go
// ✅ 小写开头，不加标点（专有名词除外）
fmt.Errorf("something bad happened")
fmt.Errorf("failed to connect to database")
errors.New("invalid input")

// ❌ 避免大写或标点
fmt.Errorf("Something bad happened.")
```

### 错误处理策略

```go
// ✅ 立即处理，提前返回
func process() error {
    data, err := fetch()
    if err != nil {
        return fmt.Errorf("failed to fetch: %w", err)
    }
    // 继续正常流程
    return save(data)
}

// ❌ 永远不要忽略错误
data, _ := fetch()  // 糟糕！

// ❌ 避免嵌套正常代码
func process() error {
    data, err := fetch()
    if err == nil {
        // 正常代码嵌套在这里 - 不好
        return save(data)
    }
    return err
}
```

### 错误包装

```go
// ✅ 使用 %w 允许调用者检查错误链
if err := doSomething(); err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// 调用者可以使用
if errors.Is(err, ErrNotFound) {}
if errors.As(err, &specificErr) {}

// ✅ 在系统边界（RPC、存储）使用 %v
// 将领域特定错误转换为规范错误空间
return status.Errorf(codes.NotFound, "user not found: %v", err)

// ✅ 错误包装时，将 %w 放在字符串末尾
return fmt.Errorf("failed to process request: %w", err)
```

### 结构化错误

```go
// ✅ 创建可编程检查的错误
var ErrNotFound = errors.New("not found")

type ValidationError struct {
    Field string
    Value interface{}
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("invalid %s: %v", e.Field, e.Value)
}

// 使用
if errors.Is(err, ErrNotFound) {
    // 处理未找到的情况
}

var valErr *ValidationError
if errors.As(err, &valErr) {
    // 处理验证错误
}
```

### 日志记录策略

```go
// ✅ 避免同时记录和返回错误
func process() error {
    if err := validate(); err != nil {
        // ❌ 不要这样做
        log.Error("validation failed", err)
        return err  // 调用者可能也会记录
    }
    return nil
}

// ✅ 让调用者决定是否记录
func process() error {
    if err := validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    return nil
}

// 调用者决定
if err := process(); err != nil {
    log.Error("processing error", err)
}
```

## 函数设计

### 函数签名格式

```go
// ✅ 保持单行
func (c *Client) Connect(ctx context.Context, addr string, timeout time.Duration) error {}

// ✅ 参数过多时，提取为选项结构
type ConnectOptions struct {
    Address string
    Timeout time.Duration
    Retry   int
}

func (c *Client) Connect(ctx context.Context, opts ConnectOptions) error {}
```

### 接收器选择（指针 vs 值）

**使用指针接收器当：**
- 方法需要修改接收器
- 接收器包含不可安全复制的字段（如 `sync.Mutex`）
- 接收器是大型结构
- 需要支持并发访问
- 包含指向可变数据的指针

**保持一致性：**
```go
// ✅ 类型的所有方法使用相同的接收器风格
type Client struct {
    conn net.Conn
}

func (c *Client) Connect() error {}     // 都用指针
func (c *Client) Disconnect() error {}  // 都用指针
func (c *Client) IsConnected() bool {}  // 都用指针

// ❌ 避免混用
func (c *Client) Connect() error {}     // 指针
func (c Client) Disconnect() error {}   // 值 - 不一致！
```

### 传值 vs 传指针

```go
// ✅ 小结构传值
type Point struct {
    X, Y int
}
func Distance(p1, p2 Point) float64 {}

// ✅ 大结构或 Protocol Buffer 传指针
type Config struct {
    // 很多字段...
}
func Apply(cfg *Config) error {}

// ✅ Protocol Buffers 始终使用指针
func Process(req *pb.Request) (*pb.Response, error) {}
```

### 命名返回值

```go
// ✅ 用于澄清调用者责任
func Split(path string) (dir, file string) {
    // 实现
    return
}

// ✅ 用于延迟闭包
func process() (err error) {
    f, err := os.Open("file")
    if err != nil {
        return err
    }
    defer func() {
        if closeErr := f.Close(); closeErr != nil && err == nil {
            err = closeErr
        }
    }()
    // 处理文件
    return nil
}

// ❌ 避免造成重复
func (n *Node) Parent() (node *Node) {}  // 冗余
func (n *Node) Parent() *Node {}         // 更好

// ✅ 裸返回仅用于小函数
func add(a, b int) (result int) {
    result = a + b
    return  // OK，函数很小
}
```

## 控制流

### 条件语句

```go
// ✅ 保持单行或提取条件
if user.IsActive && user.HasPermission("write") && user.Age >= 18 {
    // 处理
}

// ✅ 复杂条件提取为局部变量
canWrite := user.IsActive &&
            user.HasPermission("write") &&
            user.Age >= 18
if canWrite {
    // 处理
}

// ✅ 变量在左侧
if result == "foo" {  // ✅
if "foo" == result {  // ❌
```

### 循环

```go
// ✅ 保持单行
for i := 0; i < len(items); i++ {}
for key, value := range m {}

// ✅ 或提取条件到循环体
for {
    item, err := next()
    if err != nil {
        break
    }
    process(item)
}
```

### Switch 语句

```go
// ✅ case 保持单行
switch status {
case "active":
    return true
case "inactive":
    return false
default:
    return false
}

// ✅ 不需要 break（Go 自动终止）
switch x {
case 1:
    fmt.Println("one")
    // 不需要 break
case 2:
    fmt.Println("two")
}
```

## 类型和接口

### 接口设计

```go
// ✅ 在消费包中定义接口，而不是实现包
// storage 包
type Repository interface {
    Save(item Item) error
    Find(id string) (Item, error)
}

// ✅ 返回具体类型，而不是接口
func NewClient() *Client {}        // ✅
func NewClient() Interface {}      // ❌

// ✅ 仅在真正需要时创建接口
// 至少有两个实现，或者有明确的模拟需求
```

### 结构体字面量

```go
// ✅ 其他包的类型使用字段名
user := pkg.User{
    Name: "Alice",
    Age:  30,
}

// ✅ 省略零值字段（当清晰度不受影响时）
config := Config{
    Host: "localhost",
    // Port: 0,  // 可以省略
}

// ✅ 多行时匹配大括号缩进
config := Config{
    Database: DatabaseConfig{
        Host: "localhost",
        Port: 5432,
    },
    Cache: CacheConfig{
        TTL: time.Hour,
    },
}
```

### Nil Slices

```go
// ✅ 优先使用 nil 初始化局部变量
var items []Item  // nil slice

// ✅ 不要在 API 中区分 nil 和空 slice
func process(items []Item) {
    if len(items) == 0 {  // 对 nil 和空 slice 都有效
        return
    }
}

// ❌ 不需要
if items != nil && len(items) > 0 {}  // 多余
if len(items) > 0 {}                  // 足够
```

### 类型别名

```go
// ✅ 使用类型定义创建新类型
type UserID int64

// ❌ 避免类型别名（除非包迁移）
type UserID = int64  // 只在包迁移时使用
```

## 注释和文档

### 文档注释

```go
// ✅ 所有导出的名称都需要文档注释
// 以被描述的名称开头

// User 表示系统中的用户账户。
type User struct {
    Name string
    Age  int
}

// NewUser 创建一个具有给定名称的新用户。
func NewUser(name string) *User {
    return &User{Name: name}
}

// ✅ 非导出的复杂类型也应该有文档
// cache 存储用户会话以实现快速查找。
type cache struct {
    sessions map[string]*Session
}
```

### 注释长度

```go
// ✅ 目标是 80 字符行（方便窄屏阅读）
// 但不要强制 - 在标点符号和语义单元处断行

// Process 处理传入的请求，验证输入，应用业务逻辑，
// 并返回适当的响应。如果请求无效或处理失败，
// 将返回错误。
func Process(req *Request) (*Response, error) {}
```

### 包注释

```go
// ✅ 包注释必须紧接在 package 子句之前
// 每个包恰好一个

// Package user 提供用户账户管理功能。
//
// 此包处理用户创建、身份验证和授权。
// 它与数据库交互以持久化用户数据。
package user
```

### 注释风格

```go
// ✅ 完整句子应大写和标点
// Process validates and saves the user.
func Process(u *User) error {}

// ✅ 片段不需要标点
// maximum retries
const maxRetries = 3

// ✅ 解释"为什么"，而不仅仅是"是什么"
// 使用缓冲通道避免在高负载时阻塞生产者
events := make(chan Event, 100)

// ❌ 避免重述代码
// 设置 x 为 1
x := 1  // 无用的注释
```

## 测试

### 测试函数命名

```go
// ✅ 测试函数可以使用下划线
func TestUser_Create(t *testing.T) {}
func TestUser_Update_InvalidInput(t *testing.T) {}
```

### 失败消息

```go
// ✅ 包含函数名、输入、实际结果、期望结果
func TestAdd(t *testing.T) {
    got := Add(2, 3)
    want := 5
    if got != want {
        t.Errorf("Add(2, 3) = %d, want %d", got, want)
    }
}
```

### 表驱动测试

```go
// ✅ 使用字段名提高清晰度
func TestParse(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    Result
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "hello",
            want:    Result{Value: "hello"},
            wantErr: false,
        },
        {
            name:    "empty input",
            input:   "",
            want:    Result{},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Parse(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Parse(%q) error = %v, wantErr %v",
                    tt.input, err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Parse(%q) = %v, want %v",
                    tt.input, got, tt.want)
            }
        })
    }
}
```

### 比较

```go
// ✅ 使用 cmp.Equal 和 cmp.Diff
import "github.com/google/go-cmp/cmp"

if diff := cmp.Diff(want, got); diff != "" {
    t.Errorf("result mismatch (-want +got):\n%s", diff)
}

// ❌ 避免断言库或手动字段比较
```

### Fatal vs Error

```go
// ✅ 对设置失败使用 t.Fatal
func TestProcess(t *testing.T) {
    db, err := setupDatabase()
    if err != nil {
        t.Fatalf("failed to setup database: %v", err)
    }

    // 对测试失败使用 t.Error（报告所有问题）
    result := Process(db)
    if result.Count != 5 {
        t.Errorf("got count %d, want 5", result.Count)
    }
    if result.Status != "ok" {
        t.Errorf("got status %q, want %q", result.Status, "ok")
    }
}

// ✅ 表驱动测试中，在子测试中使用 t.Fatal
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got, err := Parse(tt.input)
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        // 继续测试
    })
}
```

## 并发

### Goroutine 生命周期

```go
// ✅ 明确 goroutine 何时退出
func process(ctx context.Context) {
    go func() {
        <-ctx.Done()
        cleanup()
    }()
}

// ✅ 使用 WaitGroup 等待完成
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        work(id)
    }(i)
}
wg.Wait()

// ❌ 永远不要在不知道如何终止的情况下生成 goroutine
go func() {
    for {
        // 永远运行？何时停止？
        work()
    }
}()
```

### 同步函数优先

```go
// ✅ 优先同步函数 - 让调用者添加并发
func Fetch(url string) ([]byte, error) {
    // 同步实现
    return http.Get(url)
}

// 调用者可以并发调用
go Fetch("http://example.com")

// ❌ 避免异步 API（除非有充分理由）
func FetchAsync(url string) <-chan Result {
    // 强制异步
}
```

### Context 使用

```go
// ✅ 始终将 context.Context 作为第一个参数
func Process(ctx context.Context, data []byte) error {}

// ✅ 仅在 main、init 或测试入口使用 context.Background()
func main() {
    ctx := context.Background()
    // 使用 ctx
}

// ✅ 传递 context
func handler(ctx context.Context) error {
    return process(ctx, data)
}

// ❌ 永远不要创建自定义 context 类型
type MyContext struct {
    context.Context
    CustomField string
}
```

### Channel 方向

```go
// ✅ 始终指定 channel 方向以防止误用
func producer(ch chan<- int) {
    ch <- 42
    // 无法从 ch 接收 - 编译时错误
}

func consumer(ch <-chan int) {
    val := <-ch
    // 无法向 ch 发送 - 编译时错误
}

func process() {
    ch := make(chan int)
    go producer(ch)
    consumer(ch)
}
```

## API 设计最佳实践

### 选项模式

**选项结构：**
```go
// ✅ 用于收集相关参数
type ServerOptions struct {
    Host    string
    Port    int
    Timeout time.Duration
    MaxConn int
}

func NewServer(opts ServerOptions) *Server {
    // 应用默认值
    if opts.Port == 0 {
        opts.Port = 8080
    }
    return &Server{opts: opts}
}

// 使用
srv := NewServer(ServerOptions{
    Host:    "localhost",
    Timeout: 30 * time.Second,
    // 省略 Port - 将使用默认值
})
```

**可变选项：**
```go
// ✅ 用于灵活配置
type ServerOption func(*Server)

func WithPort(port int) ServerOption {
    return func(s *Server) {
        s.port = port
    }
}

func WithTimeout(d time.Duration) ServerOption {
    return func(s *Server) {
        s.timeout = d
    }
}

func NewServer(host string, opts ...ServerOption) *Server {
    s := &Server{
        host:    host,
        port:    8080,  // 默认值
        timeout: 30 * time.Second,
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// 使用 - 简单调用保持简洁
srv := NewServer("localhost")
srv := NewServer("localhost", WithPort(9000), WithTimeout(time.Minute))
```

## 依赖哲学

**优先级顺序：**
1. 核心语言构造（channels、slices、maps、loops、structs）
2. 标准库工具
3. 项目内部库，然后才是外部依赖

```go
// ✅ 优先使用标准库
import (
    "encoding/json"
    "net/http"
    "time"
)

// ✅ 仅在必要时添加外部依赖
import "github.com/google/uuid"
```

## 变量声明

```go
// ✅ 使用 := 进行非零初始化
name := "Alice"
count := 42
users := []User{{Name: "Bob"}}

// ✅ 使用 var 声明零值（表示"空但可用"）
var buf bytes.Buffer  // 准备使用
var users []User      // nil slice，准备追加

// ✅ 复合字面量用于已知初始值
config := Config{
    Host: "localhost",
    Port: 8080,
}

// ❌ 不要预分配集合（除非有实证分析）
users := make([]User, 0, 100)  // 通常不必要
users := []User{}               // 让运行时管理增长
```

## 包设计

```go
// ✅ 包应包含概念上相关的功能
package user      // 用户管理
package auth      // 身份验证
package storage   // 数据持久化

// ❌ 避免通用名称
package util      // 太通用
package helper    // 没有意义
package common    // 不描述功能

// ✅ 包内文件应按逻辑耦合组织
user/
  user.go          // 核心类型
  validation.go    // 验证逻辑
  repository.go    // 数据访问
```

## Panic 和 Recovery

```go
// ✅ Panic 很少应该跨越包边界
// 异常：API 误用检测
func (s *Stack) Pop() int {
    if len(s.items) == 0 {
        panic("Pop from empty stack")
    }
    return s.items[len(s.items)-1]
}

// ✅ 内部解析器可能使用 panic 作为实现细节
func (p *parser) parse() (result AST, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("parse error: %v", r)
        }
    }()
    // 内部可能 panic
    return p.parseInternal(), nil
}
```

## 性能考虑

```go
// ✅ 为切片和映射提供大小提示应基于实证分析
// 大多数代码受益于运行时管理的增长

// ❌ 避免推测性优化
users := make([]User, 0, 1000)  // 为什么是 1000？

// ✅ 仅在分析显示有益时预分配
// 基准测试显示此热路径从预分配中受益
results := make([]Result, 0, len(inputs))
```

## 工具链

### 必需工具

```bash
# 格式化
go fmt ./...

# Lint
golangci-lint run

# 测试
go test ./...
go test -race ./...  # 竞态检测
go test -cover ./... # 覆盖率

# 构建
go build ./...
```

### 推荐工具

```bash
# 静态分析
go vet ./...

# 依赖管理
go mod tidy
go mod verify

# 文档
godoc -http=:6060
```

## 快速检查清单

在提交代码前检查：

- [ ] 运行 `go fmt` 格式化代码
- [ ] 所有导出的名称都有文档注释
- [ ] 使用驼峰命名，无下划线（除非例外情况）
- [ ] 错误是最后一个返回值
- [ ] 错误字符串小写，无标点
- [ ] 错误处理明确（不忽略）
- [ ] 使用 `%w` 包装需要检查的错误
- [ ] 接口在消费包中定义
- [ ] 返回具体类型，而非接口
- [ ] 测试使用表驱动方法
- [ ] Context 作为第一个参数
- [ ] Channel 指定方向
- [ ] Goroutine 有明确的退出条件
- [ ] 包名简洁、小写、无下划线
- [ ] 避免通用包名（util、common）
- [ ] 所有测试通过
- [ ] 运行竞态检测器

## 参考资源

- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)

---

**记住：代码是为人类阅读而写的，只是恰好可以被机器执行。始终优先考虑清晰度和可维护性。**
