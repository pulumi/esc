package util

import "testing"

func TestJoinKey(t *testing.T) {
	type args struct {
		root string
		k    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty root",
			args: args{root: "", k: "foo"},
			want: "foo",
		},
		{
			name: "empty key",
			args: args{root: "foo", k: ""},
			want: "foo",
		},
		{
			name: "simple key",
			args: args{root: "foo", k: "bar"},
			want: "foo.bar",
		},
		{
			name: "key with dot",
			args: args{root: "foo", k: "bar.baz"},
			want: "foo[\"bar.baz\"]",
		},
		{
			name: "key with quote",
			args: args{root: "foo", k: "bar\"baz"},
			want: "foo[\"bar\\\"baz\"]",
		},
		{
			name: "key with quote and dot",
			args: args{root: "foo", k: "bar\"baz.qux"},
			want: "foo[\"bar\\\"baz.qux\"]",
		},
		{
			name: "key with quote and dot and backslash",
			args: args{root: "foo", k: "bar\"baz.qux\\quux"},
			want: "foo[\"bar\\\"baz.qux\\quux\"]",
		},
		{
			name: "key with digit",
			args: args{root: "foo", k: "bar1"},
			want: "foo.bar1",
		},
		{
			name: "key with digit at start",
			args: args{root: "foo", k: "0bar"},
			want: "foo[\"0bar\"]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JoinKey(tt.args.root, tt.args.k); got != tt.want {
				t.Errorf("JoinKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
