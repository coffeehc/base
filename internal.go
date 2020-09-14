package base

import (
	"log"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/utils"
)

var _ errors.Error
var _ log.Logger
var _, _ = utils.GetLocalIP()
