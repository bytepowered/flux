package dubbo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHasproto(t *testing.T) {
	tester := assert.New(t)
	tester.True(hasproto("lb://host:99"))
	tester.True(hasproto("zookeeper://host:99"))
	tester.True(hasproto("dubbo://host:99"))
	tester.True(hasproto("grpc://host:99"))
	tester.False(hasproto("host:99"))
	tester.False(hasproto("abc.com:99"))
	tester.False(hasproto("127.0.0.1"))
}
