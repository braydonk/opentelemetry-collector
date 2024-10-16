// Code generated by mdatagen. DO NOT EDIT.

package custom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceBuilder(t *testing.T) {
	for _, tt := range []string{"default", "all_set", "none_set"} {
		t.Run(tt, func(t *testing.T) {
			cfg := loadResourceAttributesConfig(t, tt)
			rb := NewResourceBuilder(cfg)
			rb.SetMapResourceAttr(map[string]any{"key1": "map.resource.attr-val1", "key2": "map.resource.attr-val2"})
			rb.SetOptionalResourceAttr("optional.resource.attr-val")
			rb.SetSliceResourceAttr([]any{"slice.resource.attr-item1", "slice.resource.attr-item2"})
			rb.SetStringEnumResourceAttrOne()
			rb.SetStringResourceAttr("string.resource.attr-val")
			rb.SetStringResourceAttrDisableWarning("string.resource.attr_disable_warning-val")
			rb.SetStringResourceAttrRemoveWarning("string.resource.attr_remove_warning-val")
			rb.SetStringResourceAttrToBeRemoved("string.resource.attr_to_be_removed-val")

			res := rb.Emit()
			assert.Equal(t, 0, rb.Emit().Attributes().Len()) // Second call should return empty Resource

			switch tt {
			case "default":
				assert.Equal(t, 6, res.Attributes().Len())
			case "all_set":
				assert.Equal(t, 8, res.Attributes().Len())
			case "none_set":
				assert.Equal(t, 0, res.Attributes().Len())
				return
			default:
				assert.Failf(t, "unexpected test case: %s", tt)
			}

			val, ok := res.Attributes().Get("map.resource.attr")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, map[string]any{"key1": "map.resource.attr-val1", "key2": "map.resource.attr-val2"}, val.Map().AsRaw())
			}
			val, ok = res.Attributes().Get("optional.resource.attr")
			assert.Equal(t, tt == "all_set", ok)
			if ok {
				assert.EqualValues(t, "optional.resource.attr-val", val.Str())
			}
			val, ok = res.Attributes().Get("slice.resource.attr")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, []any{"slice.resource.attr-item1", "slice.resource.attr-item2"}, val.Slice().AsRaw())
			}
			val, ok = res.Attributes().Get("string.enum.resource.attr")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "one", val.Str())
			}
			val, ok = res.Attributes().Get("string.resource.attr")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "string.resource.attr-val", val.Str())
			}
			val, ok = res.Attributes().Get("string.resource.attr_disable_warning")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "string.resource.attr_disable_warning-val", val.Str())
			}
			val, ok = res.Attributes().Get("string.resource.attr_remove_warning")
			assert.Equal(t, tt == "all_set", ok)
			if ok {
				assert.EqualValues(t, "string.resource.attr_remove_warning-val", val.Str())
			}
			val, ok = res.Attributes().Get("string.resource.attr_to_be_removed")
			assert.True(t, ok)
			if ok {
				assert.EqualValues(t, "string.resource.attr_to_be_removed-val", val.Str())
			}
		})
	}
}
