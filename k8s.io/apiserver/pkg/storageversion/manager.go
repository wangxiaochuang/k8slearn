package storageversion

import (
	"fmt"
	"sync"
	"sync/atomic"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

type ResourceInfo struct {
	GroupResource schema.GroupResource

	EncodingVersion          string
	EquivalentResourceMapper runtime.EquivalentResourceRegistry

	DirectlyDecodableVersions []schema.GroupVersion
}
type Manager interface {
	AddResourceInfo(resources ...*ResourceInfo)
	UpdateStorageVersions(kubeAPIServerClientConfig *rest.Config, apiserverID string)
	PendingUpdate(gr schema.GroupResource) bool
	LastUpdateError(gr schema.GroupResource) error
	Completed() bool
}

var _ Manager = &defaultManager{}

type defaultManager struct {
	completed atomic.Value

	mu                   sync.RWMutex
	managedResourceInfos map[*ResourceInfo]struct{}
	managedStatus        map[schema.GroupResource]*updateStatus
}

type updateStatus struct {
	done    bool
	lastErr error
}

func NewDefaultManager() Manager {
	s := &defaultManager{}
	s.completed.Store(false)
	s.managedResourceInfos = make(map[*ResourceInfo]struct{})
	s.managedStatus = make(map[schema.GroupResource]*updateStatus)
	return s
}

func (s *defaultManager) AddResourceInfo(resources ...*ResourceInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range resources {
		s.managedResourceInfos[r] = struct{}{}
		s.addPendingManagedStatusLocked(r)
	}
}

func (s *defaultManager) addPendingManagedStatusLocked(r *ResourceInfo) {
	gvrs := r.EquivalentResourceMapper.EquivalentResourcesFor(r.GroupResource.WithVersion(""), "")
	for _, gvr := range gvrs {
		gr := gvr.GroupResource()
		if _, ok := s.managedStatus[gr]; !ok {
			s.managedStatus[gr] = &updateStatus{}
		}
	}
}

// p114
func (s *defaultManager) UpdateStorageVersions(kubeAPIServerClientConfig *rest.Config, serverID string) {
	panic("not implemented")
}

// p188
type byGroupResource []ResourceInfo

func (s byGroupResource) Len() int { return len(s) }

func (s byGroupResource) Less(i, j int) bool {
	if s[i].GroupResource.Group == s[j].GroupResource.Group {
		return s[i].GroupResource.Resource < s[j].GroupResource.Resource
	}
	return s[i].GroupResource.Group < s[j].GroupResource.Group
}

func (s byGroupResource) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// p245
func (s *defaultManager) PendingUpdate(gr schema.GroupResource) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.managedStatus[gr]; !ok {
		return false
	}
	return !s.managedStatus[gr].done
}

func (s *defaultManager) LastUpdateError(gr schema.GroupResource) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.managedStatus[gr]; !ok {
		return fmt.Errorf("couldn't find managed status for %v", gr)
	}
	return s.managedStatus[gr].lastErr
}

func (s *defaultManager) setComplete() {
	s.completed.Store(true)
}

func (s *defaultManager) Completed() bool {
	return s.completed.Load().(bool)
}
