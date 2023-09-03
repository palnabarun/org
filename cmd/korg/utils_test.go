package main

import "testing"

func Test_search(t *testing.T) {
	type args struct {
		item       string
		collection []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "user is found",
			args: args{
				item:       "foo",
				collection: []string{"bar", "baz", "foo", "qux"},
			},
			want: true,
		},
		{
			name: "user is not found",
			args: args{
				item:       "john",
				collection: []string{"bar", "baz", "foo", "qux"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringInSlice(tt.args.collection, tt.args.item); got != tt.want {
				t.Errorf("search() = %v, want %v", got, tt.want)
			}
		})
	}
}
