# ID

> 使用 sonyflake 生成全局唯一 ID，使用 Code 生成指定字符集的唯一 ID

# Sonyflake

## `options.go`

`SonyflakeOptions`结构体用于配置`Sonyflake`的参数：

```go
type SonyflakeOptions struct {
	machineId uint16    // 机器ID
	startTime time.Time // 起始时间
}
```



`getSonyflakeOptionsOrSetDefault`：用于获取或设置默认的Sonyflake选项

```go
func getSonyflakeOptionsOrSetDefault(options *SonyflakeOptions) *SonyflakeOptions {
	if options == nil {
		return &SonyflakeOptions{
			machineId: 1,
			startTime: time.Date(2022, 10, 10, 0, 0, 0, 0, time.UTC),
		}
	}
	return options
}
```



选项模式的配置函数：

```go
// WithSonyflakeMachineId 函数返回一个函数，用于设置机器ID。
func WithSonyflakeMachineId(id uint16) func(*SonyflakeOptions) {
	return func(options *SonyflakeOptions) {
		if id > 0 {
			getSonyflakeOptionsOrSetDefault(options).machineId = id
		}
	}
}

// WithSonyflakeStartTime 函数返回一个函数，用于设置起始时间。
func WithSonyflakeStartTime(startTime time.Time) func(*SonyflakeOptions) {
	return func(options *SonyflakeOptions) {
		if !startTime.IsZero() {
			getSonyflakeOptionsOrSetDefault(options).startTime = startTime
		}
	}
}
```



## sonyflake.go

创建配置`sonyflake`的配置：

```go
type SonyflakeOptions struct{
    startTime time.Time // Sonyflake 的起始时间
    machineId uint16    // Sonyflake 的机器 ID
}
```



Sonyflake实例：

```go
type Sonyflake struct {
	ops   SonyflakeOptions     // 配置选项。
	sf    *sonyflake.Sonyflake // Sonyflake 实例。
	Error error                // 初始化过程中的错误。
}
```



`NewSonyflake` 使用提供的选项函数【选项模式】创建一个新的 `Sonyflake` 实例：

```go
func NewSonyflake(options ...func(sonyflakeOptions *SonyflakeOptions)) *Sonyflake {
	// 获取默认选项或设置提供的选项
	ops := getSonyflakeOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	// 使用提供的选项初始化 Sonyflake
	sf := &Sonyflake{
		ops: *ops,
	}
	st := sonyflake.Settings{
		StartTime: ops.startTime,
	}
	if ops.machineId > 0 {
		st.MachineID = func() (uint16, error) {
			return ops.machineId, nil
		}
	}

	// 创建 Sonyflake 实例
	ins := sonyflake.NewSonyflake(st)
	if ins == nil {
		sf.Error = fmt.Errorf("create sonyflake failed")
	}

	// 检查时间是否有效，当时间错误后，NextID返回错误
	_, err := ins.NextID()
	if err != nil {
		sf.Error = fmt.Errorf("invalid start time")
	}
	sf.sf = ins
	return sf
}
```



`Id`函数：使用 `Sonyflake` 实例生成唯一`ID`

```go
func (s *Sonyflake) Id(ctx context.Context) (id uint64) {
    // 如果有错误，则直接返回
	if s.Error != nil {
		return
	}
	var err error
	// 生成 ID，如果失败，则使用指数回退策略重试。
	id, err = s.sf.NextID()
	if err == nil {
		return
	}
	sleep := 1
	for {
		time.Sleep(time.Duration(sleep) * time.Millisecond)
		id, err = s.sf.NextID()
		if err == nil {
			return
		}
		sleep *= 2
	}
}
```



# Code

## Options.go

`CodeOptions`结构体用于配置生成Code的参数

```go
type CodeOptions struct {
	chars []rune
	n1    int
	n2    int
	l     int
	salt  uint64
}
```



`getCodeOptionsOrSetDefault`函数：用于获取或设置默认的Code选项

```go
func getCodeOptionsOrSetDefault(options *CodeOptions) *CodeOptions {
	if options == nil {
		return &CodeOptions{
			// base string set, remove 0,1,I,O,U,Z
			chars: []rune{
				'2', '3', '4', '5', '6',
				'7', '8', '9', 'A', 'B',
				'C', 'D', 'E', 'F', 'G',
				'H', 'J', 'K', 'L', 'M',
				'N', 'P', 'Q', 'R', 'S',
				'T', 'V', 'W', 'X', 'Y',
			},
			// n1 / len(chars)=30 cop rime
			n1: 17,
			// n2 / l cop rime
			n2: 5,
			// code length
			l: 8,
			// random number
			salt: 123567369,
		}
	}
	return options
}
```



选项模式的配置函数：

```go
// WithCodeChars 函数返回一个函数，用于设置Code字符集。
func WithCodeChars(arr []rune) func(*CodeOptions) {
	return func(options *CodeOptions) {
		if len(arr) > 0 {
			getCodeOptionsOrSetDefault(options).chars = arr
		}
	}
}

// WithCodeN1 函数返回一个函数，用于设置与字符集长度互质的整数。
func WithCodeN1(n int) func(*CodeOptions) {
	return func(options *CodeOptions) {
		getCodeOptionsOrSetDefault(options).n1 = n
	}
}

// WithCodeN2 函数返回一个函数，用于设置与长度互质的整数。
func WithCodeN2(n int) func(*CodeOptions) {
	return func(options *CodeOptions) {
		getCodeOptionsOrSetDefault(options).n2 = n
	}
}

// WithCodeL 函数返回一个函数，用于设置Code长度。
func WithCodeL(l int) func(*CodeOptions) {
	return func(options *CodeOptions) {
		if l > 0 {
			getCodeOptionsOrSetDefault(options).l = l
		}
	}
}

// WithCodeSalt 函数返回一个函数，用于设置Salt。
func WithCodeSalt(salt uint64) func(*CodeOptions) {
	return func(options *CodeOptions) {
		if salt > 0 {
			getCodeOptionsOrSetDefault(options).salt = salt
		}
	}
}
```



## code.go

`NewCode`根据提供的ID生成唯一的Code：

```go
func NewCode(id uint64, options ...func(*CodeOptions)) string {
	// 获取或设置默认的Code选项
	ops := getCodeOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	// 扩大 ID 并添加 salt
	id = id*uint64(ops.n1) + ops.salt

	var code []rune
	slIdx := make([]byte, ops.l)

	charLen := len(ops.chars)
	charLenUI := uint64(charLen)

	// 扩散过程
	for i := 0; i < ops.l; i++ {
		slIdx[i] = byte(id % charLenUI)
		slIdx[i] = (slIdx[i] + byte(i)*slIdx[0]) % byte(charLen)
		id /= charLenUI
	}

	// 混淆过程
	for i := 0; i < ops.l; i++ {
		idx := (byte(i) * byte(ops.n2)) % byte(ops.l)
		code = append(code, ops.chars[slIdx[idx]])
	}
	return string(code)
}
```



# Use

## sonyflake.go

```go
package main

import (
	"context"
	"fmt"
	"github.com/LiangNing7/onex/pkg/id"
)

func main() {
	sf := id.NewSonyflake(
		id.WithSonyflakeMachineId(1),
	)
	if sf.Error != nil {
		fmt.Println(sf.Error)
		return
	}
	fmt.Println("生成是唯一ID为:",sf.Id(context.Background()))
}
```

`code.go`

```go
package main

import (
	"fmt"
	"github.com/LiangNing7/onex/pkg/id"
)

func main() {
	fmt.Println(id.NewCode(1))
	fmt.Println(id.NewCode(2))
	fmt.Println(id.NewCode(3))
	fmt.Println(id.NewCode(4))

	fmt.Println(id.NewCode(
		1,
		id.WithCodeChars([]rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}),
		id.WithCodeN1(9),
		id.WithCodeN2(3),
		id.WithCodeL(5),
		id.WithCodeSalt(99999),
	))
}
```

