package goava

import (
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"github.com/itskovichanton/goava/pkg/goava/utils"
	"strings"
	"sync/atomic"
	"time"
)

type IGenerator interface {
	GenerateUint64() uint64
	GenerateUuid() uuid.UUID
	GenerateUuidFromString(arg string) (uuid.UUID, error)
	Reset()
}

type GeneratorImpl struct {
	IGenerator

	ops uint64
}

func (c *GeneratorImpl) GenerateUuidFromString(arg string) (uuid.UUID, error) {
	if len(arg) == 0 {
		return [16]byte{0}, errs.NewBaseError("empty argument for uuid")
	}
	arg = base64.StdEncoding.EncodeToString([]byte(arg))
	for {
		if len(arg) > 16 {
			break
		}
		arg += "0"
	}
	arg = arg[:16]
	return uuid.FromBytes([]byte(arg))
}

func (c *GeneratorImpl) Reset() {
	initV := utils.CurrentTimeMillis() / 1000
	c.ops = uint64(initV)
}

func (c *GeneratorImpl) GenerateUuid() uuid.UUID {
	r, err := uuid.NewRandomFromReader(strings.NewReader(fmt.Sprint(c.GenerateUint64())))
	if err != nil {
		return uuid.New()
	}
	return r
}

func (c *GeneratorImpl) GenerateUint64() uint64 {
	atomic.AddUint64(&c.ops, 1)
	return c.ops
}

func GetPseudoRnd() int64 {
	return time.Now().UnixNano() / (1 << 22)
}
