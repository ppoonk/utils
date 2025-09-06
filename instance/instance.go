package instance

import (
	"errors"
	"fmt"
	"sync"
)

type Destroyable interface {
	Destroy() error
}

// Manager 实例管理器
type Manager[T any] struct {
	instances sync.Map // key: alias (string), value: T
}

func NewManager[T any]() *Manager[T] {
	m := &Manager[T]{}
	return m
}

func (m *Manager[T]) Load(alias string) (T, error) {
	var zero T
	val, ok := m.instances.Load(alias)
	if !ok {
		return zero, fmt.Errorf("instance not found: alias=%q", alias)
	}

	instance, ok := val.(T)
	if !ok {
		err := fmt.Errorf("type mismatch for alias=%q: expected %T, got %T",
			alias, zero, val)
		return zero, err
	}

	return instance, nil
}

func (m *Manager[T]) Add(alias string, instance T) error {
	if isNil(instance) {
		return fmt.Errorf("cannot store nil instance: alias=%q", alias)
	}
	if m.Exists(alias) {
		_ = m.Remove(alias)
	}
	m.instances.Store(alias, instance)
	return nil
}

func (m *Manager[T]) Remove(alias string) error {
	var zero T
	val, ok := m.instances.LoadAndDelete(alias)
	if !ok {
		return fmt.Errorf("warn: no instance to remove: alias=%q", alias)
	}
	instance, ok := val.(T)
	if !ok {
		err := fmt.Errorf("type mismatch when removing alias=%q: expected %T, got %T", alias, zero, val)
		return err
	}

	destroyable, ok := any(instance).(Destroyable)
	if ok {
		_ = destroyable.Destroy()
	}
	return nil
}

func (m *Manager[T]) Exists(alias string) bool {
	_, ok := m.instances.Load(alias)
	return ok
}

func (m *Manager[T]) List() map[string]T {
	instances := make(map[string]T)
	m.instances.Range(func(key, value any) bool {
		alias, ok := key.(string)
		if !ok {
			return true
		}

		instance, ok := value.(T)
		if !ok {
			return true
		}

		instances[alias] = instance
		return true
	})
	return instances
}

func (m *Manager[T]) Clear() error {
	var errs []error
	m.instances.Range(func(key, value any) bool {
		alias, ok := key.(string)
		if !ok {
			return true
		}

		if err := m.Remove(alias); err != nil {
			errs = append(errs, fmt.Errorf("alias=%q: %w", alias, err))
		}
		return true
	})

	if len(errs) > 0 {
		return fmt.Errorf("failed to clear %d instances: %w", len(errs), errors.Join(errs...))
	}
	return nil
}

func isNil[T any](v T) bool {
	return any(v) == nil
}
