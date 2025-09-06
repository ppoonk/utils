package balancer

import (
	"reflect"
	"slices"
	"sync"
)

func _check[T WeightedServer[T]]() {
	var _ Balancer[T] = (*SmoothWeightedRoundRobin[T])(nil)
}

/*
1、初始化阶段：泛型类型需包含 初始权重(Weight)、当前权重(CurrentWeight)、名称(Name) 相关方法
2、选择服务阶段：
   - 给所有服务的 CurrentWeight 加上各自的 Weight
   - 选择 CurrentWeight 最大的服务
   - 将选中服务的 CurrentWeight 减去所有服务的 Weight 总和
3、循环执行：重复步骤2，实现平滑权重分配
*/

// WeightedServer 是泛型约束，要求类型必须实现权重调度所需的核心方法
type WeightedServer[T any] interface {
	GetWeight() int              // 获取初始权重（固定）
	GetCurrentWeight() int       // 获取当前权重（动态调整）
	SetCurrentWeight(weight int) // 设置当前权重
	GetName() string             // 获取服务唯一名称（用于增删查）
}

// SmoothWeightedRoundRobin 平滑加权轮询调度器
type SmoothWeightedRoundRobin[T WeightedServer[T]] struct {
	servers []T
	sync.RWMutex
}

// NewSmoothWeightedRoundRobin 创建泛型版平滑加权轮询调度器
func NewSmoothWeightedRoundRobin[T WeightedServer[T]](servers []T) *SmoothWeightedRoundRobin[T] {
	validServers := make([]T, 0, len(servers))
	for _, s := range servers {
		if s.GetWeight() > 0 {
			validServers = append(validServers, s)
		}
	}

	return &SmoothWeightedRoundRobin[T]{
		servers: validServers,
	}
}

func isZero[T any](v T) bool {
	return reflect.ValueOf(v).IsZero()
}

// Select 选择下一个符合权重规则的服务
// 返回：选中的泛型服务实例（T类型），无可用服务时返回零值
func (s *SmoothWeightedRoundRobin[T]) Select() T {
	s.Lock()
	defer s.Unlock()

	if len(s.servers) == 0 {
		var zero T
		return zero
	}

	totalWeight := 0
	var selected T

	for _, server := range s.servers {
		totalWeight += server.GetWeight()
		newCurrentWeight := server.GetCurrentWeight() + server.GetWeight()
		server.SetCurrentWeight(newCurrentWeight)

		if isZero(selected) || server.GetCurrentWeight() > selected.GetCurrentWeight() {
			selected = server
		}
	}

	adjustedWeight := selected.GetCurrentWeight() - totalWeight
	selected.SetCurrentWeight(adjustedWeight)

	return selected
}

func (s *SmoothWeightedRoundRobin[T]) Add(server T) {
	if isZero(server) || server.GetWeight() <= 0 {
		return
	}
	s.Lock()
	defer s.Unlock()
	s.servers = append(s.servers, server)
}

func (s *SmoothWeightedRoundRobin[T]) Remove(name string) {
	s.Lock()
	defer s.Unlock()
	for i, server := range s.servers {
		if server.GetName() == name {
			s.servers = slices.Delete(s.servers, i, i+1)
			break
		}
	}
}
