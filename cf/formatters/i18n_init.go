package formatters

import (
	"github.com/cloudfoundry/cli/cf/i18n"
	goi18n "github.com/nicksnyder/go-i18n/i18n"
)

var T goi18n.TranslateFunc

func init() {
	if T == nil {
		T = i18n.Init("cf/formatters", i18n.GetResourcesPath())
	}
}
