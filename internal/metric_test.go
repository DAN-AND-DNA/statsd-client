package internal

import (
	tasset "github.com/stretchr/testify/assert"
	"testing"
)

func TestMetric_increment(t *testing.T) {
	err := GetSingleInst().reCreate("AAA.BBB.CCC.", []string{"127.0.0.1:10001", "127.0.0.1:10002", "127.0.0.1:10003"}, true)
	tasset.Nil(t, err)

	ret, err := GetSingleInst().increment("test_num_1", 1, 1, true)
	tasset.Nil(t, err)
	tasset.Equal(t, ret, "AAA.BBB.CCC.test_num_1:1|c")

	err = GetSingleInst().reCreate("AAA.BBB.CCC.", []string{"127.0.0.1:10001", "127.0.0.1:10002", "127.0.0.1:10003", "127.0.0.1:10004"}, true)
	tasset.Nil(t, err)

	ret, err = GetSingleInst().increment("test_num_2", 1, 1, true)
	tasset.Nil(t, err)
	tasset.Equal(t, ret, "AAA.BBB.CCC.test_num_2:1|c")
}
