package v1alpha1

import "fmt"

func (r *CacheRegion) Filename() string {
	return fmt.Sprintf("%s_%s", r.Namespace, r.Name)
}
