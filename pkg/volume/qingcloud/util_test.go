package qingcloud

import (
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"
)

func TestFixVolumeCapacity(t *testing.T) {
	testCases := []struct {
		input      resource.Quantity
		expect     resource.Quantity
		volumeType VolumeType
		error      bool
	}{
		{
			input:      resource.MustParse("1Mi"),
			expect:     resource.MustParse("10Gi"),
			volumeType: VolumeTypeHP,
			error:      false,
		},
		{
			input:      resource.MustParse("10Gi"),
			expect:     resource.MustParse("10Gi"),
			volumeType: VolumeTypeHP,
			error:      false,
		},
		{
			input:      resource.MustParse("12Gi"),
			expect:     resource.MustParse("20Gi"),
			volumeType: VolumeTypeHP,
			error:      false,
		},
		{
			input:      resource.MustParse("50Gi"),
			expect:     resource.MustParse("50Gi"),
			volumeType: VolumeTypeHP,
			error:      false,
		},
		{
			input:      resource.MustParse("51Gi"),
			expect:     resource.MustParse("60Gi"),
			volumeType: VolumeTypeHP,
			error:      false,
		},

		{
			input:      resource.MustParse("1001Gi"),
			expect:     resource.MustParse("1001Gi"),
			volumeType: VolumeTypeHP,
			error:      true,
		},

		{
			input:      resource.MustParse("1Mi"),
			expect:     resource.MustParse("100Gi"),
			volumeType: VolumeTypeHC,
			error:      false,
		},
		{
			input:      resource.MustParse("100Gi"),
			expect:     resource.MustParse("100Gi"),
			volumeType: VolumeTypeHC,
			error:      false,
		},
		{
			input:      resource.MustParse("120Gi"),
			expect:     resource.MustParse("150Gi"),
			volumeType: VolumeTypeHC,
			error:      false,
		},
		{
			input:      resource.MustParse("500Gi"),
			expect:     resource.MustParse("500Gi"),
			volumeType: VolumeTypeHC,
			error:      false,
		},
		{
			input:      resource.MustParse("501Gi"),
			expect:     resource.MustParse("550Gi"),
			volumeType: VolumeTypeHC,
			error:      false,
		},

		{
			input:      resource.MustParse("5001Gi"),
			expect:     resource.MustParse("5001Gi"),
			volumeType: VolumeTypeHC,
			error:      true,
		},
	}
	for i, tc := range testCases {
		output, err := fixVolumeCapacity(tc.input, tc.volumeType)
		if tc.error {
			if err == nil {
				t.Fatalf("case %v: expect error but no error", i)
			}
		} else {
			if err != nil {
				t.Fatalf("case %v: unexpect error: %s", i, err)
			}
			if tc.expect.Value() != output.Value() {
				t.Errorf("case %v: expect %+v (%v) but get %+v (%v)", i, tc.expect, tc.expect.Value(), output, output.Value())
			}
		}
	}
}
