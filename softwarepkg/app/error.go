package app

import (
	"github.com/opensourceways/software-package-server/allerror"
	commonrepo "github.com/opensourceways/software-package-server/common/domain/repository"
)

const (
	errorSoftwarePkgExists          = "software_pkg_exists"
	errorSoftwarePkgNoPermission    = "software_pkg_no_permission"
	errorSoftwarePkgCannotComment   = "software_pkg_cannot_comment"
	errorSoftwarePkgCommentIllegal  = "software_pkg_comment_illegal"
	errorSoftwarePkgCommentNotFound = "software_pkg_comment_not_found"
)

func errorCodeForFindingPkg(err error) string {
	// TODO no need and delete
	return ""
}

func parseErrorForFindingPkg(err error) error {
	if commonrepo.IsErrorResourceNotFound(err) {
		return allerror.NewNotFound(allerror.ErrorCodePkgNotFound, err.Error())
	}

	return err
}
