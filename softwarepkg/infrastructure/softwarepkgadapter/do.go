package softwarepkgadapter

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/opensourceways/software-package-server/softwarepkg/domain"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/dp"
	"github.com/opensourceways/software-package-server/softwarepkg/domain/repository"
)

const (
	fieldSig          = "sig"
	fieldIndex        = "_id"
	fieldPhase        = "phase"
	fieldVersion      = "version"
	fieldReviews      = "reviews"
	fieldCIStatus     = "ci.status"
	fieldImporter     = "importer"
	fieldAppliedAt    = "applied_at"
	fieldBasicDesc    = "basic.desc"
	fieldPrimaryKey   = "basic.name"
	fieldCommunityPR  = "community_pr"
	fieldInitialized  = "initialized"
	fieldRepoPlatform = "repo.platform"
)

// convert to data object

func toSoftwarePkgDO(pkg *domain.SoftwarePkg, do *softwarePkgDO) {
	*do = softwarePkgDO{
		Sig:   pkg.Sig.ImportingPkgSig(),
		Phase: pkg.Phase.PackagePhase(),
		Importer: committerDO{
			Account:    pkg.Importer.Account.Account(),
			PlatformId: pkg.Importer.PlatformId,
		},
		AppliedAt:   pkg.AppliedAt,
		Initialized: pkg.Initialized,
		CI:          toSoftwarePkgCIDO(&pkg.CI),
		Logs:        toOperationLogDOs(pkg.Logs),
		Spec:        toCodeInfoDO(&pkg.Code.Spec),
		SRPM:        toCodeInfoDO(&pkg.Code.SRPM),
		Repo:        toSoftwarePkgRepoDO(&pkg.Repo),
		Basic:       toSoftwarePkgBasicDO(&pkg.Basic),
		Reviews:     toReviewDOs(pkg.Reviews),
	}

	if pkg.CommunityPR != nil {
		do.CommunityPR = pkg.CommunityPR.URL()
	}
}

func toSoftwarePkgBasicDO(basic *domain.SoftwarePkgBasicInfo) softwarePkgBasicDO {
	return softwarePkgBasicDO{
		Name:     basic.Name.PackageName(),
		Desc:     basic.Desc.PackageDesc(),
		Purpose:  basic.Purpose.PurposeToImportPkg(),
		Upstream: basic.Upstream.URL(),
	}
}

func toSoftwarePkgRepoDO(repo *domain.SoftwarePkgRepo) softwarePkgRepoDO {
	cs := make([]committerDO, len(repo.Committers))
	for i, item := range repo.Committers {
		cs[i] = committerDO{
			Account:    item.Account.Account(),
			PlatformId: item.PlatformId,
		}
	}

	return softwarePkgRepoDO{
		Platform:   repo.Platform.PackagePlatform(),
		Committers: cs,
	}
}

func toSoftwarePkgCIDO(ci *domain.SoftwarePkgCI) softwarePkgCIDO {
	return softwarePkgCIDO{
		Id:        ci.Id,
		Status:    ci.Status().PackageCIStatus(),
		StartTime: ci.StartTime,
	}
}

func toCodeInfoDO(f *domain.SoftwarePkgCodeFile) codeInfoDO {
	do := codeInfoDO{
		Src:       f.Src.URL(),
		Dirty:     f.Dirty,
		UpdatedAt: f.UpdatedAt,
	}
	if f.DownloadAddr != nil {
		do.DownloadAddr = f.DownloadAddr.URL()
	}

	return do
}

func toOperationLogDOs(logs []domain.SoftwarePkgOperationLog) []softwarePkgOperationLogDO {
	if len(logs) == 0 {
		return nil
	}

	v := make([]softwarePkgOperationLogDO, len(logs))
	for i := range logs {
		item := &logs[i]

		v[i] = softwarePkgOperationLogDO{
			Time:   item.Time,
			User:   item.User.Account(),
			Action: item.Action.PackageOperationLogAction(),
		}
	}

	return v
}

func toReviewDOs(reviews []domain.UserReview) []userReviewDO {
	if len(reviews) == 0 {
		return nil
	}

	v := make([]userReviewDO, 0, len(reviews))
	for i := range reviews {
		item := &reviews[i]

		if len(item.Reviews) == 0 {
			continue
		}

		v = append(v, userReviewDO{
			Account: item.Account.Account(),
			GiteeID: item.GiteeID,
			Reviews: toCheckItemReviewDOs(item.Reviews),
		})
	}

	return v
}

func toCheckItemReviewDOs(reviews []domain.CheckItemReviewInfo) []checkItemReviewInfoDO {
	if len(reviews) == 0 {
		return nil
	}

	v := make([]checkItemReviewInfoDO, len(reviews))
	for i := range reviews {
		item := &reviews[i]

		v[i] = checkItemReviewInfoDO{
			Id:   item.Id,
			Pass: item.Pass,
		}
		if item.Comment != nil {
			v[i].Comment = item.Comment.ReviewComment()
		}
	}

	return v
}

// softwarePkgDO
type softwarePkgDO struct {
	Id          primitive.ObjectID `bson:"_id"           json:"-"`
	Sig         string             `bson:"sig"           json:"sig"           required:"true"`
	Phase       string             `bson:"phase"         json:"phase"         required:"true"`
	Importer    committerDO        `bson:"importer"      json:"importer"      required:"true"`
	AppliedAt   int64              `bson:"applied_at"    json:"applied_at"    required:"true"`
	CommunityPR string             `bson:"community_pr"  json:"community_pr"`
	Initialized bool               `bson:"initialized"   json:"initialized"`
	Version     int                `bson:"version"       json:"-"`

	CI      softwarePkgCIDO             `bson:"ci"      json:"ci"`
	Logs    []softwarePkgOperationLogDO `bson:"logs"    json:"logs"`
	Spec    codeInfoDO                  `bson:"spec"    json:"spec"`
	SRPM    codeInfoDO                  `bson:"srpm"    json:"srpm"`
	Repo    softwarePkgRepoDO           `bson:"repo"    json:"repo"`
	Basic   softwarePkgBasicDO          `bson:"basic"   json:"basic"`
	Reviews []userReviewDO              `bson:"reviews" json:"reviews"`
}

func (do *softwarePkgDO) toDomain(pkg *domain.SoftwarePkg) (err error) {
	pkg.Id = do.Id.Hex()

	if pkg.Sig, err = dp.NewImportingPkgSigInternal(do.Sig); err != nil {
		return
	}

	if pkg.Phase, err = dp.NewPackagePhase(do.Phase); err != nil {
		return
	}

	if err = do.Importer.toDomain(&pkg.Importer); err != nil {
		return
	}

	pkg.AppliedAt = do.AppliedAt

	if do.CommunityPR != "" {
		if pkg.CommunityPR, err = dp.NewURL(do.CommunityPR); err != nil {
			return
		}
	}

	pkg.Initialized = do.Initialized

	if pkg.CI, err = do.CI.toDomain(); err != nil {
		return
	}

	if pkg.Logs, err = do.domainLogs(); err != nil {
		return
	}

	if err = do.domainCode(&pkg.Code); err != nil {
		return
	}

	if err = do.Repo.toDomain(&pkg.Repo); err != nil {
		return
	}

	if err = do.Basic.toDomain(&pkg.Basic); err != nil {
		return
	}

	pkg.Reviews, err = do.domainReviews()

	return
}

func (do *softwarePkgDO) toSoftwarePkgInfo(info *repository.SoftwarePkgInfo) (err error) {
	info.Id = do.Id.Hex()

	if info.Sig, err = dp.NewImportingPkgSigInternal(do.Sig); err != nil {
		return
	}

	if info.Phase, err = dp.NewPackagePhase(do.Phase); err != nil {
		return
	}

	if info.PkgName, err = dp.NewPackageName(do.Basic.Name); err != nil {
		return
	}

	if info.PkgDesc, err = dp.NewPackageDesc(do.Basic.Desc); err != nil {
		return
	}

	if info.Platform, err = dp.NewPackagePlatform(do.Repo.Platform); err != nil {
		return
	}

	if info.CIStatus, err = dp.NewPackageCIStatus(do.CI.Status); err != nil {
		return
	}

	if info.Importer, err = dp.NewAccount(do.Importer.Account); err != nil {
		return
	}

	info.AppliedAt = do.AppliedAt

	return
}

func (do *softwarePkgDO) domainLogs() ([]domain.SoftwarePkgOperationLog, error) {
	if len(do.Logs) == 0 {
		return nil, nil
	}

	v := make([]domain.SoftwarePkgOperationLog, len(do.Logs))
	for i := range do.Logs {
		if err := do.Logs[i].toDomain(&v[i]); err != nil {
			return nil, err
		}
	}

	return v, nil
}

func (do *softwarePkgDO) domainCode(code *domain.SoftwarePkgCode) error {
	if err := do.Spec.toDomain(&code.Spec); err != nil {
		return err
	}

	return do.SRPM.toDomain(&code.SRPM)
}

func (do *softwarePkgDO) domainReviews() ([]domain.UserReview, error) {
	if len(do.Reviews) == 0 {
		return nil, nil
	}

	v := make([]domain.UserReview, len(do.Reviews))
	for i := range do.Reviews {
		if err := do.Reviews[i].toDomain(&v[i]); err != nil {
			return nil, err
		}
	}

	return v, nil
}

func (do *softwarePkgDO) toDoc() (bson.M, error) {
	return genDoc(do)
}

func (do *softwarePkgDO) docFilter() bson.M {
	return bson.M{
		fieldPrimaryKey: do.Basic.Name,
		fieldPhase:      bson.M{"$ne": dp.PackagePhaseClosed.PackagePhase()},
	}
}

// softwarePkgRepoDO
type softwarePkgRepoDO struct {
	Platform   string        `bson:"platform"     json:"platform"     required:"true"`
	Committers []committerDO `bson:"committers"   json:"committers"   required:"true"`
}

func (do *softwarePkgRepoDO) toDomain(repo *domain.SoftwarePkgRepo) (err error) {
	if repo.Platform, err = dp.NewPackagePlatform(do.Platform); err != nil {
		return
	}

	cs := make([]domain.PkgCommitter, len(do.Committers))

	for i := range do.Committers {
		if err = do.Committers[i].toDomain(&cs[i]); err != nil {
			return
		}
	}

	repo.Committers = cs

	return
}

// committerDO
type committerDO struct {
	Account    string `bson:"account"       json:"account"       required:"true"`
	PlatformId string `bson:"platform_Id"   json:"platform_Id"   required:"true"`
}

func (do *committerDO) toDomain(c *domain.PkgCommitter) (err error) {
	if c.Account, err = dp.NewAccount(do.Account); err != nil {
		return
	}

	c.PlatformId = do.PlatformId

	return
}

// codeInfoDO
type codeInfoDO struct {
	Src          string `bson:"src"            json:"src"         required:"true"`
	Dirty        bool   `bson:"dirty"          json:"dirty"`
	UpdatedAt    int64  `bson:"updated_at"     json:"updated_at"`
	DownloadAddr string `bson:"download_addr"  json:"download_addr"`
}

func (do *codeInfoDO) toDomain(f *domain.SoftwarePkgCodeFile) (err error) {
	if f.Src, err = dp.NewURL(do.Src); err != nil {
		return
	}

	if do.DownloadAddr != "" {
		if f.DownloadAddr, err = dp.NewURL(do.DownloadAddr); err != nil {
			return
		}
	}

	f.Dirty = do.Dirty
	f.UpdatedAt = do.UpdatedAt

	return
}

// softwarePkgBasicDO
type softwarePkgBasicDO struct {
	Name     string `bson:"name"     json:"name"     required:"true"`
	Desc     string `bson:"desc"     json:"desc"     required:"true"`
	Purpose  string `bson:"purpose"  json:"purpose"  required:"true"`
	Upstream string `bson:"upstream" json:"upstream" required:"true"`
}

func (do *softwarePkgBasicDO) toDomain(basic *domain.SoftwarePkgBasicInfo) (err error) {
	if basic.Name, err = dp.NewPackageName(do.Name); err != nil {
		return
	}

	if basic.Desc, err = dp.NewPackageDesc(do.Desc); err != nil {
		return
	}

	if basic.Purpose, err = dp.NewPurposeToImportPkg(do.Purpose); err != nil {
		return
	}

	basic.Upstream, err = dp.NewURL(do.Upstream)

	return
}

// softwarePkgCIDO
type softwarePkgCIDO struct {
	Id        int    `bson:"id"         json:"id"`
	Status    string `bson:"status"     json:"status"       required:"true"`
	StartTime int64  `bson:"start_time" json:"start_time"`
}

func (do *softwarePkgCIDO) toDomain() (domain.SoftwarePkgCI, error) {
	status, err := dp.NewPackageCIStatus(do.Status)
	if err != nil {
		return domain.SoftwarePkgCI{}, err
	}

	return domain.NewSoftwarePkgCI(do.Id, status, do.StartTime), nil
}

// softwarePkgOperationLogDO
type softwarePkgOperationLogDO struct {
	Time   int64  `bson:"time"     json:"time"    required:"true"`
	User   string `bson:"user"     json:"user"    required:"true"`
	Action string `bson:"action"   json:"action"  required:"true"`
}

func (do *softwarePkgOperationLogDO) toDomain(log *domain.SoftwarePkgOperationLog) (err error) {
	if log.User, err = dp.NewAccount(do.User); err != nil {
		return
	}

	log.Time = do.Time
	log.Action = dp.NewPackageOperationLogAction(do.Action)

	return
}

// userReviewDO
type userReviewDO struct {
	Account string                  `bson:"account"   json:"account"   required:"true"`
	GiteeID string                  `bson:"gitee_id"  json:"gitee_id"  required:"true"`
	Reviews []checkItemReviewInfoDO `bson:"reviews"   json:"reviews"   required:"true"`
}

func (do *userReviewDO) toDomain(review *domain.UserReview) (err error) {
	if review.Account, err = dp.NewAccount(do.Account); err != nil {
		return
	}

	review.GiteeID = do.GiteeID

	reviews := make([]domain.CheckItemReviewInfo, len(do.Reviews))
	for i := range do.Reviews {
		if err = do.Reviews[i].toDomain(&reviews[i]); err != nil {
			return
		}
	}

	review.Reviews = reviews

	return
}

// checkItemReviewInfoDO
type checkItemReviewInfoDO struct {
	Id      string `bson:"id"        json:"id"        required:"true"`
	Pass    bool   `bson:"pass"      json:"pass"`
	Comment string `bson:"comment"   json:"comment"`
}

func (do *checkItemReviewInfoDO) toDomain(info *domain.CheckItemReviewInfo) (err error) {
	info.Id = do.Id
	info.Pass = do.Pass
	if do.Comment != "" {
		info.Comment, err = dp.NewCheckItemComment(do.Comment)
	}

	return
}
