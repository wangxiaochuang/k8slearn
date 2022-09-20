package schema

import (
	"fmt"
	"strings"
)

// resource.group.com  resource.version.group.com
func ParseResourceArg(arg string) (*GroupVersionResource, GroupResource) {
	var gvr *GroupVersionResource
	if strings.Count(arg, ".") >= 2 {
		s := strings.SplitN(arg, ".", 3)
		gvr = &GroupVersionResource{Group: s[2], Version: s[1], Resource: s[0]}
	}

	return gvr, ParseGroupResource(arg)
}

// Kind.group.com Kind.version.group.com
func ParseKindArg(arg string) (*GroupVersionKind, GroupKind) {
	var gvk *GroupVersionKind
	if strings.Count(arg, ".") >= 2 {
		s := strings.SplitN(arg, ".", 3)
		gvk = &GroupVersionKind{Group: s[2], Version: s[1], Kind: s[0]}
	}

	return gvk, ParseGroupKind(arg)
}

type GroupResource struct {
	Group    string
	Resource string
}

func (gr GroupResource) WithVersion(version string) GroupVersionResource {
	return GroupVersionResource{Group: gr.Group, Version: version, Resource: gr.Resource}
}

func (gr GroupResource) Empty() bool {
	return len(gr.Group) == 0 && len(gr.Resource) == 0
}

func (gr GroupResource) String() string {
	if len(gr.Group) == 0 {
		return gr.Resource
	}
	return gr.Resource + "." + gr.Group
}

func ParseGroupKind(gk string) GroupKind {
	i := strings.Index(gk, ".")
	if i == -1 {
		return GroupKind{Kind: gk}
	}

	return GroupKind{Group: gk[i+1:], Kind: gk[:i]}
}

func ParseGroupResource(gr string) GroupResource {
	if i := strings.Index(gr, "."); i >= 0 {
		return GroupResource{Group: gr[i+1:], Resource: gr[:i]}
	}
	return GroupResource{Resource: gr}
}

type GroupVersionResource struct {
	Group    string
	Version  string
	Resource string
}

func (gvr GroupVersionResource) Empty() bool {
	return len(gvr.Group) == 0 && len(gvr.Version) == 0 && len(gvr.Resource) == 0
}

func (gvr GroupVersionResource) GroupResource() GroupResource {
	return GroupResource{Group: gvr.Group, Resource: gvr.Resource}
}

func (gvr GroupVersionResource) GroupVersion() GroupVersion {
	return GroupVersion{Group: gvr.Group, Version: gvr.Version}
}

func (gvr GroupVersionResource) String() string {
	return strings.Join([]string{gvr.Group, "/", gvr.Version, ", Resource=", gvr.Resource}, "")
}

type GroupKind struct {
	Group string
	Kind  string
}

func (gk GroupKind) Empty() bool {
	return len(gk.Group) == 0 && len(gk.Kind) == 0
}

func (gk GroupKind) WithVersion(version string) GroupVersionKind {
	return GroupVersionKind{Group: gk.Group, Version: version, Kind: gk.Kind}
}

func (gk GroupKind) String() string {
	if len(gk.Group) == 0 {
		return gk.Kind
	}
	return gk.Kind + "." + gk.Group
}

type GroupVersionKind struct {
	Group   string
	Version string
	Kind    string
}

func (gvk GroupVersionKind) Empty() bool {
	return len(gvk.Group) == 0 && len(gvk.Version) == 0 && len(gvk.Kind) == 0
}

func (gvk GroupVersionKind) GroupKind() GroupKind {
	return GroupKind{Group: gvk.Group, Kind: gvk.Kind}
}

func (gvk GroupVersionKind) GroupVersion() GroupVersion {
	return GroupVersion{Group: gvk.Group, Version: gvk.Version}
}

func (gvk GroupVersionKind) String() string {
	return gvk.Group + "/" + gvk.Version + ", Kind=" + gvk.Kind
}

type GroupVersion struct {
	Group   string
	Version string
}

func (gv GroupVersion) Empty() bool {
	return len(gv.Group) == 0 && len(gv.Version) == 0
}

// String puts "group" and "version" into a single "group/version" string. For the legacy v1
// it returns "v1".
func (gv GroupVersion) String() string {
	if len(gv.Group) > 0 {
		return gv.Group + "/" + gv.Version
	}
	return gv.Version
}

// Identifier implements runtime.GroupVersioner interface.
func (gv GroupVersion) Identifier() string {
	return gv.String()
}

func (gv GroupVersion) KindForGroupVersionKinds(kinds []GroupVersionKind) (target GroupVersionKind, ok bool) {
	for _, gvk := range kinds {
		if gvk.Group == gv.Group && gvk.Version == gv.Version {
			return gvk, true
		}
	}
	for _, gvk := range kinds {
		if gvk.Group == gv.Group {
			return gv.WithKind(gvk.Kind), true
		}
	}
	return GroupVersionKind{}, false
}

func ParseGroupVersion(gv string) (GroupVersion, error) {
	if (len(gv) == 0) || (gv == "/") {
		return GroupVersion{}, nil
	}

	switch strings.Count(gv, "/") {
	case 0:
		return GroupVersion{"", gv}, nil
	case 1:
		i := strings.Index(gv, "/")
		return GroupVersion{gv[:i], gv[i+1:]}, nil
	default:
		return GroupVersion{}, fmt.Errorf("unexpected GroupVersion string: %v", gv)
	}
}

func (gv GroupVersion) WithKind(kind string) GroupVersionKind {
	return GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: kind}
}

func (gv GroupVersion) WithResource(resource string) GroupVersionResource {
	return GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: resource}
}

type GroupVersions []GroupVersion

func (gvs GroupVersions) Identifier() string {
	groupVersions := make([]string, 0, len(gvs))
	for i := range gvs {
		groupVersions = append(groupVersions, gvs[i].String())
	}
	return fmt.Sprintf("[%s]", strings.Join(groupVersions, ","))
}

func (gvs GroupVersions) KindForGroupVersionKinds(kinds []GroupVersionKind) (GroupVersionKind, bool) {
	var targets []GroupVersionKind
	for _, gv := range gvs {
		target, ok := gv.KindForGroupVersionKinds(kinds)
		if !ok {
			continue
		}
		targets = append(targets, target)
	}
	if len(targets) == 1 {
		return targets[0], true
	}
	if len(targets) > 1 {
		return bestMatch(kinds, targets), true
	}
	return GroupVersionKind{}, false
}

func bestMatch(kinds []GroupVersionKind, targets []GroupVersionKind) GroupVersionKind {
	for _, gvk := range targets {
		for _, k := range kinds {
			if k == gvk {
				return k
			}
		}
	}
	return targets[0]
}

func (gvk GroupVersionKind) ToAPIVersionAndKind() (string, string) {
	if gvk.Empty() {
		return "", ""
	}
	return gvk.GroupVersion().String(), gvk.Kind
}

func FromAPIVersionAndKind(apiVersion, kind string) GroupVersionKind {
	if gv, err := ParseGroupVersion(apiVersion); err == nil {
		return GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: kind}
	}
	return GroupVersionKind{Kind: kind}
}
