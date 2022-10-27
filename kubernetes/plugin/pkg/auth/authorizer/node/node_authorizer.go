package node

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/featuregate"
	coordapi "k8s.io/kubernetes/pkg/apis/coordination"
	api "k8s.io/kubernetes/pkg/apis/core"
	storageapi "k8s.io/kubernetes/pkg/apis/storage"
	"k8s.io/kubernetes/pkg/auth/nodeidentifier"
	"k8s.io/kubernetes/third_party/forked/gonum/graph"
	"k8s.io/kubernetes/third_party/forked/gonum/graph/traverse"
)

type NodeAuthorizer struct {
	graph      *Graph
	identifier nodeidentifier.NodeIdentifier
	nodeRules  []rbacv1.PolicyRule

	// allows overriding for testing
	features featuregate.FeatureGate
}

var _ = authorizer.Authorizer(&NodeAuthorizer{})
var _ = authorizer.RuleResolver(&NodeAuthorizer{})

func NewAuthorizer(graph *Graph, identifier nodeidentifier.NodeIdentifier, rules []rbacv1.PolicyRule) *NodeAuthorizer {
	return &NodeAuthorizer{
		graph:      graph,
		identifier: identifier,
		nodeRules:  rules,
		features:   utilfeature.DefaultFeatureGate,
	}
}

var (
	configMapResource = api.Resource("configmaps")
	secretResource    = api.Resource("secrets")
	pvcResource       = api.Resource("persistentvolumeclaims")
	pvResource        = api.Resource("persistentvolumes")
	vaResource        = storageapi.Resource("volumeattachments")
	svcAcctResource   = api.Resource("serviceaccounts")
	leaseResource     = coordapi.Resource("leases")
	csiNodeResource   = storageapi.Resource("csinodes")
)

func (r *NodeAuthorizer) RulesFor(user user.Info, namespace string) ([]authorizer.ResourceRuleInfo, []authorizer.NonResourceRuleInfo, bool, error) {
	if _, isNode := r.identifier.NodeIdentity(user); isNode {
		// indicate nodes do not have fully enumerated permissions
		return nil, nil, true, fmt.Errorf("node authorizer does not support user rule resolution")
	}
	return nil, nil, false, nil
}

func (r *NodeAuthorizer) Authorize(ctx context.Context, attrs authorizer.Attributes) (authorizer.Decision, string, error) {
	panic("not implemented")
}

func (r *NodeAuthorizer) authorizeStatusUpdate(nodeName string, startingType vertexType, attrs authorizer.Attributes) (authorizer.Decision, string, error) {
	switch attrs.GetVerb() {
	case "update", "patch":
		// ok
	default:
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, "can only get/update/patch this type", nil
	}

	if attrs.GetSubresource() != "status" {
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, "can only update status subresource", nil
	}

	return r.authorize(nodeName, startingType, attrs)
}

// authorizeGet authorizes "get" requests to objects of the specified type if they are related to the specified node
func (r *NodeAuthorizer) authorizeGet(nodeName string, startingType vertexType, attrs authorizer.Attributes) (authorizer.Decision, string, error) {
	if attrs.GetVerb() != "get" {
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, "can only get individual resources of this type", nil
	}
	if len(attrs.GetSubresource()) > 0 {
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, "cannot get subresource", nil
	}
	return r.authorize(nodeName, startingType, attrs)
}

// authorizeReadNamespacedObject authorizes "get", "list" and "watch" requests to single objects of a
// specified types if they are related to the specified node.
func (r *NodeAuthorizer) authorizeReadNamespacedObject(nodeName string, startingType vertexType, attrs authorizer.Attributes) (authorizer.Decision, string, error) {
	switch attrs.GetVerb() {
	case "get", "list", "watch":
		//ok
	default:
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, "can only read resources of this type", nil
	}

	if len(attrs.GetSubresource()) > 0 {
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, "cannot read subresource", nil
	}
	if len(attrs.GetNamespace()) == 0 {
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, "can only read namespaced object of this type", nil
	}
	return r.authorize(nodeName, startingType, attrs)
}

func (r *NodeAuthorizer) authorize(nodeName string, startingType vertexType, attrs authorizer.Attributes) (authorizer.Decision, string, error) {
	if len(attrs.GetName()) == 0 {
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, "No Object name found", nil
	}

	ok, err := r.hasPathFrom(nodeName, startingType, attrs.GetNamespace(), attrs.GetName())
	if err != nil {
		klog.V(2).InfoS("NODE DENY", "err", err)
		return authorizer.DecisionNoOpinion, fmt.Sprintf("no relationship found between node '%s' and this object", nodeName), nil
	}
	if !ok {
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, fmt.Sprintf("no relationship found between node '%s' and this object", nodeName), nil
	}
	return authorizer.DecisionAllow, "", nil
}

// authorizeCreateToken authorizes "create" requests to serviceaccounts 'token'
// subresource of pods running on a node
func (r *NodeAuthorizer) authorizeCreateToken(nodeName string, startingType vertexType, attrs authorizer.Attributes) (authorizer.Decision, string, error) {
	if attrs.GetVerb() != "create" || len(attrs.GetName()) == 0 {
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, "can only create tokens for individual service accounts", nil
	}

	if attrs.GetSubresource() != "token" {
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, "can only create token subresource of serviceaccount", nil
	}

	ok, err := r.hasPathFrom(nodeName, startingType, attrs.GetNamespace(), attrs.GetName())
	if err != nil {
		klog.V(2).Infof("NODE DENY: %v", err)
		return authorizer.DecisionNoOpinion, fmt.Sprintf("no relationship found between node '%s' and this object", nodeName), nil
	}
	if !ok {
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, fmt.Sprintf("no relationship found between node '%s' and this object", nodeName), nil
	}
	return authorizer.DecisionAllow, "", nil
}

// authorizeLease authorizes node requests to coordination.k8s.io/leases.
func (r *NodeAuthorizer) authorizeLease(nodeName string, attrs authorizer.Attributes) (authorizer.Decision, string, error) {
	// allowed verbs: get, create, update, patch, delete
	verb := attrs.GetVerb()
	switch verb {
	case "get", "create", "update", "patch", "delete":
		//ok
	default:
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, "can only get, create, update, patch, or delete a node lease", nil
	}

	// the request must be against the system namespace reserved for node leases
	if attrs.GetNamespace() != api.NamespaceNodeLease {
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, fmt.Sprintf("can only access leases in the %q system namespace", api.NamespaceNodeLease), nil
	}

	// the request must come from a node with the same name as the lease
	// note we skip this check for create, since the authorizer doesn't know the name on create
	// the noderestriction admission plugin is capable of performing this check at create time
	if verb != "create" && attrs.GetName() != nodeName {
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, "can only access node lease with the same name as the requesting node", nil
	}

	return authorizer.DecisionAllow, "", nil
}

// authorizeCSINode authorizes node requests to CSINode storage.k8s.io/csinodes
func (r *NodeAuthorizer) authorizeCSINode(nodeName string, attrs authorizer.Attributes) (authorizer.Decision, string, error) {
	// allowed verbs: get, create, update, patch, delete
	verb := attrs.GetVerb()
	switch verb {
	case "get", "create", "update", "patch", "delete":
		//ok
	default:
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, "can only get, create, update, patch, or delete a CSINode", nil
	}

	if len(attrs.GetSubresource()) > 0 {
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, "cannot authorize CSINode subresources", nil
	}

	// the request must come from a node with the same name as the CSINode
	// note we skip this check for create, since the authorizer doesn't know the name on create
	// the noderestriction admission plugin is capable of performing this check at create time
	if verb != "create" && attrs.GetName() != nodeName {
		klog.V(2).Infof("NODE DENY: '%s' %#v", nodeName, attrs)
		return authorizer.DecisionNoOpinion, "can only access CSINode with the same name as the requesting node", nil
	}

	return authorizer.DecisionAllow, "", nil
}

// hasPathFrom returns true if there is a directed path from the specified type/namespace/name to the specified Node
func (r *NodeAuthorizer) hasPathFrom(nodeName string, startingType vertexType, startingNamespace, startingName string) (bool, error) {
	r.graph.lock.RLock()
	defer r.graph.lock.RUnlock()

	nodeVertex, exists := r.graph.getVertex_rlocked(nodeVertexType, "", nodeName)
	if !exists {
		return false, fmt.Errorf("unknown node '%s' cannot get %s %s/%s", nodeName, vertexTypes[startingType], startingNamespace, startingName)
	}

	startingVertex, exists := r.graph.getVertex_rlocked(startingType, startingNamespace, startingName)
	if !exists {
		return false, fmt.Errorf("node '%s' cannot get unknown %s %s/%s", nodeName, vertexTypes[startingType], startingNamespace, startingName)
	}

	// Fast check to see if we know of a destination edge
	if r.graph.destinationEdgeIndex[startingVertex.ID()].has(nodeVertex.ID()) {
		return true, nil
	}

	found := false
	traversal := &traverse.VisitingDepthFirst{
		EdgeFilter: func(edge graph.Edge) bool {
			if destinationEdge, ok := edge.(*destinationEdge); ok {
				if destinationEdge.DestinationID() != nodeVertex.ID() {
					// Don't follow edges leading to other nodes
					return false
				}
				// We found an edge leading to the node we want
				found = true
			}
			// Visit this edge
			return true
		},
	}
	traversal.Walk(r.graph.graph, startingVertex, func(n graph.Node) bool {
		if n.ID() == nodeVertex.ID() {
			// We found the node we want
			found = true
		}
		// Stop visiting if we've found the node we want
		return found
	})
	if !found {
		return false, fmt.Errorf("node '%s' cannot get %s %s/%s, no relationship to this object was found in the node authorizer graph", nodeName, vertexTypes[startingType], startingNamespace, startingName)
	}
	return true, nil
}