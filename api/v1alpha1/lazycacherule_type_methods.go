package v1alpha1

import (
	"google.golang.org/protobuf/encoding/protojson"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func (r *LazyCacheRule) NamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      r.Name,
		Namespace: r.Namespace,
	}
}

func (r *LazyCacheRule) Finalizer() string {
	return schema.GroupKind{Group: Group, Kind: KindLazyCacheRule}.String()
}

func (r *LazyCacheRule) CacheService() CacheService {
	return CacheService{
		Name:      r.Spec.CacheRef.Name,
		Namespace: r.Spec.CacheRef.Namespace,
	}
}

func (r *LazyCacheRule) ConfigMap() string {
	return r.CacheService().LazyCacheConfigMap()
}

func (r *LazyCacheRule) MarshallSpec() ([]byte, error) {
	return protojson.MarshalOptions{Multiline: true}.Marshal(&r.Spec)
}
