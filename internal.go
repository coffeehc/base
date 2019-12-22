package base

import (
	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/utils"
	"log"
)

var _ errors.Error
var _ log.Logger
var _, _ = utils.GetLocalIP()
