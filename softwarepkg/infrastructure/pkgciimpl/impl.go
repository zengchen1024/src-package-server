package pkgciimpl

import (
	"io/ioutil"

	"github.com/opensourceways/robot-gitee-lib/client"
	"github.com/opensourceways/server-common-lib/utils"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
)

var instance *pkgCIImpl

func Init(cfg *Config) error {
	if err := cloneRepo(cfg); err != nil {
		return err
	}

	instance = &pkgCIImpl{
		cli: client.NewClient(func() []byte {
			return []byte(cfg.GitUser.Token)
		}),
		cfg:         *cfg,
		repoDir:     cfg.WorkDir + "/" + cfg.CIRepo.Repo,
		pkgInfoFile: cfg.WorkDir + "/pkginfo.yaml",
	}

	return nil
}

func cloneRepo(cfg *Config) error {
	user := &cfg.GitUser

	params := []string{
		cfg.CloneScript,
		cfg.WorkDir,
		user.User,
		user.Email,
		cfg.CIRepo.Repo,
		cfg.CIRepo.cloneURL(user),
	}

	_, err, _ := utils.RunCmd(params...)

	return err
}

func PkgCI() *pkgCIImpl {
	return instance
}

type softwarePkgInfo struct {
	PkgId   string `json:"pkg_id"`
	PkgName string `json:"pkg_name"`
	Service string `json:"service"`
}

func (s *softwarePkgInfo) toYaml() ([]byte, error) {
	return yaml.Marshal(s)
}

// pkgCIImpl
type pkgCIImpl struct {
	cli         client.Client
	cfg         Config
	repoDir     string
	pkgInfoFile string
}

func (impl *pkgCIImpl) SendTest(info *domain.SoftwarePkgBasicInfo) error {
	if err := impl.createBranch(info); err != nil {
		return err
	}

	pr, err := impl.cli.CreatePullRequest(
		impl.cfg.CIRepo.Org,
		impl.cfg.CIRepo.Repo,
		info.PkgName.PackageName(),
		info.PkgName.PackageName(),
		impl.branch(info.PkgName),
		impl.cfg.TargetBranch,
		true,
	)
	if err != nil {
		return err
	}

	return impl.createPRComment(pr.Number)
}

func (impl *pkgCIImpl) createPRComment(id int32) error {
	err := impl.cli.CreatePRComment(
		impl.cfg.CIRepo.Org,
		impl.cfg.CIRepo.Repo, id,
		impl.cfg.CIComment,
	)
	if err != nil {
		logrus.Errorf("create pr %d comment failed, err:%s", id, err.Error())
	}

	return err
}

func (impl *pkgCIImpl) genPkgInfoFile(info *domain.SoftwarePkgBasicInfo) error {
	v := &softwarePkgInfo{
		PkgId:   info.Id,
		PkgName: info.PkgName.PackageName(),
		Service: impl.cfg.CIService,
	}

	content, err := v.toYaml()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(impl.pkgInfoFile, content, 0644)
}

func (impl *pkgCIImpl) createBranch(info *domain.SoftwarePkgBasicInfo) error {
	if err := impl.genPkgInfoFile(info); err != nil {
		return err
	}

	cfg := &impl.cfg
	code := &info.Application.SourceCode
	params := []string{
		cfg.PRScript,
		impl.repoDir,
		cfg.GitUser.Token,
		cfg.TargetBranch,
		impl.branch(info.PkgName),
		impl.pkgInfoFile,
		code.SpecURL.URL(),
		code.SrcRPMURL.URL(),
	}

	return impl.runcmd(params)
}

func (impl *pkgCIImpl) runcmd(params []string) error {
	out, err, _ := utils.RunCmd(params...)
	if err != nil {
		logrus.Errorf(
			"run create pull request shell, err=%s, out=%s, params=%v",
			err.Error(), out, params[:len(params)-1],
		)
	}

	return err
}

func (impl *pkgCIImpl) branch(pkg dp.PackageName) string {
	return pkg.PackageName()
}
