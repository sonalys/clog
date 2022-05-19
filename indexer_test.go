package main

import (
	"reflect"
	"testing"
)

func Test_mapIndexer_Find(t *testing.T) {
	type fields struct {
		keys map[string]node
	}
	type args struct {
		fields []field
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp []line
	}{
		{
			name: "one key, one value",
			fields: fields{
				keys: map[string]node{
					"key": {
						values: leafs{
							{
								Value: []byte("value"),
								Lines: []line{
									{startOffset: 1, endOffset: 2},
								},
							},
						},
					},
				},
			},
			args: args{[]field{{key: []byte("key"), value: []byte("value")}}},
			wantResp: []line{
				{startOffset: 1, endOffset: 2},
			},
		},
		{
			name: "one key, two values",
			fields: fields{
				keys: map[string]node{
					"key": {
						values: leafs{
							{
								Value: []byte("foo"),
								Lines: []line{
									{startOffset: 3, endOffset: 4},
								},
							},
							{
								Value: []byte("value"),
								Lines: []line{
									{startOffset: 1, endOffset: 2},
								},
							},
						},
					},
				},
			},
			args: args{[]field{{key: []byte("key"), value: []byte("value")}}},
			wantResp: []line{
				{startOffset: 1, endOffset: 2},
			},
		},
		{
			name: "two keys, one value",
			fields: fields{
				keys: map[string]node{
					"key": {
						values: leafs{
							{
								Value: []byte("foo"),
								Lines: []line{
									{startOffset: 3, endOffset: 4},
								},
							},
							{
								Value: []byte("value"),
								Lines: []line{
									{startOffset: 1, endOffset: 2},
								},
							},
						},
					},
					"foo": {
						values: leafs{
							{
								Value: []byte("value"),
								Lines: []line{
									{startOffset: 5, endOffset: 6},
								},
							},
						},
					},
				},
			},
			args: args{[]field{{key: []byte("key"), value: []byte("value")}}},
			wantResp: []line{
				{startOffset: 1, endOffset: 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indexer := &mapIndexer{
				keys: tt.fields.keys,
			}
			if gotResp := indexer.Find(tt.args.fields); !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("mapIndexer.Find() = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}
