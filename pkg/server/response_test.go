package server

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapList_WrapsSliceInObject(t *testing.T) {
	items := []string{"a", "b", "c"}
	result := wrapList(items)

	assert.Equal(t, 3, result["count"])
	assert.Equal(t, items, result["items"])
}

func TestWrapList_EmptySlice(t *testing.T) {
	items := []int{}
	result := wrapList(items)

	assert.Equal(t, 0, result["count"])
	assert.Equal(t, items, result["items"])
}

func TestWrapList_NilSlice(t *testing.T) {
	var items []string
	result := wrapList(items)

	assert.Equal(t, 0, result["count"])
	assert.Nil(t, result["items"])
}

func TestWrapList_MarshalToJSONObject(t *testing.T) {
	// The key requirement: wrapList output must marshal to a JSON object, not an array
	items := []string{"x", "y"}
	result := wrapList(items)

	data, err := json.Marshal(result)
	require.NoError(t, err)

	// Must start with '{' (object), not '[' (array)
	assert.Equal(t, byte('{'), data[0], "structuredContent must marshal to a JSON object")

	// Verify it round-trips correctly
	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)
	assert.Equal(t, float64(2), parsed["count"])

	itemsList, ok := parsed["items"].([]interface{})
	require.True(t, ok, "items should be an array")
	assert.Len(t, itemsList, 2)
}

func TestWrapList_StructSlice(t *testing.T) {
	type mockCallback struct {
		ID   int    `json:"id"`
		Host string `json:"host"`
	}

	items := []mockCallback{
		{ID: 1, Host: "host1"},
		{ID: 2, Host: "host2"},
	}
	result := wrapList(items)

	assert.Equal(t, 2, result["count"])

	data, err := json.Marshal(result)
	require.NoError(t, err)

	// Verify it's a JSON object with items array
	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	itemsList, ok := parsed["items"].([]interface{})
	require.True(t, ok)
	assert.Len(t, itemsList, 2)

	first := itemsList[0].(map[string]interface{})
	assert.Equal(t, float64(1), first["id"])
	assert.Equal(t, "host1", first["host"])
}

func TestWrapList_SingleElement(t *testing.T) {
	items := []int{42}
	result := wrapList(items)

	assert.Equal(t, 1, result["count"])

	data, err := json.Marshal(result)
	require.NoError(t, err)

	// Still a JSON object even with single element
	assert.Equal(t, byte('{'), data[0])
}

func TestWrapList_LargeSlice(t *testing.T) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}
	result := wrapList(items)

	assert.Equal(t, 1000, result["count"])
	assert.Len(t, result, 2) // map has 2 keys: count + items
	wrapped := result["items"].([]int)
	assert.Len(t, wrapped, 1000)
}

func TestRawSlice_MarshalToArray(t *testing.T) {
	// Demonstrate the bug: raw slices marshal to JSON arrays
	items := []string{"a", "b"}
	data, err := json.Marshal(items)
	require.NoError(t, err)

	// Raw slice marshals to '[' (array) - this is what was breaking structuredContent
	assert.Equal(t, byte('['), data[0], "raw slice marshals to JSON array (the bug)")

	// wrapList fixes this
	wrapped := wrapList(items)
	data, err = json.Marshal(wrapped)
	require.NoError(t, err)
	assert.Equal(t, byte('{'), data[0], "wrapped list marshals to JSON object (the fix)")
}
