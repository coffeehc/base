package base

import (
	"log"

	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/utils"
)

var _ errors.Error
var _ log.Logger
var _, _ = utils.GetLocalIP()
