// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

// EagerCacheRuleSpecApplyConfiguration represents an declarative configuration of the EagerCacheRuleSpec type for use
// with apply.
type EagerCacheRuleSpecApplyConfiguration struct {
	CacheRef  *NamespacedObjectReferenceApplyConfiguration `json:"cache_ref,omitempty"`
	TableName *string                                      `json:"table_name,omitempty"`
	Key       *KeyApplyConfiguration                       `json:"key,omitempty"`
	Value     *ValueApplyConfiguration                     `json:"value,omitempty"`
}

// EagerCacheRuleSpecApplyConfiguration constructs an declarative configuration of the EagerCacheRuleSpec type for use with
// apply.
func EagerCacheRuleSpec() *EagerCacheRuleSpecApplyConfiguration {
	return &EagerCacheRuleSpecApplyConfiguration{}
}

// WithCacheRef sets the CacheRef field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the CacheRef field is set to the value of the last call.
func (b *EagerCacheRuleSpecApplyConfiguration) WithCacheRef(value *NamespacedObjectReferenceApplyConfiguration) *EagerCacheRuleSpecApplyConfiguration {
	b.CacheRef = value
	return b
}

// WithTableName sets the TableName field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TableName field is set to the value of the last call.
func (b *EagerCacheRuleSpecApplyConfiguration) WithTableName(value string) *EagerCacheRuleSpecApplyConfiguration {
	b.TableName = &value
	return b
}

// WithKey sets the Key field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Key field is set to the value of the last call.
func (b *EagerCacheRuleSpecApplyConfiguration) WithKey(value *KeyApplyConfiguration) *EagerCacheRuleSpecApplyConfiguration {
	b.Key = value
	return b
}

// WithValue sets the Value field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Value field is set to the value of the last call.
func (b *EagerCacheRuleSpecApplyConfiguration) WithValue(value *ValueApplyConfiguration) *EagerCacheRuleSpecApplyConfiguration {
	b.Value = value
	return b
}
