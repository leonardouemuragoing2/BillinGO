package models

import (
	"billingo/config"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres" // using postgres sql
	"gorm.io/gorm"
)

// SetupModels setup database
func SetupModels(conf *config.Config) *gorm.DB {
	// Load variables
	dbName := conf.DBName
	dbHost := conf.DBHost
	dbPort := conf.DBPort
	dbUser := conf.DBUser
	dbPwd := conf.DBPassword
	interval := conf.CompressionIntervalDays

	// Create postgresql url
	postgresConn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s", dbHost, dbPort, dbUser, dbName, dbPwd)
	fmt.Println("postgresConn :>>", postgresConn)

	// Open connection
	db, err := gorm.Open(postgres.Open(postgresConn), &gorm.Config{
		SkipDefaultTransaction: true,
		// Logger:                 logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Println(err)
		panic("Failed to connect to database:")
	}

	// Create tables, if not yet
	db.AutoMigrate(&Data{}, &DataRaw{})

	// Apply additional migrations
	AddSyncStatusMigration(db)

	setupHypertables(db)
	setupIndexes(db)
	if interval > 0 {
		setupCompression(db, interval)
	}
	return db
}

// setupHypertables creates Timescale Hypertable for data table
//   - this will create a warning on the log, but we can ignore it.
func setupHypertables(db *gorm.DB) {

	queryCreateHypertable := `
		SELECT create_hypertable('data', 'time', if_not_exists => TRUE);
	`
	db.Exec(queryCreateHypertable)
}

// setupIndexes creates system indexes
func setupIndexes(db *gorm.DB) {

	// The second index appears to have a better performance for queries, since after it was added, the queries speed
	// increased in ~10 times
	queryIndex := `
		CREATE INDEX IF NOT EXISTS data_index ON data(time DESC, vm_id);
		CREATE INDEX IF NOT EXISTS data_index_reverse ON data(vm_id, time DESC);
	`
	db.Exec(queryIndex)
}

func compressionConfigDiffers(db *gorm.DB, interval int, table string) bool {
	// SELECT if the current configuration matches the interval
	query := `
		SELECT config->>'compress_after' from timescaledb_information.jobs 
		WHERE proc_name='policy_compression' AND hypertable_name=? AND config->>'compress_after' = ?;
	`
	if ret := db.Exec(query, table, fmt.Sprintf("%d days", interval)); ret.Error != nil {
		if strings.Contains(ret.Error.Error(), "cannot change configuration on already compressed chunks") {
			// This installation was an old one and therefore we cannot compress the chunks on it.
			// TODO: Create a logic to enable this compression on already created tables
			log.Error("This installation currently doesn't allow for compression since it was first installed without it.")
		} else {
			log.Panic(ret.Error)
		}
	} else if ret.RowsAffected == 0 {
		return true // No Matches found, so the current configuration must be different
	}
	return false // We found a match, so it is not different.
}

// getTimescaleDBLicense queries the database for the TimescaleDB license type.
func getTimescaleDBLicense(db *gorm.DB) string {
	var license string
	query := "SHOW timescaledb.license;"

	// Execute the query and fetch the license value
	err := db.Raw(query).Scan(&license).Error
	if err != nil {
		fmt.Printf("Error querying TimescaleDB license: %v\n", err)
		return "unknown" // Return "unknown" if the query fails
	}

	return license
}

// setupCompression handles the creation of compression rules
func setupCompression(db *gorm.DB, interval int) {
	// Enable Compression of data
	license := getTimescaleDBLicense(db)
	if license == "apache" {
		return
	}

	queryCreateDataCompression := `
		ALTER TABLE data SET (
			timescaledb.compress,
			timescaledb.compress_orderby = 'time DESC',
			timescaledb.compress_segmentby = 'vm_id'
		);
	`
	execDB(db, queryCreateDataCompression)

	if compressionConfigDiffers(db, interval, "data") {
		log.Infof("New Compression Policy of %d days for data table. Removing old configuration", interval)
		// First we delete the old configuration and then we create it again
		deleteDataCompressionPolicy := `
			SELECT remove_compression_policy('data', if_exists => TRUE);
		`
		execDB(db, deleteDataCompressionPolicy)
	}

	querySetDataCompressionInterval := fmt.Sprintf(`
		SELECT add_compression_policy('data', INTERVAL '%d days', if_not_exists => TRUE);
	`, interval)
	execDB(db, querySetDataCompressionInterval)

}

// execDB will run a SQL query and panic if an error occurs
func execDB(db *gorm.DB, sql string) {
	if err := db.Exec(sql).Error; err != nil {
		if strings.Contains(err.Error(), "cannot change configuration on already compressed chunks") {
			// This installation was an old one and therefore we cannot compress the chunks on it.
			// TODO: Create a logic to enable this compression on already created tables
			log.Error("This installation currently doesn't allow for compression since it was first installed without it.")
		} else {
			log.Panic(err)
		}
	}
}

func InitTestDB() *gorm.DB {
	os.Setenv("ENVIRONMENT", "TESTING")
	gormDb := SetupModels(config.LoadConfig())
	db, err := gormDb.DB()
	if err != nil {
		panic(fmt.Errorf("init_db failed to get db instance: %w", err))
	}

	db.Exec("TRUNCATE public.servers, public.tags, public.tag_watchers, public.annotations, public.groups, public.emails, public.telegram_bots, public.telegram_chats CASCADE;")
	return gormDb
}
