package clipboard

import "testing"

func TestCopyAndPaste(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"basic", "gopher"},
		{"Japanese", "ゴーファー"},
		{"emoji", "😀😁😂🤣😃😄😅"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Set(tt.want); err != nil {
				t.Error(err)
			}
			got, err := Get()
			if err != nil {
				t.Error(err)
			}
			if got != tt.want {
				t.Errorf("copy and paste mismatch: got %v, want %v", got, tt.want)
			}
		})
	}
}
