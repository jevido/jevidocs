package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"dev.jevido/jevidocs/services/api/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000001CreateJobsTable{},
		&migrations.M20260924000001CreateDocsTables{},
		&migrations.M20260924000002CreateFeedbackAndRevisions{},
		&migrations.M20260924000003AddPagesRenderVersion{},
		&migrations.M20260924000004AddEditURLAndBanner{},
		&migrations.M20260924000005AddPagesRoot{},
		&migrations.M20260924000030AddProjectTheme{},
		&migrations.M20260924000040AddProjectSource{},
		&migrations.M20260924000010CreateAssetsTable{},
		&migrations.M20260924000020CreateAnalyticsTables{},
		&migrations.M20260924000050AddVersionsAndTrgm{},
		&migrations.M20260924000060AddLocales{},
		&migrations.M20260924000070AddUserRoles{},
		&migrations.M20260924000080CreateAccessTables{},
		&migrations.M20260924000090AddProjectDomains{},
	}
}
