package id

import (
	"context"
	"fmt"
	"github.com/sony/sonyflake"
	"time"
)

type Sonyflake struct {
	ops   SonyflakeOptions
	sf    *sonyflake.Sonyflake
	Error error
}

// NewSonyflake 根据提供的选项函数创建一个新的 Sonyflake 实例
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

// Id 使用 Sonyflake 实例生成唯一 ID
func (s *Sonyflake) Id(ctx context.Context) (id uint64) {
	// 如果有错误，则直接返回
	if s.Error != nil {
		return
	}

	var err error
	// 生成ID，生成成功则返回
	id, err = s.sf.NextID()
	if err == nil {
		return
	}
	// 生成失败则使用指数回避算法重试
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
