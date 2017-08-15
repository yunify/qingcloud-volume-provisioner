package qingcloud

import (
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"
)

func TestRoundUpVolumeCapacity(t *testing.T) {
	testCases := []struct {
		input      resource.Quantity
		expect     int
		volumeType VolumeType
		error      bool
	}{
		{
			input:      resource.MustParse("1Mi"),
			expect:     10,
			volumeType: VolumeTypeHP,
			error:      false,
		},
		{
			input:      resource.MustParse("3Gi"),
			expect:     10,
			volumeType: VolumeTypeHP,
			error:      false,
		},
		{
			input:      resource.MustParse("10Gi"),
			expect:     10,
			volumeType: VolumeTypeHP,
			error:      false,
		},
		{
			input:      resource.MustParse("12Gi"),
			expect:     20,
			volumeType: VolumeTypeHP,
			error:      false,
		},
		{
			input:      resource.MustParse("50Gi"),
			expect:     50,
			volumeType: VolumeTypeHP,
			error:      false,
		},
		{
			input:      resource.MustParse("51Gi"),
			expect:     60,
			volumeType: VolumeTypeHP,
			error:      false,
		},

		{
			input:      resource.MustParse("1001Gi"),
			expect:     1001,
			volumeType: VolumeTypeHP,
			error:      true,
		},

		{
			input:      resource.MustParse("1Mi"),
			expect:     100,
			volumeType: VolumeTypeHC,
			error:      false,
		},
		{
			input:      resource.MustParse("100Gi"),
			expect:     100,
			volumeType: VolumeTypeHC,
			error:      false,
		},
		{
			input:      resource.MustParse("120Gi"),
			expect:     150,
			volumeType: VolumeTypeHC,
			error:      false,
		},
		{
			input:      resource.MustParse("500Gi"),
			expect:     500,
			volumeType: VolumeTypeHC,
			error:      false,
		},
		{
			input:      resource.MustParse("501Gi"),
			expect:     550,
			volumeType: VolumeTypeHC,
			error:      false,
		},

		{
			input:      resource.MustParse("5001Gi"),
			expect:     5001,
			volumeType: VolumeTypeHC,
			error:      true,
		},
	}
	for i, tc := range testCases {
		output, err := RoundUpVolumeCapacity(tc.input, tc.volumeType)
		if tc.error {
			if err == nil {
				t.Fatalf("case %v: expect error but no error", i)
			}
		} else {
			if err != nil {
				t.Fatalf("case %v: unexpect error: %s", i, err)
			}
			if tc.expect != output {
				t.Errorf("case %v: expect %v but get %v", i, tc.expect, output)
			}
		}
	}
}

//func TestQuantity(t *testing.T){
//	q := resource.MustParse("1Gi")
//	println("q.Value()",q.Value())
//	//cv,b := q.AsScale(resource.Giga)
//	//cv.AsCanonicalBase1024Bytes()
//	//println("q.AsScale(resource.Giga)",, b)
//	println("q.Size()",q.Size())
//	i,b := q.AsInt64()
//	println("q.AsInt64()",i,b)
//	result := make([]byte, 0, 18)
//	b1,b2 := q.AsCanonicalBytes(result)
//	println("q.AsCanonicalBytes()",string(b1),string(b2))
//	b3,_ := q.MarshalJSON()
//	println("q.MarshalJSON()", string(b3))
//	println("q.ScaledValue(0)",q.ScaledValue(0))
//	println("q.ScaledValue(resource.Giga)",q.ScaledValue(resource.Giga))
//	println("q.ScaledValue(resource.Mega)",q.ScaledValue(resource.Mega))
//	q2 := resource.NewQuantity(q.Value(), resource.BinarySI)
//
//	println("q2.ScaledValue(0)",q2.ScaledValue(0))
//	println("q2.ScaledValue(resource.Giga)",q2.ScaledValue(resource.Giga))
//	println("q2.ScaledValue(resource.Mega)",q2.ScaledValue(resource.Mega))
//}

