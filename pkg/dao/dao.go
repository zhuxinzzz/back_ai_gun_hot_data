package dao

func Init() {
	var err error

	pgDB, err = connectPostgresSQL()
	if err != nil {
		panic(err)
	}

	// 自动迁移表结构，确保GORM钩子正常工作
	//if err := pgDB.AutoMigrate(&dto.CmcToken{}, &dto.CmcTokenPrice{}); err != nil {
	//	panic(err)
	//}
	//print("Database tables migrated successfully")
}
