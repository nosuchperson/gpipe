package gpipe

import "testing"

func TestConfigMapUnmarshal(t *testing.T) {
	TEST_CASE := map[string]interface{}{
		"bool":    true,
		"int":     123,
		"float64": 3.1415926,
		"array":   []int{1, 2, 3, 4, 5, 6},
	}
	if result, err := ConfigMapUnmarshal(TEST_CASE, &struct {
		Bool    bool    `yaml:"bool"`
		Int     int     `yaml:"int"`
		Float64 float64 `yaml:"float64"`
		Array   []int   `yaml:"array"`
	}{}); err != nil {
		t.Fatal(err)
	} else {
		if result.Bool != true {
			t.Errorf("assert result.Bool == true, actual %v", result.Bool)
		}
		if result.Int != 123 {
			t.Errorf("assert result.Int == 123, actual %v", result.Int)
		}
		if result.Float64 != 3.1415926 {
			t.Errorf("assert result.Float64 == 3.1415926, actual %v", result.Float64)
		}
		if result.Array == nil {
			t.Errorf("assert result.Array = []int{}, actual %v", result.Array)
		}
		if len(result.Array) != 6 {
			t.Errorf("assert result.Array = [6]int{}, actual %v", result.Array)
		}
		for i, v := range result.Array {
			if v != i+1 {
				t.Errorf("assert result.Array[%d] = %d, actual %d", i, i+1, v)
			}
		}
	}
}

func TestConfigMapUnmarshal2(t *testing.T) {
	var TEST_CASE interface{} = []int{1, 2, 3, 4, 5, 6}
	RESULT := []int{}
	if result, err := ConfigMapUnmarshal(TEST_CASE, &RESULT); err != nil {
		t.Fatal(err)
	} else {
		if result == nil {
			t.Errorf("assert result = []int{}, actual %v", result)
		}
		if len(RESULT) != 6 {
			t.Errorf("assert result.Array = [6]int{}, actual %v", RESULT)
		}
		for i, v := range RESULT {
			if v != i+1 {
				t.Errorf("assert result.Array[%d] = %d, actual %d", i, i+1, v)
			}
		}
	}
}
