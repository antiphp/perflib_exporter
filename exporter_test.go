package main

import (
	"errors"
	"github.com/leoluk/perflib_exporter/perflib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

var (
	strict    = true
	notStrict = false
)

func Test_reduceObjects(t *testing.T) {
	type args struct {
		objects       []*perflib.PerfObject
		disallowedIDs *[]uint32
	}
	tests := []struct {
		name string
		args args
		want []*perflib.PerfObject
	}{
		{
			"objects=[] & disallowedIDs=[]; noop",
			args{
				[]*perflib.PerfObject{},
				&[]uint32{},
			},
			[]*perflib.PerfObject{},
		},
		{
			"objects=[123] & disallowedIDs=[]; noop",
			args{
				[]*perflib.PerfObject{{NameIndex: 123}},
				&[]uint32{},
			},
			[]*perflib.PerfObject{{NameIndex: 123}},
		},
		{
			"objects=[] & disallowedIDs=[123]; noop",
			args{
				[]*perflib.PerfObject{},
				&[]uint32{123},
			},
			[]*perflib.PerfObject{},
		},
		{
			"objects=[123] & disallowedIDs=[123]; reduce",
			args{
				[]*perflib.PerfObject{{NameIndex: 123}},
				&[]uint32{123},
			},
			[]*perflib.PerfObject{},
		},
		{
			"objects=[123] & disallowedIDs=[234]; noop",
			args{
				[]*perflib.PerfObject{{NameIndex: 123}},
				&[]uint32{234},
			},
			[]*perflib.PerfObject{{NameIndex: 123}},
		},
		{
			"objects=[123,234] & disallowedIDs=[234]; reduce",
			args{
				[]*perflib.PerfObject{{NameIndex: 123}, {NameIndex: 234}},
				&[]uint32{234},
			},
			[]*perflib.PerfObject{{NameIndex: 123}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reduceObjects(tt.args.objects, tt.args.disallowedIDs)
			require.Len(t, got, len(tt.want))
			assert.True(t, reflect.DeepEqual(got, tt.want))
		})
	}
}

func Test_removeObject(t *testing.T) {
	type args struct {
		object        *perflib.PerfObject
		disallowedIDs *[]uint32
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"disallowedIDs=nil",
			args{
				object: &perflib.PerfObject{
					NameIndex: 123,
				},
				disallowedIDs: nil,
			},
			false,
		},
		{
			"disallowedIDs=[]",
			args{
				object: &perflib.PerfObject{
					NameIndex: 123,
				},
				disallowedIDs: &[]uint32{},
			},
			false,
		},
		{
			"object.NameIndex=123 & disallowedIDs:[234]",
			args{
				object: &perflib.PerfObject{
					NameIndex: 123,
				},
				disallowedIDs: &[]uint32{234},
			},
			false,
		},
		{
			"object.NameIndex=123 & disallowedIDs:[123] => remove=true",
			args{
				object: &perflib.PerfObject{
					NameIndex: 123,
				},
				disallowedIDs: &[]uint32{123},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeObject(tt.args.object, tt.args.disallowedIDs); got != tt.want {
				t.Errorf("removeObject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newReductableQueryFunc_Error(t *testing.T) {
	badQueryFunc := func() ([]*perflib.PerfObject, error) {
		return nil, errors.New("test")
	}

	_, err := newReductableQueryFunc(badQueryFunc, &strict, nil)()
	assert.Error(t, err)

	_, err = newReductableQueryFunc(badQueryFunc, &notStrict, nil)()
	assert.Error(t, err)
}

func Test_newReductableQueryFunc(t *testing.T) {

	type args struct {
		queryObjects  []*perflib.PerfObject
		strict        *bool
		disallowedIDs *[]uint32
	}
	tests := []struct {
		name string
		args args
		want []*perflib.PerfObject
	}{
		{
			"strict; reduce",
			args{
				[]*perflib.PerfObject{{NameIndex: 123}},
				&strict,
				&[]uint32{123},
			},
			[]*perflib.PerfObject{},
		},
		{
			"strict; noop",
			args{
				[]*perflib.PerfObject{{NameIndex: 123}},
				&strict,
				&[]uint32{234},
			},
			[]*perflib.PerfObject{{NameIndex: 123}},
		},
		{
			"not strict; noop",
			args{
				[]*perflib.PerfObject{{NameIndex: 123}},
				&notStrict, // !
				&[]uint32{123},
			}, []*perflib.PerfObject{{NameIndex: 123}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := func() ([]*perflib.PerfObject, error) {
				return tt.args.queryObjects, nil
			}
			got, err := newReductableQueryFunc(query, tt.args.strict, tt.args.disallowedIDs)()
			require.NoError(t, err)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newReductableQueryFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}
