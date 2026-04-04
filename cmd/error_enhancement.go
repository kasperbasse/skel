package cmd

import (
	"errors"
	"strings"

	appdoctor "github.com/kasperbasse/skel/internal/app/doctor"
	errorx "github.com/kasperbasse/skel/internal/app/errorx"
)

func applyErrorEnhancementRules(errMsg string) error {
	msg, ok := errorx.EnhanceMessage(errMsg, errorx.EnhanceOptions{
		ValidSections:  cyan(strings.Join(allRestoreKeys(), ", ")),
		ToolHint:       appdoctor.ToolNotFoundHint,
		SuggestProfile: errorx.SuggestSimilarProfileName,
	})
	if ok {
		return errors.New(msg)
	}
	return nil
}
