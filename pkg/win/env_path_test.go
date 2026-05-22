package win

import "testing"

func TestPrependPathValue(t *testing.T) {
	tests := []struct {
		name string
		path string
		entry string
		want string
	}{
		{
			name:  "empty path",
			path:  "",
			entry: `C:\Users\dev\.pvm\current`,
			want:  `C:\Users\dev\.pvm\current`,
		},
		{
			name:  "prepend to existing",
			path:  `C:\php;C:\tools`,
			entry: `C:\Users\dev\.pvm\current`,
			want:  `C:\Users\dev\.pvm\current;C:\php;C:\tools`,
		},
		{
			name:  "move duplicate from middle to front",
			path:  `C:\php;C:\Users\dev\.pvm\current;C:\tools`,
			entry: `C:\Users\dev\.pvm\current`,
			want:  `C:\Users\dev\.pvm\current;C:\php;C:\tools`,
		},
		{
			name:  "move duplicate from end to front",
			path:  `C:\php;C:\tools;C:\users\dev\.pvm\current`,
			entry: `C:\Users\dev\.pvm\current`,
			want:  `C:\Users\dev\.pvm\current;C:\php;C:\tools`,
		},
		{
			name:  "already at front unchanged order of rest",
			path:  `C:\Users\dev\.pvm\current;C:\php`,
			entry: `C:\Users\dev\.pvm\current`,
			want:  `C:\Users\dev\.pvm\current;C:\php`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PrependPathValue(tt.path, tt.entry)
			if got != tt.want {
				t.Fatalf("PrependPathValue(%q, %q) = %q, want %q", tt.path, tt.entry, got, tt.want)
			}
		})
	}
}
